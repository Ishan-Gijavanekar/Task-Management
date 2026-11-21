import { useMutation, useQueryClient } from '@tanstack/react-query'
import { type FormEvent, useState } from 'react'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import { useWorkspace } from './workspace-context'

export function WorkspacesPage() {
  const queryClient = useQueryClient()
  const { workspace, workspaces, select, remember, forget } = useWorkspace()
  const [error, setError] = useState('')

  const create = useMutation({
    mutationFn: (name: string) => api.createWorkspace({ name }),
    onSuccess: (created) => {
      remember(created)
      queryClient.setQueryData(['workspace', created.id], created)
    },
    onError: (cause) => setError(errorMessage(cause)),
  })
  const rename = useMutation({
    mutationFn: (name: string) => {
      if (!workspace) throw new Error('No workspace selected')
      return api.updateWorkspace(workspace.id, {
        name,
        version: workspace.version,
      })
    },
    onSuccess: (updated) => {
      remember(updated)
      queryClient.setQueryData(['workspace', updated.id], updated)
    },
    onError: (cause) => setError(errorMessage(cause)),
  })

  function createWorkspace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    const form = event.currentTarget
    create.mutate(String(new FormData(form).get('name')), {
      onSuccess: () => form.reset(),
    })
  }

  function openWorkspace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    const form = event.currentTarget
    const data = new FormData(form)
    const id = String(data.get('workspaceId')).trim()
    api
      .getWorkspace(id)
      .then((found) => {
        remember(found)
        queryClient.setQueryData(['workspace', found.id], found)
        form.reset()
      })
      .catch((cause: unknown) => setError(errorMessage(cause)))
  }

  function renameWorkspace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    rename.mutate(
      String(new FormData(event.currentTarget).get('workspaceName')),
    )
  }

  return (
    <div className="page workspace-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Organization</p>
          <h1>Workspaces</h1>
          <p>Create a new home or reopen a workspace shared with you.</p>
        </div>
      </header>
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      <div className="settings-grid">
        <section className="panel">
          <h2>Your workspaces</h2>
          {workspaces.length ? (
            <div className="workspace-list">
              {workspaces.map((item) => (
                <div
                  className={item.id === workspace?.id ? 'selected' : ''}
                  key={item.id}
                >
                  <button type="button" onClick={() => select(item.id)}>
                    <span className="workspace-avatar">
                      {item.name.slice(0, 1).toUpperCase()}
                    </span>
                    <span>
                      <strong>{item.name}</strong>
                      <small>{item.id}</small>
                    </span>
                  </button>
                  <button
                    className="icon-button"
                    type="button"
                    onClick={() => forget(item.id)}
                    aria-label={`Forget ${item.name}`}
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <p className="muted">
              No workspace has been opened in this browser yet.
            </p>
          )}
        </section>

        <section className="panel form-panel">
          <h2>Create workspace</h2>
          <form onSubmit={createWorkspace}>
            <label className="field" htmlFor="name">
              <span>Name</span>
              <input id="name" name="name" maxLength={100} required />
            </label>
            <button
              className="button primary"
              type="submit"
              disabled={create.isPending}
            >
              {create.isPending ? 'Creating…' : 'Create workspace'}
            </button>
          </form>
          <div className="form-separator">
            <span>or</span>
          </div>
          <form onSubmit={openWorkspace}>
            <label className="field" htmlFor="workspaceId">
              <span>Workspace ID</span>
              <input
                id="workspaceId"
                name="workspaceId"
                placeholder="Paste a shared workspace ID"
                required
              />
            </label>
            <button className="button secondary" type="submit">
              Open workspace
            </button>
          </form>
        </section>
      </div>

      {workspace && (
        <section className="panel workspace-settings">
          <div>
            <p className="eyebrow">Current workspace</p>
            <h2>Workspace settings</h2>
            <p className="muted">
              Changes are protected by optimistic concurrency.
            </p>
          </div>
          <form onSubmit={renameWorkspace}>
            <label className="field" htmlFor="workspaceName">
              <span>Workspace name</span>
              <input
                key={workspace.version}
                id="workspaceName"
                name="workspaceName"
                defaultValue={workspace.name}
                maxLength={100}
                required
              />
            </label>
            <button
              className="button secondary"
              type="submit"
              disabled={rename.isPending}
            >
              {rename.isPending ? 'Saving…' : 'Save name'}
            </button>
          </form>
        </section>
      )}
    </div>
  )
}
