// Kept in sync with gateway-service/api/openapi.json by the Phase 12 contract check.
export interface User {
  id: string
  name: string
  email: string
  role: string
  status: 'pending_verification' | 'active' | 'suspended' | 'deleted'
  version: number
  createdAt: string
  updatedAt: string
}

export interface AuthResponse {
  access_token: string
  expires_in: number
  token_type: string
  user: User
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface ErrorEnvelope {
  error: {
    code: string
    message: string
    requestId?: string
    fields?: Record<string, unknown>
  }
}

export interface Workspace {
  id: string
  name: string
  ownerId: string
  version: number
  createdAt: string
  updatedAt: string
}

export interface Project {
  id: string
  workspaceId: string
  key: string
  name: string
  description?: string
  visibility: 'private' | 'workspace'
  version: number
}

export interface Task {
  id: string
  workspaceId: string
  projectId: string
  title: string
  status: 'backlog' | 'todo' | 'in_progress' | 'blocked' | 'done' | 'cancelled'
  priority: 'none' | 'low' | 'medium' | 'high' | 'urgent'
  assigneeId?: string
  dueDate?: string
  version: number
}

export interface Membership {
  userId: string
  workspaceId: string
  role: 'owner' | 'admin' | 'member' | 'viewer'
  status: 'active' | 'suspended'
}

export interface TaskList {
  tasks: Task[]
  nextCursor?: string
}

export interface MessageResponse {
  message: string
}
