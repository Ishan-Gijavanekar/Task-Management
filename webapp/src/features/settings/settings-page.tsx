import './settings.css'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type FormEvent, useState } from 'react'
import { Link, Navigate, Route, Routes } from 'react-router-dom'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import type { LabelList } from '../../api/types'
import { LoadingScreen } from '../../ui/states'
import { useAuth } from '../auth/auth-context'
import { useWorkspacePermission } from '../members/permissions'
import { useWorkspace } from '../workspaces/workspace-context'

export function SettingsPage() {
  return (
    <Routes>
      <Route index element={<SettingsOverview />} />
      <Route path="labels" element={<LabelsPage />} />
      <Route path="*" element={<Navigate replace to="." />} />
    </Routes>
  )
}

function SettingsOverview() {
  const { user } = useAuth()
  const { workspace } = useWorkspace()
  const { membership } = useWorkspacePermission()

  return (
    <div className="page settings-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Preferences</p>
          <h1>Settings</h1>
          <p>Manage your account and the current workspace.</p>
        </div>
      </header>
      <div className="settings-sections">
        <section className="panel settings-summary">
          <div className="panel-heading">
            <div>
              <p className="eyebrow">Account</p>
              <h2>Your profile</h2>
            </div>
          </div>
          <dl className="settings-details">
            <div>
              <dt>Name</dt>
              <dd>{user?.name || 'Not provided'}</dd>
            </div>
            <div>
              <dt>Email</dt>
              <dd>{user?.email}</dd>
            </div>
          </dl>
        </section>

        <section className="panel settings-summary">
          <div className="panel-heading">
            <div>
              <p className="eyebrow">Workspace</p>
              <h2>{workspace?.name}</h2>
            </div>
            <span className="role-chip">{membership?.role ?? 'Member'}</span>
          </div>
          <dl className="settings-details">
            <div>
              <dt>Workspace ID</dt>
              <dd className="settings-id">{workspace?.id}</dd>
            </div>
          </dl>
          <Link className="text-link" to="/workspaces">
            Manage workspaces
          </Link>
        </section>

        <Link className="panel settings-link" to="labels">
          <span className="settings-icon" aria-hidden="true">
            ◇
          </span>
          <span>
            <strong>Labels</strong>
            <small>
              Create reusable labels for organizing and filtering tasks.
            </small>
          </span>
          <span aria-hidden="true">→</span>
        </Link>
        <Link className="panel settings-link" to="/notifications">
          <span className="settings-icon" aria-hidden="true">
            ◎
          </span>
          <span>
            <strong>Notifications</strong>
            <small>Choose which task updates and mentions reach you.</small>
          </span>
          <span aria-hidden="true">→</span>
        </Link>
        <Link className="panel settings-link" to="/members">
          <span className="settings-icon" aria-hidden="true">
            ♙
          </span>
          <span>
            <strong>Members</strong>
            <small>Review workspace access, roles, and invitations.</small>
          </span>
          <span aria-hidden="true">→</span>
        </Link>
      </div>
    </div>
  )
}

function LabelsPage() {
  const { workspace } = useWorkspace()
  const { canWrite } = useWorkspacePermission()
  const queryClient = useQueryClient()
  const [error, setError] = useState('')
  const workspaceId = workspace!.id
  const labels = useQuery({
    queryKey: ['labels', workspaceId],
    queryFn: ({ signal }) => api.listLabels(workspaceId, signal),
  })
  const create = useMutation({
    mutationFn: ({ name, color }: { name: string; color: string }) =>
      api.createLabel(workspaceId, name, color),
    onSuccess: (created) => {
      queryClient.setQueryData<LabelList>(
        ['labels', workspaceId],
        (current) => ({
          labels: [...(current?.labels ?? []), created].sort((left, right) =>
            left.name.localeCompare(right.name),
          ),
        }),
      )
    },
    onError: (cause) => setError(errorMessage(cause)),
  })
  const remove = useMutation({
    mutationFn: (labelId: string) => api.deleteLabel(workspaceId, labelId),
    onSuccess: (_, labelId) => {
      queryClient.setQueryData<LabelList>(
        ['labels', workspaceId],
        (current) => ({
          labels: current?.labels.filter((label) => label.id !== labelId) ?? [],
        }),
      )
      void queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] })
    },
    onError: (cause) => setError(errorMessage(cause)),
  })

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    const form = event.currentTarget
    const data = new FormData(form)
    create.mutate(
      {
        name: String(data.get('name')).trim(),
        color: String(data.get('color')),
      },
      { onSuccess: () => form.reset() },
    )
  }

  return (
    <div className="page narrow-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Settings</p>
          <h1>Labels</h1>
          <p>
            Create labels once, then assign them while creating or editing
            tasks.
          </p>
        </div>
        <Link className="button secondary" to="/settings">
          Back to settings
        </Link>
      </header>
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      {canWrite && (
        <section className="panel label-create-panel">
          <div>
            <h2>Create a label</h2>
            <p className="muted">Names must be unique in this workspace.</p>
          </div>
          <form onSubmit={submit}>
            <label className="field" htmlFor="labelName">
              <span>Name</span>
              <input id="labelName" name="name" maxLength={50} required />
            </label>
            <label className="field color-field" htmlFor="labelColor">
              <span>Color</span>
              <input
                id="labelColor"
                name="color"
                type="color"
                defaultValue="#4f78df"
                required
              />
            </label>
            <button
              className="button primary"
              type="submit"
              disabled={create.isPending}
            >
              {create.isPending ? 'Creating…' : 'Create label'}
            </button>
          </form>
        </section>
      )}
      <section className="panel label-panel">
        <div className="panel-heading">
          <div>
            <h2>Workspace labels</h2>
            <p className="muted">
              Deleting a label also removes it from assigned tasks.
            </p>
          </div>
        </div>
        {labels.isLoading ? (
          <LoadingScreen label="Loading labels" />
        ) : labels.isError ? (
          <div className="notice error" role="alert">
            {errorMessage(labels.error)}
          </div>
        ) : labels.data?.labels.length ? (
          <div className="label-list">
            {labels.data.labels.map((label) => (
              <div key={label.id}>
                <span
                  className="label-swatch"
                  style={{ backgroundColor: label.color }}
                />
                <span>
                  <strong>{label.name}</strong>
                  <small>{label.color}</small>
                </span>
                {canWrite && (
                  <button
                    className="button danger compact"
                    type="button"
                    disabled={remove.isPending && remove.variables === label.id}
                    onClick={() => {
                      if (
                        window.confirm(
                          `Delete “${label.name}”? It will be removed from every task.`,
                        )
                      ) {
                        setError('')
                        remove.mutate(label.id)
                      }
                    }}
                  >
                    {remove.isPending && remove.variables === label.id
                      ? 'Deleting…'
                      : 'Delete'}
                  </button>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className="empty-state label-empty">
            <span className="empty-icon" aria-hidden="true">
              ◇
            </span>
            <h2>No labels yet</h2>
            <p>Create the first label to categorize tasks.</p>
          </div>
        )}
      </section>
    </div>
  )
}
