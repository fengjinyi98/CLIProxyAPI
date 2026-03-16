package handlers

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	internalconfig "github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	sdkconfig "github.com/router-for-me/CLIProxyAPI/v6/sdk/config"
)

func TestVisibleModels_FiltersByClientAPIKeyGroup(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	now := time.Now().Unix()

	reg.RegisterClient("visible-normal-auth", "openai", []*registry.ModelInfo{{ID: "model-normal", Created: now}})
	reg.RegisterClient("visible-team-auth", "openai", []*registry.ModelInfo{{ID: "model-team", Created: now + 1}})
	t.Cleanup(func() {
		reg.UnregisterClient("visible-normal-auth")
		reg.UnregisterClient("visible-team-auth")
	})

	manager := coreauth.NewManager(nil, nil, nil)
	manager.SetConfig(&internalconfig.Config{
		AuthGroups: []internalconfig.ClientAuthGroup{
			{Name: "normal", APIKeys: []string{"sk-normal"}},
			{Name: "team", APIKeys: []string{"sk-team"}},
		},
	})
	if _, err := manager.Register(nil, &coreauth.Auth{
		ID:       "visible-normal-auth",
		Provider: "openai",
		Attributes: map[string]string{
			"auth_group": "normal",
		},
	}); err != nil {
		t.Fatalf("register normal auth: %v", err)
	}
	if _, err := manager.Register(nil, &coreauth.Auth{
		ID:       "visible-team-auth",
		Provider: "openai",
		Attributes: map[string]string{
			"auth_group": "team",
		},
	}); err != nil {
		t.Fatalf("register team auth: %v", err)
	}

	handler := NewBaseAPIHandlers(&sdkconfig.SDKConfig{}, manager)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("apiKey", "sk-team")

	models := handler.VisibleModels(ctx, "openai")
	if len(models) != 1 {
		t.Fatalf("visible models len = %d, want 1", len(models))
	}
	if got, _ := models[0]["id"].(string); got != "model-team" {
		t.Fatalf("visible model id = %q, want %q", got, "model-team")
	}
}
