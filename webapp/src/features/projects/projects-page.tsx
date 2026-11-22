import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type FormEvent, useState } from 'react'
import {
  Link,
  Navigate,
  Route,
  Routes,
  useParams,
  useSearchParams,
} from 'react-router-dom'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import { LoadingScreen } from '../../ui/states'
import { useWorkspacePermission } from '../members/permissions'
import { TaskExplorer } from '../tasks/tasks-page'
import { useWorkspace } from '../workspaces/workspace-context'

const defaultWorkflow = [
  { key: 'backlog', name: 'Backlog' },
  { key: 'todo', name: 'To do' },
  { key: 'in_progress', name: 'In progress' },
  { key: 'blocked', name: 'Blocked' },
  { key: 'done', name: 'Done' },
  { key: 'cancelled', name: 'Cancelled' },
]

export function ProjectsPage() {
  return (
    <Routes>
      <Route index element={<ProjectList />} />
      <Route path=":projectId/settings" element={<ProjectSettings />} />
      <Route path=":projectId/:view" element={<ProjectTasks />} />
      <Route path="*" element={<Navigate replace to="." />} />
    </Routes>
  )
}

function ProjectList() {
  const { workspace } = useWorkspace()
  const { canWrite } = useWorkspacePermission()
  const queryClient = useQueryClient()
  const [search, setSearch] = useSearchParams()
  const [open, setOpen] = useState(search.get('new') === 'true')
  const [error, setError] = useState('')
  const projects = useQuery({
    queryKey: ['projects', workspace!.id],
    queryFn: ({ signal }) => api.listProjects(workspace!.id, signal),
  })
  const create = useMutation({
    mutationFn: (input: { name: string; key: string; description: string }) =>
      api.createProject(workspace!.id, {
        ...input,
        visibility: 'workspace',
        workflow: defaultWorkflow,
      }),
    onSuccess: async () => {
      closeDialog()
      await queryClient.invalidateQueries({
        queryKey: ['projects', workspace!.id],
      })
    },
    onError: (cause) => setError(errorMessage(cause)),
  })

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    const data = new FormData(event.currentTarget)
    create.mutate({
      name: String(data.get('name')).trim(),
      key: String(data.get('key')).trim().toUpperCase(),
      description: String(data.get('description')).trim(),
    })
  }

  function closeDialog() {
    setOpen(false)
    if (search.has('new')) {
      const next = new URLSearchParams(search)
      next.delete('new')
      setSearch(next, { replace: true })
    }
  }

  return (
    <div className="page">
      <header className="page-header project-header">
        <div>
          <p className="eyebrow">Planning</p>
          <h1>Projects</h1>
          <p>Turn outcomes into a visible plan.</p>
        </div>
        {canWrite && (
          <button
            className="button primary"
            type="button"
            onClick={() => {
              setError('')
              setOpen(true)
            }}
          >
            ＋ New project
          </button>
        )}
      </header>
      {projects.isLoading ? (
        <LoadingScreen label="Loading projects" />
      ) : projects.data?.projects.length ? (
        <div className="project-grid">
          {projects.data.projects.map((project, index) => (
            <article className="project-card panel" key={project.id}>
              <span className={`project-symbol tone-${index % 4}`}>
                {project.key.slice(0, 2)}
              </span>
              <div>
                <div className="project-title">
                  <h2>{project.name}</h2>
                  <span>{project.visibility}</span>
                </div>
                <p>{project.description || 'No description yet.'}</p>
              </div>
              <div className="project-actions">
                <Link className="button secondary" to={`${project.id}/board`}>
                  Open board
                </Link>
                <Link
                  className="icon-button"
                  aria-label={`Settings for ${project.name}`}
                  to={`${project.id}/settings`}
                >
                  ⚙
                </Link>
              </div>
            </article>
          ))}
        </div>
      ) : (
        <section className="empty-state panel wide">
          <span className="empty-icon" aria-hidden="true">
            ▦
          </span>
          <h2>No projects yet</h2>
          <p>Create the first project to start organizing tasks.</p>
          {canWrite && (
            <button
              className="button primary"
              type="button"
              onClick={() => {
                setError('')
                setOpen(true)
              }}
            >
              Create project
            </button>
          )}
        </section>
      )}
      {open && (
        <div
          className="dialog-backdrop"
          role="presentation"
          onMouseDown={closeDialog}
        >
          <section
            className="dialog-card"
            role="dialog"
            aria-modal="true"
            aria-labelledby="new-project-title"
            onMouseDown={(event) => event.stopPropagation()}
          >
            <div className="dialog-heading">
              <div>
                <p className="eyebrow">New project</p>
                <h2 id="new-project-title">Shape the next outcome</h2>
              </div>
              <button
                className="icon-button"
                type="button"
                onClick={closeDialog}
                aria-label="Close"
              >
                ×
              </button>
            </div>
            {error && (
              <div className="notice error" role="alert">
                {error}
              </div>
            )}
            <form className="stacked-form" onSubmit={submit}>
              <label className="field" htmlFor="name">
                <span>Name</span>
                <input
                  id="name"
                  name="name"
                  maxLength={100}
                  autoFocus
                  required
                />
              </label>
              <div className="field">
                <label htmlFor="key">Key</label>
                <input
                  id="key"
                  name="key"
                  pattern="[A-Za-z][A-Za-z0-9]{1,9}"
                  placeholder="WEB"
                  maxLength={10}
                  aria-describedby="key-hint"
                  required
                />
                <small className="field-hint" id="key-hint">
                  2–10 letters or numbers, starting with a letter.
                </small>
              </div>
              <label className="field" htmlFor="description">
                <span>Description</span>
                <textarea
                  id="description"
                  name="description"
                  rows={3}
                  maxLength={2000}
                />
              </label>
              <div className="dialog-actions">
                <button
                  className="button secondary"
                  type="button"
                  onClick={closeDialog}
                >
                  Cancel
                </button>
                <button
                  className="button primary"
                  type="submit"
                  disabled={create.isPending}
                >
                  {create.isPending ? 'Creating…' : 'Create project'}
                </button>
              </div>
            </form>
          </section>
        </div>
      )}
    </div>
  )
}

function ProjectTasks() {
  const { projectId = '', view = 'board' } = useParams()
  if (!['board', 'list'].includes(view))
    return <Navigate replace to={`/projects/${projectId}/board`} />
  return <TaskExplorer projectId={projectId} view={view as 'board' | 'list'} />
}

function ProjectSettings() {
  const { workspace } = useWorkspace()
  const { projectId = '' } = useParams()
  const { canWrite } = useWorkspacePermission()
  const queryClient = useQueryClient()
  const [error, setError] = useState('')
  const project = useQuery({
    queryKey: ['project', workspace!.id, projectId],
    queryFn: ({ signal }) => api.getProject(workspace!.id, projectId, signal),
  })
  const update = useMutation({
    mutationFn: (input: {
      name: string
      description: string
      visibility: 'private' | 'workspace'
    }) =>
      api.updateProject(workspace!.id, projectId, {
        ...input,
        version: project.data!.version,
        workflow: project.data!.workflow,
      }),
    onSuccess: (updated) => {
      queryClient.setQueryData(['project', workspace!.id, projectId], updated)
      void queryClient.invalidateQueries({
        queryKey: ['projects', workspace!.id],
      })
    },
    onError: (cause) => setError(errorMessage(cause)),
  })
  if (project.isLoading)
    return <LoadingScreen label="Loading project settings" />
  if (!project.data)
    return (
      <div className="page">
        <div className="notice error">Project could not be loaded.</div>
      </div>
    )

  return (
    <div className="page task-page">
      <header className="page-header task-header">
        <div>
          <p className="eyebrow">{project.data.key}</p>
          <h1>Project settings</h1>
          <p>Update how this project appears to the workspace.</p>
        </div>
      </header>
      <nav className="view-tabs" aria-label="Project views">
        <Link to={`/projects/${projectId}/board`}>Board</Link>
        <Link to={`/projects/${projectId}/list`}>List</Link>
        <Link className="active" to={`/projects/${projectId}/settings`}>
          Settings
        </Link>
      </nav>
      <section className="panel form-panel project-settings-panel">
        {error && (
          <div className="notice error" role="alert">
            {error}
          </div>
        )}
        <form
          className="stacked-form"
          onSubmit={(event) => {
            event.preventDefault()
            const data = new FormData(event.currentTarget)
            update.mutate({
              name: String(data.get('name')).trim(),
              description: String(data.get('description')).trim(),
              visibility: String(data.get('visibility')) as
                'private' | 'workspace',
            })
          }}
        >
          <label className="field" htmlFor="projectName">
            <span>Name</span>
            <input
              id="projectName"
              name="name"
              defaultValue={project.data.name}
              disabled={!canWrite}
              maxLength={100}
              required
            />
          </label>
          <label className="field" htmlFor="projectDescription">
            <span>Description</span>
            <textarea
              id="projectDescription"
              name="description"
              rows={4}
              defaultValue={project.data.description}
              disabled={!canWrite}
              maxLength={2000}
            />
          </label>
          <label className="field" htmlFor="visibility">
            <span>Visibility</span>
            <select
              id="visibility"
              name="visibility"
              defaultValue={project.data.visibility}
              disabled={!canWrite}
            >
              <option value="workspace">Everyone in workspace</option>
              <option value="private">Owners and admins</option>
            </select>
          </label>
          {canWrite && (
            <button
              className="button primary"
              type="submit"
              disabled={update.isPending}
            >
              {update.isPending ? 'Saving…' : 'Save project'}
            </button>
          )}
        </form>
      </section>
    </div>
  )
}
