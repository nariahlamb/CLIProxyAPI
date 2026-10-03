package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
)

// GetCooldowns lists all per-client model cooldowns (quota-exceeded) and
// suspensions currently held by the model registry.
//
// Endpoint: GET /v0/management/cooldowns
func (h *Handler) GetCooldowns(c *gin.Context) {
	reg := registry.GetGlobalRegistry()
	entries := reg.CooldownSnapshot()
	c.JSON(http.StatusOK, gin.H{"cooldowns": entries, "count": len(entries)})
}

// ClearCooldowns removes cooldown marks. Body (all optional):
//   - client_id: clear only this client's marks
//   - model:     restrict clearing to this model
//   - all:       when true (or no client_id given), clear everything
//
// Endpoint: POST /v0/management/cooldowns/clear
func (h *Handler) ClearCooldowns(c *gin.Context) {
	var body struct {
		ClientID string `json:"client_id"`
		Model    string `json:"model"`
		All      bool   `json:"all"`
	}
	_ = c.ShouldBindJSON(&body) // empty body is allowed = clear all
	reg := registry.GetGlobalRegistry()
	clientID := strings.TrimSpace(body.ClientID)
	model := strings.TrimSpace(body.Model)
	var cleared int
	if body.All || clientID == "" {
		cleared = reg.ClearAllCooldowns()
	} else {
		cleared = reg.ClearClientCooldowns(clientID, model)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "cleared": cleared})
}
