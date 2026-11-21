import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import { useWorkspace } from '../workspaces/workspace-context'

export function AcceptInvitationPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { remember } = useWorkspace()
  const token = useMemo(
    () =>
      new URLSearchParams(location.hash.slice(1)).get('token')?.trim() ?? '',
    [location.hash],
  )
  const accept = useMutation({
    mutationFn: () => api.acceptInvitation(token),
    onSuccess: async (membership) => {
      const workspace = await api.getWorkspace(membership.workspaceId)
      remember(workspace)
      queryClient.setQueryData(['workspace', workspace.id], workspace)
      await queryClient.invalidateQueries({
        queryKey: ['members', workspace.id],
      })
      navigate('/', { replace: true })
    },
  })

  return (
    <div className="page narrow-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Workspace invitation</p>
          <h1>Join the workspace</h1>
          <p>Accept this invitation to access its projects and tasks.</p>
        </div>
      </header>
      <section className="panel status-panel">
        {!token ? (
          <div className="notice error" role="alert">
            This invitation link is missing its token.
          </div>
        ) : accept.isError ? (
          <div className="notice error" role="alert">
            {errorMessage(accept.error)}
          </div>
        ) : (
          <p className="muted">
            Invitation links are single-use and should not be shared publicly.
          </p>
        )}
        <button
          className="button primary"
          type="button"
          disabled={!token || accept.isPending}
          onClick={() => accept.mutate()}
        >
          {accept.isPending ? 'Joining…' : 'Accept invitation'}
        </button>
      </section>
    </div>
  )
}
