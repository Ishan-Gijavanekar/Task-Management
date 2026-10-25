package gateway

import (
	"strings"
	"time"
)

type Service string

const (
	UserService Service = "user"
	TaskService Service = "task"
)

type Route struct {
	Name        string
	Service     Service
	Protected   bool
	Limit       int
	Window      time.Duration
	ComposeUser bool
}

func Match(method, rawPath string) (Route, bool) {
	method = strings.ToUpper(method)
	segments := splitPath(rawPath)
	if len(segments) < 3 || segments[0] != "api" || segments[1] != "v1" {
		return Route{}, false
	}
	resource := segments[2]
	if resource == "auth" {
		return authRoute(method, segments)
	}
	if resource == "me" && len(segments) == 3 && method == "GET" {
		return Route{Name: "getMe", Service: UserService, Protected: true, Limit: 300, Window: time.Minute, ComposeUser: true}, true
	}
	if resource == "users" {
		return userRoute(method, segments)
	}
	if resource == "invitations" && len(segments) == 4 && segments[3] == "accept" && method == "POST" {
		return writeRoute("acceptInvitation", TaskService, true, 30), true
	}
	if resource == "workspaces" {
		return workspaceRoute(method, segments)
	}
	return Route{}, false
}

func authRoute(method string, s []string) (Route, bool) {
	if len(s) != 4 || method != "POST" {
		return Route{}, false
	}
	limits := map[string]int{"register": 10, "login": 10, "refresh": 30, "logout": 30, "forgot-password": 5, "reset-password": 5, "verify-email": 10, "resend-verification": 5}
	limit, ok := limits[s[3]]
	if !ok {
		return Route{}, false
	}
	return Route{Name: "auth." + s[3], Service: UserService, Limit: limit, Window: time.Minute}, true
}
func userRoute(method string, s []string) (Route, bool) {
	if len(s) == 3 && method == "GET" {
		return readRoute("listUsers", UserService, true), true
	}
	if len(s) == 4 && (method == "GET" || method == "PUT" || method == "DELETE") {
		if method == "GET" {
			return readRoute("getUser", UserService, true), true
		}
		return writeRoute("mutateUser", UserService, true, 60), true
	}
	return Route{}, false
}
func workspaceRoute(method string, s []string) (Route, bool) {
	if len(s) == 3 && method == "POST" {
		return writeRoute("createWorkspace", TaskService, true, 60), true
	}
	if len(s) < 4 || s[3] == "" {
		return Route{}, false
	}
	if len(s) == 4 && (method == "GET" || method == "PATCH" || method == "DELETE") {
		if method == "GET" {
			return readRoute("getWorkspace", TaskService, true), true
		}
		return writeRoute("mutateWorkspace", TaskService, true, 60), true
	}
	if len(s) == 5 && s[4] == "transfer" && method == "POST" {
		return writeRoute("transferWorkspace", TaskService, true, 20), true
	}
	if len(s) >= 5 {
		switch s[4] {
		case "projects":
			if len(s) == 5 && (method == "GET" || method == "POST") {
				if method == "GET" {
					return readRoute("listProjects", TaskService, true), true
				}
				return writeRoute("createProject", TaskService, true, 120), true
			}
			if len(s) == 6 && (method == "GET" || method == "PATCH" || method == "DELETE") {
				if method == "GET" {
					return readRoute("getProject", TaskService, true), true
				}
				return writeRoute("mutateProject", TaskService, true, 120), true
			}
		case "invitations":
			if len(s) == 5 && method == "POST" {
				return writeRoute("inviteMember", TaskService, true, 10), true
			}
			if len(s) == 6 && method == "DELETE" {
				return writeRoute("revokeInvitation", TaskService, true, 30), true
			}
			if len(s) == 7 && s[6] == "resend" && method == "POST" {
				return writeRoute("resendInvitation", TaskService, true, 10), true
			}
		case "members":
			if len(s) == 5 && method == "GET" {
				return readRoute("listMembers", TaskService, true), true
			}
			if len(s) == 6 && (method == "PATCH" || method == "DELETE") {
				return writeRoute("mutateMember", TaskService, true, 60), true
			}
		case "audit-events":
			if len(s) == 5 && method == "GET" {
				return readRoute("listAudit", TaskService, true), true
			}
		case "tasks", "tasks:batch":
			if s[4] == "tasks:batch" && len(s) == 5 && method == "PATCH" {
				return writeRoute("batchTasks", TaskService, true, 120), true
			}
			if s[4] == "tasks" {
				if len(s) == 5 && (method == "GET" || method == "POST") {
					if method == "GET" {
						return readRoute("listTasks", TaskService, true), true
					}
					return writeRoute("createTask", TaskService, true, 120), true
				}
				if len(s) == 6 && (method == "GET" || method == "PATCH" || method == "DELETE") {
					if method == "GET" {
						return readRoute("getTask", TaskService, true), true
					}
					return writeRoute("mutateTask", TaskService, true, 120), true
				}
				if len(s) == 7 && (s[6] == "restore" || s[6] == "reorder") && method == "POST" {
					return writeRoute("task."+s[6], TaskService, true, 120), true
				}
			}
		}
	}
	return Route{}, false
}
func splitPath(raw string) []string {
	trimmed := strings.Trim(raw, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil
		}
	}
	return parts
}
func readRoute(name string, service Service, protected bool) Route {
	return Route{Name: name, Service: service, Protected: protected, Limit: 600, Window: time.Minute}
}
func writeRoute(name string, service Service, protected bool, limit int) Route {
	return Route{Name: name, Service: service, Protected: protected, Limit: limit, Window: time.Minute}
}
