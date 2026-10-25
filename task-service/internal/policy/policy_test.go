package policy

import (
	"testing"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
)

func TestRoleActionMatrix(t *testing.T) {
	tests := []struct {
		role    domain.WorkspaceRole
		allowed map[Action]bool
	}{
		{domain.RoleOwner, map[Action]bool{WorkspaceRead: true, WorkspaceUpdate: true, WorkspaceArchive: true, WorkspaceTransfer: true, ProjectRead: true, ProjectWrite: true, MemberRead: true, MemberManage: true, InvitationManage: true, AuditRead: true, TaskRead: true, TaskWrite: true}},
		{domain.RoleAdmin, map[Action]bool{WorkspaceRead: true, WorkspaceUpdate: true, ProjectRead: true, ProjectWrite: true, MemberRead: true, MemberManage: true, InvitationManage: true, AuditRead: true, TaskRead: true, TaskWrite: true}},
		{domain.RoleMember, map[Action]bool{WorkspaceRead: true, ProjectRead: true, MemberRead: true, TaskRead: true, TaskWrite: true}},
		{domain.RoleViewer, map[Action]bool{WorkspaceRead: true, ProjectRead: true, MemberRead: true, TaskRead: true}},
	}
	for _, tt := range tests {
		m := &domain.Membership{Role: tt.role, Status: domain.MembershipActive}
		for _, action := range Actions() {
			if got := Allows(m, action); got != tt.allowed[action] {
				t.Errorf("role %s action %s: got %v want %v", tt.role, action, got, tt.allowed[action])
			}
		}
	}
}

func TestSuspendedAndMissingMembershipDenyEveryAction(t *testing.T) {
	for _, action := range Actions() {
		if Allows(nil, action) || Allows(&domain.Membership{Role: domain.RoleOwner, Status: domain.MembershipSuspended}, action) {
			t.Errorf("action %s must be denied", action)
		}
	}
}
