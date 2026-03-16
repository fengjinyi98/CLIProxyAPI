package auth

import (
	"context"
	"net/http"
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
)

type authGroupTestExecutor struct {
	id string
}

func (e *authGroupTestExecutor) Identifier() string { return e.id }

func (e *authGroupTestExecutor) Execute(_ context.Context, auth *Auth, _ cliproxyexecutor.Request, _ cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	payload := []byte("ok")
	if auth != nil {
		payload = []byte(auth.ID)
	}
	return cliproxyexecutor.Response{Payload: payload}, nil
}

func (e *authGroupTestExecutor) ExecuteStream(context.Context, *Auth, cliproxyexecutor.Request, cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	return nil, nil
}

func (e *authGroupTestExecutor) Refresh(_ context.Context, auth *Auth) (*Auth, error) {
	return auth, nil
}

func (e *authGroupTestExecutor) CountTokens(_ context.Context, auth *Auth, _ cliproxyexecutor.Request, _ cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	payload := []byte("count")
	if auth != nil {
		payload = []byte(auth.ID)
	}
	return cliproxyexecutor.Response{Payload: payload}, nil
}

func (e *authGroupTestExecutor) HttpRequest(context.Context, *Auth, *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestManager_Execute_FiltersAuthsByClientAPIKeyGroup(t *testing.T) {
	mgr := NewManager(nil, nil, nil)
	mgr.SetConfig(&internalconfig.Config{
		AuthGroups: []internalconfig.ClientAuthGroup{
			{Name: "normal", APIKeys: []string{"sk-normal"}},
			{Name: "team", APIKeys: []string{"sk-team"}},
		},
	})
	mgr.RegisterExecutor(&authGroupTestExecutor{id: "claude"})

	normalAuth := &Auth{
		ID:       "auth-normal",
		Provider: "claude",
		Attributes: map[string]string{
			"auth_group": "normal",
		},
	}
	teamAuth := &Auth{
		ID:       "auth-team",
		Provider: "claude",
		Attributes: map[string]string{
			"auth_group": "team",
		},
	}

	if _, err := mgr.Register(context.Background(), normalAuth); err != nil {
		t.Fatalf("register normal auth: %v", err)
	}
	if _, err := mgr.Register(context.Background(), teamAuth); err != nil {
		t.Fatalf("register team auth: %v", err)
	}

	reg := registry.GetGlobalRegistry()
	reg.RegisterClient(normalAuth.ID, "claude", []*registry.ModelInfo{{ID: "test-model"}})
	reg.RegisterClient(teamAuth.ID, "claude", []*registry.ModelInfo{{ID: "test-model"}})
	t.Cleanup(func() {
		reg.UnregisterClient(normalAuth.ID)
		reg.UnregisterClient(teamAuth.ID)
	})

	testCases := []struct {
		name      string
		clientKey string
		wantAuth  string
	}{
		{name: "normal key", clientKey: "sk-normal", wantAuth: "auth-normal"},
		{name: "team key", clientKey: "sk-team", wantAuth: "auth-team"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var selected string
			resp, err := mgr.Execute(
				context.Background(),
				[]string{"claude"},
				cliproxyexecutor.Request{Model: "test-model"},
				cliproxyexecutor.Options{
					Metadata: map[string]any{
						cliproxyexecutor.ClientAPIKeyMetadataKey:         tc.clientKey,
						cliproxyexecutor.SelectedAuthCallbackMetadataKey: func(id string) { selected = id },
					},
				},
			)
			if err != nil {
				t.Fatalf("execute: %v", err)
			}
			if selected != tc.wantAuth {
				t.Fatalf("selected auth = %q, want %q", selected, tc.wantAuth)
			}
			if got := string(resp.Payload); got != tc.wantAuth {
				t.Fatalf("response payload = %q, want %q", got, tc.wantAuth)
			}
		})
	}
}
