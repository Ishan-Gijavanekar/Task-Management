import { type FormEvent } from 'react'
import type { Membership, Project, TaskRecord } from '../../api/types'

interface TaskFormProps {
  task?: TaskRecord
  projectId?: string | undefined
  projects: Project[]
  members: Membership[]
  pending: boolean
  error: string
  onCancel(): void
  onSubmit(input: Record<string, unknown>): void
}

function objectId() {
  return Array.from(crypto.getRandomValues(new Uint8Array(12)), (value) =>
    value.toString(16).padStart(2, '0'),
  ).join('')
}

function dateValue(value?: string) {
  return value?.slice(0, 10) ?? ''
}

function dateTime(value: FormDataEntryValue | null) {
  const text = String(value ?? '')
  return text ? new Date(`${text}T00:00:00.000Z`).toISOString() : undefined
}

export function TaskForm({
  task,
  projectId,
  projects,
  members,
  pending,
  error,
  onCancel,
  onSubmit,
}: TaskFormProps) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const checklist = String(data.get('checklist'))
      .split('\n')
      .map((text) => text.trim())
      .filter(Boolean)
      .map((text, index) => {
        const existing = task?.checklist?.find((item) => item.text === text)
        return (
          existing ?? {
            id: objectId(),
            text,
            completed: false,
            position: (index + 1) * 1024,
          }
        )
      })
    onSubmit({
      ...(!task && { projectId: String(data.get('projectId')) }),
      title: String(data.get('title')).trim(),
      description: String(data.get('description')).trim(),
      status: task ? String(data.get('status')) : undefined,
      priority: String(data.get('priority')),
      assigneeId: String(data.get('assigneeId')),
      startDate: dateTime(data.get('startDate')),
      dueDate: dateTime(data.get('dueDate')),
      estimateMinutes: Number(data.get('estimateMinutes') || 0),
      labelIds: String(data.get('labelIds'))
        .split(',')
        .map((value) => value.trim())
        .filter(Boolean),
      dependencies: String(data.get('dependencies'))
        .split(',')
        .map((value) => value.trim())
        .filter(Boolean),
      checklist,
      ...(task ? { version: task.version } : {}),
    })
  }

  return (
    <form className="task-form" onSubmit={submit}>
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      <div className="form-grid">
        <label className="field span-2" htmlFor="taskTitle">
          <span>Title</span>
          <input
            id="taskTitle"
            name="title"
            defaultValue={task?.title}
            maxLength={300}
            autoFocus
            required
          />
        </label>
        <label className="field span-2" htmlFor="taskDescription">
          <span>Description</span>
          <textarea
            id="taskDescription"
            name="description"
            defaultValue={task?.description}
            rows={4}
          />
        </label>
        <label className="field" htmlFor="taskProject">
          <span>Project</span>
          <select
            id="taskProject"
            name="projectId"
            defaultValue={task?.projectId ?? projectId}
            disabled={Boolean(task)}
            required
          >
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.key} · {project.name}
              </option>
            ))}
          </select>
        </label>
        {task && (
          <label className="field" htmlFor="taskStatus">
            <span>Status</span>
            <select id="taskStatus" name="status" defaultValue={task.status}>
              <option value="backlog">Backlog</option>
              <option value="todo">To do</option>
              <option value="in_progress">In progress</option>
              <option value="blocked">Blocked</option>
              <option value="done">Done</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </label>
        )}
        <label className="field" htmlFor="taskPriority">
          <span>Priority</span>
          <select
            id="taskPriority"
            name="priority"
            defaultValue={task?.priority ?? 'none'}
          >
            <option value="none">No priority</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </select>
        </label>
        <label className="field" htmlFor="taskAssignee">
          <span>Assignee</span>
          <select
            id="taskAssignee"
            name="assigneeId"
            defaultValue={task?.assigneeId ?? ''}
          >
            <option value="">Unassigned</option>
            {members
              .filter((member) => member.status === 'active')
              .map((member) => (
                <option key={member.userId} value={member.userId}>
                  {member.userName || member.userId}
                </option>
              ))}
          </select>
        </label>
        <label className="field" htmlFor="taskStart">
          <span>Start date</span>
          <input
            id="taskStart"
            name="startDate"
            type="date"
            defaultValue={dateValue(task?.startDate)}
          />
        </label>
        <label className="field" htmlFor="taskDue">
          <span>Due date</span>
          <input
            id="taskDue"
            name="dueDate"
            type="date"
            defaultValue={dateValue(task?.dueDate)}
          />
        </label>
        <label className="field" htmlFor="taskEstimate">
          <span>Estimate (minutes)</span>
          <input
            id="taskEstimate"
            name="estimateMinutes"
            type="number"
            min="0"
            defaultValue={task?.estimateMinutes ?? 0}
          />
        </label>
        <label className="field" htmlFor="taskLabels">
          <span>Label IDs</span>
          <input
            id="taskLabels"
            name="labelIds"
            defaultValue={task?.labelIds?.join(', ')}
            placeholder="Comma-separated IDs"
          />
        </label>
        <label className="field span-2" htmlFor="taskChecklist">
          <span>Checklist</span>
          <textarea
            id="taskChecklist"
            name="checklist"
            rows={3}
            defaultValue={task?.checklist?.map((item) => item.text).join('\n')}
            placeholder="One checklist item per line"
          />
        </label>
        <label className="field span-2" htmlFor="taskDependencies">
          <span>Dependencies</span>
          <input
            id="taskDependencies"
            name="dependencies"
            defaultValue={task?.dependencies
              ?.map((item) => item.dependsOnTaskId)
              .join(', ')}
            placeholder="Comma-separated task IDs"
          />
        </label>
      </div>
      <div className="dialog-actions">
        <button className="button secondary" type="button" onClick={onCancel}>
          Cancel
        </button>
        <button className="button primary" type="submit" disabled={pending}>
          {pending ? 'Saving…' : task ? 'Save changes' : 'Create task'}
        </button>
      </div>
    </form>
  )
}
