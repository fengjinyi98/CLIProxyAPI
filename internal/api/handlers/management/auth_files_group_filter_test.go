package management

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func TestListAuthFiles_FiltersByAuthGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := coreauth.NewManager(nil, nil, nil)
	if _, err := manager.Register(nil, &coreauth.Auth{
		ID:       "normal-auth",
		Provider: "claude",
		FileName: "normal.json",
		Attributes: map[string]string{
			"path":       "/tmp/normal.json",
			"auth_group": "normal",
		},
	}); err != nil {
		t.Fatalf("register normal auth: %v", err)
	}
	if _, err := manager.Register(nil, &coreauth.Auth{
		ID:       "team-auth",
		Provider: "claude",
		FileName: "team.json",
		Attributes: map[string]string{
			"path":       "/tmp/team.json",
			"auth_group": "team",
		},
	}); err != nil {
		t.Fatalf("register team auth: %v", err)
	}

	handler := NewHandlerWithoutConfigFilePath(&config.Config{}, manager)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("GET", "/v0/management/auth-files?group=team", nil)
	ctx.Request = req

	handler.ListAuthFiles(ctx)

	var payload struct {
		Files []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.Files) != 1 {
		t.Fatalf("files len = %d, want 1", len(payload.Files))
	}
	if got, _ := payload.Files[0]["auth_group"].(string); got != "team" {
		t.Fatalf("auth_group = %q, want %q", got, "team")
	}
	if got, _ := payload.Files[0]["name"].(string); got != "team.json" {
		t.Fatalf("name = %q, want %q", got, "team.json")
	}
}
