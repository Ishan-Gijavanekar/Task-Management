// Type-only imports keep the browser contract tied to Phase 12's generated OpenAPI types.
export type {
  AuthResponse,
  CreateProjectRequest,
  CreateTaskRequest,
  CreateWorkspaceRequest,
  ErrorEnvelope,
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

export interface MessageResponse {
  message: string
}
