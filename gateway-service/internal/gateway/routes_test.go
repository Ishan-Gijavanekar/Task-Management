package gateway

import "testing"

func TestPublicRouteAllowlist(t *testing.T) {
	tests := []struct {
		method, path string
		service      Service
		protected    bool
	}{{"POST", "/api/v1/auth/login", UserService, false}, {"GET", "/api/v1/me", UserService, true}, {"GET", "/api/v1/users/user-1", UserService, true}, {"POST", "/api/v1/workspaces", TaskService, true}, {"GET", "/api/v1/workspaces/w/projects", TaskService, true}, {"PATCH", "/api/v1/workspaces/w/tasks:batch", TaskService, true}, {"POST", "/api/v1/workspaces/w/tasks/t/reorder", TaskService, true}, {"POST", "/api/v1/invitations/accept", TaskService, true}}
	for _, tt := range tests {
		route, ok := Match(tt.method, tt.path)
		if !ok || route.Service != tt.service || route.Protected != tt.protected {
			t.Errorf("%s %s: %+v %v", tt.method, tt.path, route, ok)
		}
	}
}
func TestRejectsUnknownRoutesAndMethods(t *testing.T) {
	for _, tt := range [][2]string{{"GET", "/api/v2/workspaces"}, {"GET", "/api/v1/auth/login"}, {"POST", "/api/v1/unknown"}, {"GET", "/api/v1/workspaces//tasks"}, {"GET", "/api/v1/workspaces/w/tasks/t/unknown"}, {"CONNECT", "/api/v1/users/x"}} {
		if _, ok := Match(tt[0], tt[1]); ok {
			t.Errorf("allowed %s %s", tt[0], tt[1])
		}
	}
}
