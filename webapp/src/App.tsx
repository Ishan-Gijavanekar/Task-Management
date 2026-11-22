import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthLayout } from './features/auth/auth-layout'
import {
  ForgotPasswordPage,
  LoginPage,
  RegisterPage,
  ResendVerificationPage,
  ResetPasswordPage,
  VerifyEmailPage,
} from './features/auth/auth-pages'
import { AnonymousRoute, ProtectedRoute } from './features/auth/route-guards'
import { DashboardPage } from './features/dashboard/dashboard-page'
import { MembersPage } from './features/members/members-page'
import { AcceptInvitationPage } from './features/invitations/accept-invitation-page'
import { NotificationsPage } from './features/notifications/notifications-page'
import { ProjectsPage } from './features/projects/projects-page'
import { SettingsPage } from './features/settings/settings-page'
import { AppShell } from './features/shell/app-shell'
import { TasksPage } from './features/tasks/tasks-page'
import { WorkspacesPage } from './features/workspaces/workspaces-page'
import { WorkspaceGate } from './features/workspaces/workspace-gate'
import './App.css'
import './Phase14.css'

export default function App() {
  return (
    <Routes>
      <Route element={<AnonymousRoute />}>
        <Route element={<AuthLayout />}>
          <Route path="login" element={<LoginPage />} />
          <Route path="register" element={<RegisterPage />} />
          <Route path="forgot-password" element={<ForgotPasswordPage />} />
          <Route path="reset-password" element={<ResetPasswordPage />} />
          <Route path="verify-email" element={<VerifyEmailPage />} />
          <Route
            path="resend-verification"
            element={<ResendVerificationPage />}
          />
        </Route>
      </Route>
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route
            index
            element={
              <WorkspaceGate>
                <DashboardPage />
              </WorkspaceGate>
            }
          />
          <Route path="workspaces" element={<WorkspacesPage />} />
          <Route path="invitations/accept" element={<AcceptInvitationPage />} />
          <Route
            path="projects/*"
            element={
              <WorkspaceGate>
                <ProjectsPage />
              </WorkspaceGate>
            }
          />
          <Route
            path="tasks/*"
            element={
              <WorkspaceGate>
                <TasksPage />
              </WorkspaceGate>
            }
          />
          <Route
            path="members"
            element={
              <WorkspaceGate>
                <MembersPage />
              </WorkspaceGate>
            }
          />
          <Route path="notifications" element={<NotificationsPage />} />
          <Route
            path="settings/*"
            element={
              <WorkspaceGate>
                <SettingsPage />
              </WorkspaceGate>
            }
          />
        </Route>
      </Route>
      <Route path="*" element={<Navigate replace to="/" />} />
    </Routes>
  )
}
