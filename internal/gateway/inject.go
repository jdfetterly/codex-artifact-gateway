package gateway

import (
	"bytes"
	"html"
)

func InjectFeedbackDrawer(content []byte, artifactPath string) []byte {
	drawer := []byte(feedbackDrawer(artifactPath))
	lower := bytes.ToLower(content)
	idx := bytes.LastIndex(lower, []byte("</body>"))
	if idx == -1 {
		out := make([]byte, 0, len(content)+len(drawer))
		out = append(out, content...)
		out = append(out, drawer...)
		return out
	}
	out := make([]byte, 0, len(content)+len(drawer))
	out = append(out, content[:idx]...)
	out = append(out, drawer...)
	out = append(out, content[idx:]...)
	return out
}

func feedbackDrawer(artifactPath string) string {
	escaped := html.EscapeString(artifactPath)
	return `
<div id="codex-gateway-feedback" data-artifact-path="` + escaped + `" style="position:fixed;right:12px;bottom:12px;z-index:2147483647;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
  <details style="width:min(360px,calc(100vw - 24px));background:#101114;color:#fff;border:1px solid rgba(255,255,255,.18);border-radius:10px;box-shadow:0 12px 34px rgba(0,0,0,.35);overflow:hidden;">
    <summary style="list-style:none;cursor:pointer;padding:12px 14px;font-size:15px;font-weight:650;">Feedback</summary>
    <form id="codex-gateway-feedback-form" style="display:grid;gap:10px;padding:0 14px 14px;">
      <select name="kind" aria-label="Feedback type" style="font:inherit;min-height:40px;border-radius:8px;border:1px solid #3d414a;background:#181a20;color:#fff;padding:8px;">
        <option value="needs_changes">Needs changes</option>
        <option value="looks_good">Looks good</option>
        <option value="question">Question</option>
      </select>
      <textarea name="comment" aria-label="Comment" placeholder="Leave feedback..." required style="font:inherit;min-height:96px;border-radius:8px;border:1px solid #3d414a;background:#181a20;color:#fff;padding:10px;resize:vertical;"></textarea>
      <button type="submit" style="font:inherit;min-height:42px;border:0;border-radius:8px;background:#f4f5f7;color:#111;font-weight:700;">Send feedback</button>
      <p data-status style="min-height:18px;margin:0;color:#b8bdc7;font-size:13px;"></p>
    </form>
  </details>
</div>
<script>
(() => {
  const host = document.getElementById('codex-gateway-feedback');
  const form = document.getElementById('codex-gateway-feedback-form');
  if (!host || !form) return;
  const status = form.querySelector('[data-status]');
  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const data = new FormData(form);
    status.textContent = 'Sending...';
    const payload = {
      artifact_path: host.dataset.artifactPath || location.pathname,
      kind: String(data.get('kind') || ''),
      comment: String(data.get('comment') || ''),
      href: location.href,
      title: document.title || ''
    };
    try {
      const response = await fetch('/api/feedback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!response.ok) throw new Error('Request failed');
      status.textContent = 'Feedback saved on the Mac.';
      form.reset();
    } catch (error) {
      status.textContent = 'Could not save feedback.';
    }
  });
})();
</script>`
}
