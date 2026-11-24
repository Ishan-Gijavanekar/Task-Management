import './members.css'
import { useMutation, useQuery } from '@tanstack/react-query'
import { type FormEvent, useState } from 'react'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import type { InviteRequest } from '../../api/types'
import { LoadingScreen } from '../../ui/states'
import { useWorkspace } from '../workspaces/workspace-context'
import { useWorkspacePermission } from './permissions'

export function MembersPage() {
  const { workspace } = useWorkspace()
  const { canAdmin } = useWorkspacePermission()
  const [error, setError] = useState('')
  const [inviteUrl, setInviteUrl] = useState('')
  const [copied, setCopied] = useState(false)
  const members = useQuery({
    queryKey: ['members', workspace!.id],
    queryFn: ({ signal }) => api.listMembers(workspace!.id, signal),
  })
  const invite = useMutation({
    mutationFn: (input: InviteRequest) =>
      api.inviteMember(workspace!.id, input),
    onSuccess: ({ token }) => {
      setInviteUrl(
        `${window.location.origin}/invitations/accept#token=${encodeURIComponent(token)}`,
      )
      setCopied(false)
    },
    onError: (cause) => setError(errorMessage(cause)),
  })

  function submitInvite(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setInviteUrl('')
    const data = new FormData(event.currentTarget)
    invite.mutate({
      email: String(data.get('email')).trim(),
      role: String(data.get('role')) as InviteRequest['role'],
    })
  }

  async function copyInvite() {
    await navigator.clipboard.writeText(inviteUrl)
    setCopied(true)
  }

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
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      {canAdmin && (
        <section className="panel invite-panel">
          <div>
            <h2>Invite a member</h2>
            <p className="muted">
              Create a private link that expires in seven days.
            </p>
          </div>
          <form onSubmit={submitInvite}>
            <label className="field" htmlFor="invite-email">
              <span>Email</span>
              <input
                id="invite-email"
                name="email"
                type="email"
                autoComplete="email"
                required
              />
            </label>
            <label className="field" htmlFor="invite-role">
              <span>Role</span>
              <select id="invite-role" name="role" defaultValue="member">
                <option value="member">Member</option>
                <option value="viewer">Viewer</option>
                <option value="admin">Admin</option>
              </select>
            </label>
            <button
              className="button primary"
              type="submit"
              disabled={invite.isPending}
            >
              {invite.isPending ? 'Creating…' : 'Create invite'}
            </button>
          </form>
          {inviteUrl && (
            <div className="invite-result" role="status">
              <label className="field" htmlFor="invite-link">
                <span>Share this link privately</span>
                <input id="invite-link" value={inviteUrl} readOnly />
              </label>
              <button
                className="button secondary"
                type="button"
                onClick={copyInvite}
              >
                {copied ? 'Copied' : 'Copy link'}
              </button>
            </div>
          )}
        </section>
      )}
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
