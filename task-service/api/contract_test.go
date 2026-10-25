package api_test

import (
	"os"
	"strings"
	"testing"

	taskv1 "github.com/Ishan-Gijavanekar/task-service/api/proto/task/v1"
)

func TestAdministrationRESTContractIncludesEveryOperation(t *testing.T) {
	raw, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	contract := string(raw)
	operations := []string{
		"createWorkspace", "getWorkspace", "updateWorkspace", "archiveWorkspace", "transferWorkspaceOwnership",
		"createProject", "getProject", "listProjects", "updateProject", "archiveProject",
		"inviteWorkspaceMember", "resendInvitation", "revokeInvitation", "acceptInvitation",
		"listWorkspaceMembers", "updateWorkspaceMember", "removeWorkspaceMember", "listWorkspaceAuditEvents",
	}
	for _, operation := range operations {
		if !strings.Contains(contract, "operationId: "+operation) {
			t.Errorf("OpenAPI operation %s is missing", operation)
		}
	}
}

func TestAdministrationGRPCContractIncludesEveryMethod(t *testing.T) {
	want := []string{"CreateWorkspace", "GetWorkspace", "UpdateWorkspace", "ArchiveWorkspace", "TransferOwnership", "CreateProject", "GetProject", "ListProjects", "UpdateProject", "ArchiveProject", "InviteMember", "ResendInvitation", "RevokeInvitation", "AcceptInvitation", "ListMembers", "UpdateMember", "RemoveMember", "ListAuditEvents"}
	methods := map[string]bool{}
	for _, method := range taskv1.WorkspaceService_ServiceDesc.Methods {
		methods[method.MethodName] = true
	}
	for _, method := range want {
		if !methods[method] {
			t.Errorf("gRPC method %s is missing", method)
		}
	}
}
