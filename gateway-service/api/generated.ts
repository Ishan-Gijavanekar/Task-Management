// Generated from api/openapi.json. Do not edit.

export interface AcceptInvitationRequest {
  token: string;
}

export interface AuditEvent {
  action: string;
  actorId: string;
  createdAt: string;
  details?: Record<string, unknown>;
  id: string;
  resource: string;
  resourceId: string;
  workspaceId: string;
}

export interface AuthResponse {
  access_token: string;
  expires_in: number;
  token_type: string;
  user: User;
}

export interface BatchUpdateItem {
  assigneeId?: string;
  id: string;
  priority?: "high" | "low" | "medium" | "none" | "urgent";
  status?: "backlog" | "blocked" | "cancelled" | "done" | "in_progress" | "todo";
  version: number;
}

export interface BatchUpdateTasksRequest {
  items: (BatchUpdateItem)[];
}

export interface ChecklistItem {
  done: boolean;
  id: string;
  text: string;
}

export interface CreateProjectRequest {
  description?: string;
  key: string;
  name: string;
  visibility?: "private" | "workspace";
  workflow?: (WorkflowStatus)[];
}

export interface CreateTaskRequest {
  assigneeId?: string;
  checklist?: (ChecklistItem)[];
  dependencies?: (string)[];
  description?: string;
  dueDate?: string;
  estimateMinutes?: number;
  labelIds?: (string)[];
  position?: number;
  priority?: "high" | "low" | "medium" | "none" | "urgent";
  projectId: string;
  startDate?: string;
  title: string;
}

export interface CreateWorkspaceRequest {
  name: string;
}

export interface EmailRequest {
  email: string;
}

export interface ErrorDetail {
  code: string;
  fields?: Record<string, unknown>;
  message: string;
  requestId: string;
}

export interface ErrorEnvelope {
  error: ErrorDetail;
}

export interface Invitation {
  email: string;
  expiresAt: string;
  id: string;
  role: "admin" | "member" | "viewer";
  version: number;
  workspaceId: string;
}

export interface InvitationResult {
  invitation: Invitation;
  token: string;
}

export interface InviteRequest {
  email: string;
  role: "admin" | "member" | "viewer";
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LogoutRequest {
  all_sessions?: boolean;
  refresh_token?: string;
}

export interface Membership {
  role: "admin" | "member" | "owner" | "viewer";
  status: "active" | "suspended";
  userId: string;
  workspaceId: string;
}

export interface Project {
  description?: string;
  id: string;
  key: string;
  name: string;
  version: number;
  visibility: "private" | "workspace";
  workflow: (WorkflowStatus)[];
  workspaceId: string;
}

export interface RefreshRequest {
  refresh_token?: string;
}

export interface RegisterRequest {
  email: string;
  name: string;
  password: string;
}

export interface ReorderTaskRequest {
  after?: number;
  before?: number;
  version: number;
}

export interface ResetPasswordRequest {
  new_password: string;
  token: string;
}

export interface RestoreTaskRequest {
  version: number;
}

export interface Task {
  assigneeId?: string;
  checklist?: (ChecklistItem)[];
  createdAt: string;
  deletedAt?: string;
  dependencies?: (string)[];
  description?: string;
  dueDate?: string;
  estimateMinutes?: number;
  id: string;
  labelIds?: (string)[];
  position: number;
  priority: "high" | "low" | "medium" | "none" | "urgent";
  projectId: string;
  reporterId: string;
  startDate?: string;
  status: "backlog" | "blocked" | "cancelled" | "done" | "in_progress" | "todo";
  title: string;
  updatedAt: string;
  version: number;
  workspaceId: string;
}

export interface TaskList {
  nextCursor?: string;
  tasks: (Task)[];
}

export interface TokenRequest {
  token: string;
}

export interface TransferWorkspaceRequest {
  newOwnerId: string;
  version: number;
}

export interface UpdateMemberRequest {
  role: "admin" | "member" | "viewer";
  status: "active" | "suspended";
}

export interface UpdateProjectRequest {
  description?: string;
  name: string;
  version: number;
  visibility?: "private" | "workspace";
  workflow?: (WorkflowStatus)[];
}

export interface UpdateTaskRequest {
  assigneeId?: string;
  checklist?: (ChecklistItem)[];
  dependencies?: (string)[];
  description?: string;
  dueDate?: string;
  estimateMinutes?: number;
  labelIds?: (string)[];
  priority?: "high" | "low" | "medium" | "none" | "urgent";
  startDate?: string;
  status?: "backlog" | "blocked" | "cancelled" | "done" | "in_progress" | "todo";
  title: string;
  version: number;
}

export interface UpdateUserRequest {
  email: string;
  name: string;
}

export interface UpdateWorkspaceRequest {
  name: string;
  version: number;
}

export interface User {
  createdAt: string;
  email: string;
  id: string;
  name: string;
  role: string;
  status: "active" | "deleted" | "pending_verification" | "suspended";
  updatedAt: string;
  version: number;
}

export interface UserList {
  limit: number;
  page: number;
  total: number;
  users: (User)[];
}

export interface WorkflowStatus {
  key: string;
  name: string;
}

export interface Workspace {
  archivedAt?: string;
  createdAt: string;
  id: string;
  name: string;
  ownerId: string;
  updatedAt: string;
  version: number;
}

export const apiOperations = {
  acceptInvitation: { method: "POST", path: "/invitations/accept" },
  archiveProject: { method: "DELETE", path: "/workspaces/{workspaceId}/projects/{projectId}" },
  archiveTask: { method: "DELETE", path: "/workspaces/{workspaceId}/tasks/{taskId}" },
  archiveWorkspace: { method: "DELETE", path: "/workspaces/{workspaceId}" },
  batchUpdateTasks: { method: "PATCH", path: "/workspaces/{workspaceId}/tasks:batch" },
  createProject: { method: "POST", path: "/workspaces/{workspaceId}/projects" },
  createTask: { method: "POST", path: "/workspaces/{workspaceId}/tasks" },
  createWorkspace: { method: "POST", path: "/workspaces" },
  deleteUser: { method: "DELETE", path: "/users/{userId}" },
  forgotPassword: { method: "POST", path: "/auth/forgot-password" },
  getCurrentUser: { method: "GET", path: "/me" },
  getProject: { method: "GET", path: "/workspaces/{workspaceId}/projects/{projectId}" },
  getTask: { method: "GET", path: "/workspaces/{workspaceId}/tasks/{taskId}" },
  getUser: { method: "GET", path: "/users/{userId}" },
  getWorkspace: { method: "GET", path: "/workspaces/{workspaceId}" },
  inviteWorkspaceMember: { method: "POST", path: "/workspaces/{workspaceId}/invitations" },
  listProjects: { method: "GET", path: "/workspaces/{workspaceId}/projects" },
  listTasks: { method: "GET", path: "/workspaces/{workspaceId}/tasks" },
  listUsers: { method: "GET", path: "/users" },
  listWorkspaceAuditEvents: { method: "GET", path: "/workspaces/{workspaceId}/audit-events" },
  listWorkspaceMembers: { method: "GET", path: "/workspaces/{workspaceId}/members" },
  login: { method: "POST", path: "/auth/login" },
  logout: { method: "POST", path: "/auth/logout" },
  refreshSession: { method: "POST", path: "/auth/refresh" },
  register: { method: "POST", path: "/auth/register" },
  removeWorkspaceMember: { method: "DELETE", path: "/workspaces/{workspaceId}/members/{userId}" },
  reorderTask: { method: "POST", path: "/workspaces/{workspaceId}/tasks/{taskId}/reorder" },
  resendInvitation: { method: "POST", path: "/workspaces/{workspaceId}/invitations/{invitationId}/resend" },
  resendVerification: { method: "POST", path: "/auth/resend-verification" },
  resetPassword: { method: "POST", path: "/auth/reset-password" },
  restoreTask: { method: "POST", path: "/workspaces/{workspaceId}/tasks/{taskId}/restore" },
  revokeInvitation: { method: "DELETE", path: "/workspaces/{workspaceId}/invitations/{invitationId}" },
  transferWorkspaceOwnership: { method: "POST", path: "/workspaces/{workspaceId}/transfer" },
  updateProject: { method: "PATCH", path: "/workspaces/{workspaceId}/projects/{projectId}" },
  updateTask: { method: "PATCH", path: "/workspaces/{workspaceId}/tasks/{taskId}" },
  updateUser: { method: "PUT", path: "/users/{userId}" },
  updateWorkspace: { method: "PATCH", path: "/workspaces/{workspaceId}" },
  updateWorkspaceMember: { method: "PATCH", path: "/workspaces/{workspaceId}/members/{userId}" },
  verifyEmail: { method: "POST", path: "/auth/verify-email" },
} as const;
