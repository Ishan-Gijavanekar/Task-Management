// Type-only imports keep the browser contract tied to Phase 12's generated OpenAPI types.
export type {
  AuthResponse,
  ErrorEnvelope,
  LoginRequest,
  Membership,
  Project,
  RegisterRequest,
  Task,
  TaskList,
  User,
  Workspace,
} from '../../../gateway-service/api/generated'

export interface MessageResponse {
  message: string
}
