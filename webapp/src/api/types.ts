// Type-only imports keep the browser contract tied to Phase 12's generated OpenAPI types.
export type {
  AuthResponse,
  CreateProjectRequest,
  CreateTaskRequest,
  CreateWorkspaceRequest,
  ErrorEnvelope,
  Label,
  LoginRequest,
  Membership,
  Project,
  RegisterRequest,
  ReorderTaskRequest,
  Task,
  TaskList,
  UpdateProjectRequest,
  UpdateTaskRequest,
  UpdateWorkspaceRequest,
  User,
  WorkflowStatus,
  Workspace,
} from '../../../gateway-service/api/generated'

import type { Task } from '../../../gateway-service/api/generated'

export interface ChecklistItem {
  id: string
  text: string
  completed: boolean
  position: number
}

export interface TaskDependency {
  taskId: string
  dependsOnTaskId: string
}

export type TaskRecord = Omit<Task, 'checklist' | 'dependencies'> & {
  checklist?: ChecklistItem[]
  dependencies?: TaskDependency[]
}

export interface TaskPage {
  tasks: TaskRecord[]
  nextCursor?: string
}

export interface ProjectList {
  projects: import('../../../gateway-service/api/generated').Project[]
}

export interface MemberList {
  members: import('../../../gateway-service/api/generated').Membership[]
}

export interface LabelList {
  labels: import('../../../gateway-service/api/generated').Label[]
}

export interface CommentRecord {
  id: string
  workspaceId: string
  taskId: string
  authorId: string
  body: string
  mentionIds?: string[]
  version: number
  createdAt: string
  updatedAt: string
}

export interface CommentPage {
  comments: CommentRecord[]
  nextCursor?: string
}

export interface ActivityRecord {
  id: string
  taskId: string
  actorId: string
  type: string
  version: number
  createdAt: string
}

export interface ActivityPage {
  activities: ActivityRecord[]
  nextCursor?: string
}

export interface NotificationRecord {
  id: string
  workspaceId: string
  taskId: string
  actorId: string
  type: string
  message: string
  readAt?: string
  createdAt: string
}

export interface NotificationList {
  notifications: NotificationRecord[]
}

export interface NotificationPreferences {
  userId: string
  mentions: boolean
  taskUpdates: boolean
  emailEnabled: boolean
}

export interface MessageResponse {
  message: string
}
