import { useQuery } from '@tanstack/react-query'
import { api } from '../../api/client'
import { useAuth } from '../auth/auth-context'
import { useWorkspace } from '../workspaces/workspace-context'

export function useWorkspacePermission() {
  const { user } = useAuth()
  const { workspace } = useWorkspace()
  const members = useQuery({
    queryKey: ['members', workspace?.id],
    queryFn: ({ signal }) => api.listMembers(workspace!.id, signal),
    enabled: Boolean(workspace),
  })
  const membership = members.data?.members.find(
    (item) => item.userId === user?.id,
  )
  return {
    membership,
    canWrite: membership?.status === 'active' && membership.role !== 'viewer',
    canAdmin:
      membership?.status === 'active' &&
      ['owner', 'admin'].includes(membership.role),
    loading: members.isLoading,
  }
}
