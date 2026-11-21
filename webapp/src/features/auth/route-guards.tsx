import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { LoadingScreen } from '../../ui/states'
import { useAuth } from './auth-context'
import { safeReturnUrl } from './safe-return'

export function ProtectedRoute() {
  const auth = useAuth()
  const location = useLocation()
  if (auth.status === 'loading')
    return <LoadingScreen label="Restoring your session" />
  if (auth.status === 'anonymous') {
    const returnTo = `${location.pathname}${location.search}${location.hash}`
    return (
      <Navigate
        replace
        to={`/login?returnTo=${encodeURIComponent(returnTo)}`}
      />
    )
  }
  return <Outlet />
}

export function AnonymousRoute() {
  const auth = useAuth()
  const location = useLocation()
  if (auth.status === 'loading')
    return <LoadingScreen label="Checking your session" />
  if (auth.status === 'authenticated') {
    const returnTo = new URLSearchParams(location.search).get('returnTo')
    return <Navigate replace to={safeReturnUrl(returnTo)} />
  }
  return <Outlet />
}
