package helps

import (
	"net/url"
	"strings"
	"sync"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

// claudeCompatHosts tracks the host:port of every claude-api-key upstream whose
// base-url points at a third-party Anthropic-compatible relay (not
// api.anthropic.com). Those relays are typically Cloudflare-fronted and filter
// on TLS fingerprint (JA3) plus User-Agent; sending Go's native TLS gets an
// immediate 403, which then trips CPA's cooldown. Routing them through the
// Claude uTLS fingerprint (like api.anthropic.com) lets them pass.
var claudeCompatHosts sync.Map // map[string]struct{}

func claudeHostKey(u *url.URL) string {
	if u == nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return ""
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	return host + ":" + port
}

// RegisterClaudeCompatUpstreams records the base-url hosts of all configured
// claude-api-key credentials so their requests use the Claude uTLS fingerprint.
// Called whenever config (re)loads; rebuilds the set from scratch.
func RegisterClaudeCompatUpstreams(cfg *config.Config) {
	claudeCompatHosts.Range(func(k, _ any) bool { claudeCompatHosts.Delete(k); return true })
	if cfg == nil {
		return
	}
	for _, ck := range cfg.ClaudeKey {
		base := strings.TrimSpace(ck.BaseURL)
		if base == "" {
			continue
		}
		u, err := url.Parse(base)
		if err != nil || u.Hostname() == "" {
			continue
		}
		if strings.EqualFold(u.Hostname(), "api.anthropic.com") {
			continue
		}
		claudeCompatHosts.Store(claudeHostKey(u), struct{}{})
	}
}

// IsClaudeCompatUpstreamURL reports whether u targets a registered third-party
// Anthropic-compatible upstream that should use the Claude uTLS fingerprint.
func IsClaudeCompatUpstreamURL(u *url.URL) bool {
	if u == nil || !strings.EqualFold(u.Scheme, "https") {
		return false
	}
	_, ok := claudeCompatHosts.Load(claudeHostKey(u))
	return ok
}
