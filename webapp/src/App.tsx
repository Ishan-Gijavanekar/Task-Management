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
import { NotificationsPage } from './features/notifications/notifications-page'
import { ProjectsPage } from './features/projects/projects-page'
import { SettingsPage } from './features/settings/settings-page'
import { AppShell } from './features/shell/app-shell'
import { TasksPage } from './features/tasks/tasks-page'
import { WorkspacesPage } from './features/workspaces/workspaces-page'
import './App.css'

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
          <Route index element={<DashboardPage />} />
          <Route path="workspaces" element={<WorkspacesPage />} />
          <Route path="projects" element={<ProjectsPage />} />
          <Route path="tasks/*" element={<TasksPage />} />
          <Route path="members" element={<MembersPage />} />
          <Route path="notifications" element={<NotificationsPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate replace to="/" />} />
    </Routes>
  )
}
