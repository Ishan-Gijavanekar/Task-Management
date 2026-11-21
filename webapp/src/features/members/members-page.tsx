import { useQuery } from '@tanstack/react-query'
import { api } from '../../api/client'
import { LoadingScreen } from '../../ui/states'
import { useWorkspace } from '../workspaces/workspace-context'
import { useWorkspacePermission } from './permissions'

export function MembersPage() {
  const { workspace } = useWorkspace()
  const { canAdmin } = useWorkspacePermission()
  const members = useQuery({
    queryKey: ['members', workspace!.id],
    queryFn: ({ signal }) => api.listMembers(workspace!.id, signal),
  })
  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Team</p>
          <h1>Members</h1>
          <p>See who can contribute to {workspace!.name}.</p>
        </div>
        {canAdmin && (
          <span className="permission-note">
            Admin controls use backend-enforced roles
          </span>
        )}
      </header>
      {members.isLoading ? (
        <LoadingScreen label="Loading members" />
      ) : (
        <section className="panel member-table">
          <div className="member-row heading">
            <span>Member</span>
            <span>Role</span>
            <span>Status</span>
          </div>
          {members.data?.members.map((member) => (
            <div className="member-row" key={member.userId}>
              <span>
                <span className="profile-avatar">
                  {(member.userName || member.userId).slice(0, 2).toUpperCase()}
                </span>
                <strong>{member.userName || member.userId}</strong>
              </span>
              <span className="role-chip">{member.role}</span>
              <span className={`membership-status ${member.status}`}>
                {member.status}
              </span>
            </div>
          ))}
        </section>
      )}
    </div>
  )
}
