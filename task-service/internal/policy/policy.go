package policy

import "github.com/Ishan-Gijavanekar/task-service/internal/domain"

type Action string

const (
	WorkspaceRead     Action = "workspace.read"
	WorkspaceUpdate   Action = "workspace.update"
	WorkspaceArchive  Action = "workspace.archive"
	WorkspaceTransfer Action = "workspace.transfer"
	ProjectRead       Action = "project.read"
	ProjectWrite      Action = "project.write"
	MemberRead        Action = "member.read"
	MemberManage      Action = "member.manage"
	InvitationManage  Action = "invitation.manage"
	AuditRead         Action = "audit.read"
	TaskRead          Action = "task.read"
	TaskWrite         Action = "task.write"
)

var grants = map[domain.WorkspaceRole]map[Action]bool{
	domain.RoleOwner: {
		WorkspaceRead: true, WorkspaceUpdate: true, WorkspaceArchive: true, WorkspaceTransfer: true,
		ProjectRead: true, ProjectWrite: true, MemberRead: true, MemberManage: true,
		InvitationManage: true, AuditRead: true, TaskRead: true, TaskWrite: true,
	},
	domain.RoleAdmin: {
		WorkspaceRead: true, WorkspaceUpdate: true, ProjectRead: true, ProjectWrite: true,
		MemberRead: true, MemberManage: true, InvitationManage: true, AuditRead: true,
		TaskRead: true, TaskWrite: true,
	},
	domain.RoleMember: {
		WorkspaceRead: true, ProjectRead: true, MemberRead: true, TaskRead: true, TaskWrite: true,
	},
	domain.RoleViewer: {
		WorkspaceRead: true, ProjectRead: true, MemberRead: true, TaskRead: true,
	},
}

func Allows(m *domain.Membership, action Action) bool {
	return m != nil && m.Status == domain.MembershipActive && grants[m.Role][action]
}

func Actions() []Action {
	return []Action{WorkspaceRead, WorkspaceUpdate, WorkspaceArchive, WorkspaceTransfer, ProjectRead, ProjectWrite, MemberRead, MemberManage, InvitationManage, AuditRead, TaskRead, TaskWrite}
}
