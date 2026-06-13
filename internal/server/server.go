package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jdfetterly/codex-artifact-gateway/internal/gateway"
)

const (
	maxResolveBodyBytes  = 64 * 1024
	maxFeedbackBodyBytes = 128 * 1024
)

type Config struct {
	Policy      *gateway.Policy
	FeedbackDir string
}

func NewHandler(config Config) http.Handler {
	mux := http.NewServeMux()
	app := &app{config: config, store: gateway.FeedbackStore{Dir: config.FeedbackDir}, startedAt: time.Now()}
	mux.HandleFunc("/", app.handleHome)
	mux.HandleFunc("/health", app.handleHealth)
	mux.HandleFunc("/recent", app.handleRecent)
	mux.HandleFunc("/resolve", app.handleResolve)
	mux.HandleFunc("/open", app.handleOpen)
	mux.HandleFunc("/view/", app.handleView)
	mux.HandleFunc("/api/feedback", app.handleFeedback)
	return mux
}

type app struct {
	config    Config
	store     gateway.FeedbackStore
	startedAt time.Time
}

func (a *app) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/recent", http.StatusFound)
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":             true,
		"root_count":     len(a.config.Policy.Roots()),
		"uptime_seconds": int(time.Since(a.startedAt).Seconds()),
	})
}

func (a *app) handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	resolved, err := a.config.Policy.ResolveInput(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, resolved.ViewPath, http.StatusFound)
}

func (a *app) handleResolve(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeHTML(w, resolvePage(""))
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, maxResolveBodyBytes)
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resolved, err := a.config.Policy.ResolveInput(r.FormValue("path"))
		if err != nil {
			writeHTML(w, resolvePage(err.Error()))
			return
		}
		http.Redirect(w, r, resolved.ViewPath, http.StatusFound)
	default:
		methodNotAllowed(w)
	}
}

func (a *app) handleView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/view/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	resolved, err := a.config.Policy.ResolveViewPath(parts[0], parts[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	content, err := os.ReadFile(resolved.AbsolutePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(resolved.AbsolutePath)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	if gateway.IsHTML(resolved.AbsolutePath) {
		content = gateway.InjectFeedbackDrawer(content, resolved.ViewPath)
	}
	_, _ = w.Write(content)
}

func (a *app) handleFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFeedbackBodyBytes)
	defer r.Body.Close()
	var entry gateway.FeedbackEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(entry.ArtifactPath) == "" || strings.TrimSpace(entry.Comment) == "" {
		http.Error(w, "artifact_path and comment are required", http.StatusBadRequest)
		return
	}
	entry.UserAgent = r.UserAgent()
	if _, err := a.store.Append(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (a *app) handleRecent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	items, err := recentItems(a.config.Policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var builder strings.Builder
	builder.WriteString(pageStart("Recent Artifacts"))
	builder.WriteString(`<main><header><h1>Recent Artifacts</h1><a class="secondary" href="/resolve">Resolve path</a></header><section class="list">`)
	if len(items) == 0 {
		builder.WriteString(`<p class="empty">No HTML files found under configured roots.</p>`)
	}
	for _, item := range items {
		builder.WriteString(`<a class="item" href="`)
		builder.WriteString(template.HTMLEscapeString(item.ViewPath))
		builder.WriteString(`"><strong>`)
		builder.WriteString(softBreakHTML(item.Name))
		builder.WriteString(`</strong><span>`)
		builder.WriteString(template.HTMLEscapeString(item.DisplayPath))
		builder.WriteString(`</span></a>`)
	}
	builder.WriteString(`</section></main>`)
	builder.WriteString(pageEnd())
	writeHTML(w, builder.String())
}

type recentItem struct {
	Name        string
	DisplayPath string
	ViewPath    string
	ModTime     time.Time
}

func recentItems(policy *gateway.Policy) ([]recentItem, error) {
	var items []recentItem
	for _, root := range policy.Roots() {
		err := filepath.WalkDir(root.Path, func(filePath string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if entry.IsDir() {
				if filePath == root.Path {
					return nil
				}
				rel, err := filepath.Rel(root.Path, filePath)
				if err == nil {
					if private, _ := gateway.IsPrivateRelativePath(rel); private || entry.Name() == "node_modules" {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if !gateway.IsHTML(filePath) {
				return nil
			}
			resolved, err := policy.ResolveInput(filePath)
			if err != nil {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return nil
			}
			items = append(items, recentItem{
				Name:        filepath.Base(filePath),
				DisplayPath: root.Name + "/" + resolved.RelativePath,
				ViewPath:    resolved.ViewPath,
				ModTime:     info.ModTime(),
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ModTime.After(items[j].ModTime)
	})
	if len(items) > 200 {
		items = items[:200]
	}
	return items, nil
}

func resolvePage(message string) string {
	var builder strings.Builder
	builder.WriteString(pageStart("Resolve Artifact"))
	builder.WriteString(`<main><header><h1>Resolve Artifact</h1><a class="secondary" href="/recent">Recent</a></header>`)
	builder.WriteString(`<form method="post" action="/resolve" class="resolver"><label for="path">Local path or file URL</label><textarea id="path" name="path" required placeholder="file:///Users/.../report.html"></textarea><button type="submit">Open artifact</button></form>`)
	if message != "" {
		builder.WriteString(`<p class="error">`)
		builder.WriteString(template.HTMLEscapeString(message))
		builder.WriteString(`</p>`)
	}
	builder.WriteString(`</main>`)
	builder.WriteString(pageEnd())
	return builder.String()
}

func pageStart(title string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` + template.HTMLEscapeString(title) + `</title><style>
body{margin:0;background:#f7f8fb;color:#14161a;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;overflow-x:hidden}
main{width:min(920px,100%);box-sizing:border-box;margin:0 auto;padding:18px;display:grid;gap:16px}
header{display:grid;grid-template-columns: 1fr;gap:10px;align-items:start}
h1{font-size:28px;line-height:1.1;margin:0;overflow-wrap: anywhere;word-break: break-word}
a{color:inherit}
.secondary{font-size:15px;color:#335c96}
.list{display:grid;grid-template-columns: 1fr;gap:10px}
.item{display:grid;gap:5px;padding:14px;border:1px solid #dde1ea;border-radius:8px;background:#fff;text-decoration:none;overflow-wrap: anywhere;word-break: break-word}
.item strong{font-size:16px}
.item span{font-size:13px;color:#5c6472}
.resolver{display:grid;gap:10px}
textarea{font:inherit;min-height:120px;border:1px solid #c8ced8;border-radius:8px;padding:10px;resize:vertical}
button{font:inherit;min-height:44px;border:0;border-radius:8px;background:#161a22;color:#fff;font-weight:700}
.error{color:#9c1c1c;overflow-wrap: anywhere;word-break: break-word}
.empty{color:#5c6472}
</style></head><body>`
}

func pageEnd() string {
	return `</body></html>`
}

func writeHTML(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(content))
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func softBreakHTML(s string) string {
	return strings.ReplaceAll(template.HTMLEscapeString(s), "-", "-<wbr>")
}

func Serve(config Config, addr string) error {
	if err := ValidateListenAddr(addr); err != nil {
		return err
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           NewHandler(config),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return server.ListenAndServe()
}

func StartupMessage(addr string, roots []string, feedbackDir string) string {
	return fmt.Sprintf("Codex Artifact Gateway listening on http://%s\nRecent: http://%s/recent\nResolve: http://%s/resolve\nFeedback: %s\nTailscale: tailscale serve --bg http://%s\nRoots: %s\n", addr, addr, addr, feedbackDir, addr, strings.Join(roots, ", "))
}

func ValidateListenAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("listen address must include a loopback host and port, such as 127.0.0.1:8767: %w", err)
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("listen address must be loopback-only; got %q", addr)
}
