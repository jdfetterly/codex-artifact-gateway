package gateway

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Root struct {
	Name string
	Path string
}

type ResolvedFile struct {
	RootName     string
	AbsolutePath string
	RelativePath string
	ViewPath     string
}

type Policy struct {
	roots  []Root
	byName map[string]Root
}

func NewPolicy(rootPaths []string) (*Policy, error) {
	roots := make([]Root, 0, len(rootPaths))
	byName := make(map[string]Root, len(rootPaths))
	seen := make(map[string]bool, len(rootPaths))
	for _, rootPath := range rootPaths {
		if strings.TrimSpace(rootPath) == "" {
			return nil, fmt.Errorf("root path cannot be empty")
		}
		abs, err := filepath.Abs(rootPath)
		if err != nil {
			return nil, fmt.Errorf("resolve root %q: %w", rootPath, err)
		}
		abs = filepath.Clean(abs)
		info, err := os.Lstat(abs)
		if err != nil {
			return nil, fmt.Errorf("inspect root %q: %w", rootPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("root %q is a symlink", rootPath)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("root %q is not a directory", rootPath)
		}
		name := filepath.Base(abs)
		if seen[name] {
			return nil, fmt.Errorf("duplicate root name %q", name)
		}
		seen[name] = true
		root := Root{Name: name, Path: abs}
		roots = append(roots, root)
		byName[name] = root
	}
	sort.Slice(roots, func(i, j int) bool {
		return len(roots[i].Path) > len(roots[j].Path)
	})
	return &Policy{roots: roots, byName: byName}, nil
}

func (p *Policy) Roots() []Root {
	out := make([]Root, len(p.roots))
	copy(out, p.roots)
	return out
}

func (p *Policy) ResolveInput(input string) (ResolvedFile, error) {
	localPath, err := normalizeInputPath(input)
	if err != nil {
		return ResolvedFile{}, err
	}
	abs, err := filepath.Abs(localPath)
	if err != nil {
		return ResolvedFile{}, err
	}
	return p.resolveAbsolute(filepath.Clean(abs), nil)
}

func (p *Policy) ResolveViewPath(rootName string, relativeURLPath string) (ResolvedFile, error) {
	root, ok := p.byName[rootName]
	if !ok {
		return ResolvedFile{}, fmt.Errorf("unknown root %q", rootName)
	}
	decoded, err := url.PathUnescape(relativeURLPath)
	if err != nil {
		return ResolvedFile{}, fmt.Errorf("decode relative path: %w", err)
	}
	cleanSlash := path.Clean("/" + decoded)
	relSlash := strings.TrimPrefix(cleanSlash, "/")
	if relSlash == "" || relSlash == "." {
		return ResolvedFile{}, fmt.Errorf("relative path cannot be empty")
	}
	relOS := filepath.FromSlash(relSlash)
	if filepath.IsAbs(relOS) || relOS == ".." || strings.HasPrefix(relOS, ".."+string(os.PathSeparator)) {
		return ResolvedFile{}, fmt.Errorf("relative path escapes root")
	}
	return p.resolveAbsolute(filepath.Join(root.Path, relOS), &root)
}

func (p *Policy) resolveAbsolute(abs string, forcedRoot *Root) (ResolvedFile, error) {
	if !isAllowedExtension(abs) {
		return ResolvedFile{}, fmt.Errorf("unsupported file type %q", filepath.Ext(abs))
	}
	candidates := p.roots
	if forcedRoot != nil {
		candidates = []Root{*forcedRoot}
	}
	for _, root := range candidates {
		rel, err := filepath.Rel(root.Path, abs)
		if err != nil {
			continue
		}
		if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}
		if err := rejectSymlinkComponents(root.Path, abs); err != nil {
			return ResolvedFile{}, err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return ResolvedFile{}, fmt.Errorf("inspect file %q: %w", abs, err)
		}
		if info.IsDir() {
			return ResolvedFile{}, fmt.Errorf("path %q is a directory", abs)
		}
		relSlash := filepath.ToSlash(rel)
		return ResolvedFile{
			RootName:     root.Name,
			AbsolutePath: abs,
			RelativePath: relSlash,
			ViewPath:     "/view/" + url.PathEscape(root.Name) + "/" + escapePath(relSlash),
		}, nil
	}
	return ResolvedFile{}, fmt.Errorf("path %q is outside configured roots", abs)
}

func normalizeInputPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	if strings.HasPrefix(trimmed, "file://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("parse file URL: %w", err)
		}
		if parsed.Scheme != "file" {
			return "", fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
		}
		if parsed.Host != "" && parsed.Host != "localhost" {
			return "", fmt.Errorf("unsupported file URL host %q", parsed.Host)
		}
		return parsed.Path, nil
	}
	return trimmed, nil
}

func rejectSymlinkComponents(rootAbs string, abs string) error {
	rel, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return err
	}
	current := rootAbs
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect path component %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink component %q is not allowed", current)
		}
	}
	return nil
}

func isAllowedExtension(filePath string) bool {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".html", ".htm", ".css", ".js", ".json", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".woff", ".woff2":
		return true
	default:
		return false
	}
}

func IsHTML(filePath string) bool {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func escapePath(relSlash string) string {
	parts := strings.Split(relSlash, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
