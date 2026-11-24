import { useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import './design-lab.css'

type Status = 'Backlog' | 'To do' | 'In progress' | 'Blocked' | 'Done'
type Priority = 'Urgent' | 'High' | 'Medium' | 'Low'
type Field = 'title' | 'project' | 'status' | 'priority' | 'dueDate'
type Task = {
  id: string
  title: string
  description: string
  project: string
  assignee: string
  status: Status
  priority: Priority
  dueDate: string
}
type Editor = Omit<Task, 'id'>
type ErrorField = Partial<
  Record<'title' | 'project' | 'dueDate' | 'assignee', string>
>
const KEY = 'task-management.ui-lab.v1'
const statuses: Status[] = [
  'Backlog',
  'To do',
  'In progress',
  'Blocked',
  'Done',
]
const priorities: Priority[] = ['Urgent', 'High', 'Medium', 'Low']
const day = (offset: number) => {
  const date = new Date()
  date.setDate(date.getDate() + offset)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
const seed: Task[] = [
  {
    id: 'TSK-1042',
    title: 'Review workspace invitation permissions',
    description: 'Verify member roles and invitation expiry behavior.',
    project: 'Platform',
    assignee: 'Ananya Shah',
    status: 'In progress',
    priority: 'High',
    dueDate: day(2),
  },
  {
    id: 'TSK-1041',
    title: 'Finalize dashboard analytics copy',
    description: 'Ensure task summary labels explain what is counted.',
    project: 'Product design',
    assignee: 'Rohan Mehta',
    status: 'To do',
    priority: 'Medium',
    dueDate: day(4),
  },
  {
    id: 'TSK-1040',
    title: 'Investigate duplicate task events under retry',
    description: 'Check idempotency keys and concurrent edits.',
    project: 'Platform',
    assignee: 'Maya Iyer',
    status: 'Blocked',
    priority: 'Urgent',
    dueDate: day(-2),
  },
  {
    id: 'TSK-1039',
    title: 'Publish notification preference guidelines',
    description: 'Document email and in-app options.',
    project: 'Documentation',
    assignee: 'Priya Nair',
    status: 'Done',
    priority: 'Low',
    dueDate: day(-3),
  },
  {
    id: 'TSK-1038',
    title: 'Improve keyboard navigation for task creation',
    description: 'Focus search, escape drawers and support creation shortcuts.',
    project: 'Accessibility',
    assignee: 'Arjun Rao',
    status: 'In progress',
    priority: 'High',
    dueDate: day(1),
  },
  {
    id: 'TSK-1037',
    title: 'Create regression checks for project settings',
    description: 'Cover rename, visibility, and invalid submissions.',
    project: 'Quality',
    assignee: 'Devika Patel',
    status: 'Backlog',
    priority: 'Medium',
    dueDate: day(9),
  },
  {
    id: 'TSK-1036',
    title:
      'Complete mobile navigation audit and resolve bottom safe-area spacing on compact devices',
    description: 'Check safe areas on multiple screen sizes.',
    project: 'Product design',
    assignee: 'Rohan Mehta',
    status: 'To do',
    priority: 'High',
    dueDate: day(5),
  },
  {
    id: 'TSK-1035',
    title: 'Confirm task archiving and restoration versions',
    description: 'Validate optimistic conflict recovery.',
    project: 'Platform',
    assignee: 'Ananya Shah',
    status: 'Blocked',
    priority: 'Medium',
    dueDate: day(-1),
  },
  {
    id: 'TSK-1034',
    title: 'Review accessibility of status badges',
    description: 'Ensure state is conveyed beyond color.',
    project: 'Accessibility',
    assignee: 'Arjun Rao',
    status: 'Done',
    priority: 'Low',
    dueDate: day(0),
  },
  {
    id: 'TSK-1033',
    title: 'Prepare release readiness checklist',
    description: 'Track sign-off and build verification.',
    project: 'Quality',
    assignee: 'Priya Nair',
    status: 'To do',
    priority: 'Urgent',
    dueDate: day(3),
  },
]
const blank = (): Editor => ({
  title: '',
  description: '',
  project: 'Platform',
  assignee: '',
  status: 'To do',
  priority: 'Medium',
  dueDate: day(7),
})
function readTasks(): Task[] {
  try {
    const stored = localStorage.getItem(KEY)
    if (!stored) return seed
    const parsed: unknown = JSON.parse(stored)
    if (
      Array.isArray(parsed) &&
      parsed.every(
        (x) => x && typeof x.id === 'string' && typeof x.title === 'string',
      )
    )
      return parsed as Task[]
  } catch {
    /* Browser storage can be unavailable or contain invalid data. */
  }
  return seed
}
function validate(value: Editor): ErrorField {
  const error: ErrorField = {}
  if (value.title.trim().length < 3)
    error.title = 'Enter at least 3 characters.'
  if (value.title.trim().length > 120)
    error.title = 'Use 120 characters or fewer.'
  if (!value.project.trim()) error.project = 'Choose a project.'
  if (!value.assignee.trim()) error.assignee = 'Enter an assignee.'
  if (
    !/^\d{4}-\d{2}-\d{2}$/.test(value.dueDate) ||
    Number.isNaN(new Date(`${value.dueDate}T12:00:00`).getTime())
  )
    error.dueDate = 'Choose a valid due date.'
  return error
}
const glyph = (
  name:
    | 'search'
    | 'close'
    | 'plus'
    | 'arrow'
    | 'sliders'
    | 'trash'
    | 'check'
    | 'refresh',
) => {
  const paths = {
    search: (
      <>
        <circle cx="11" cy="11" r="7" />
        <path d="m16 16 4 4" />
      </>
    ),
    close: <path d="M5 5 19 19M19 5 5 19" />,
    plus: <path d="M12 5v14M5 12h14" />,
    arrow: <path d="m7 10 5 5 5-5" />,
    sliders: (
      <>
        <path d="M4 7h16M4 17h16" />
        <circle cx="9" cy="7" r="2" fill="#fff" />
        <circle cx="15" cy="17" r="2" fill="#fff" />
      </>
    ),
    trash: (
      <>
        <path d="M4 7h16M9 7V4h6v3M7 7l1 13h8l1-13" />
      </>
    ),
    check: <path d="m5 12 5 5L19 7" />,
    refresh: (
      <path d="M20 7v5h-5M4 17v-5h5M6 8a7 7 0 0 1 12-1l2 5M4 12l2 5a7 7 0 0 0 12-1" />
    ),
  }
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {paths[name]}
    </svg>
  )
}
export function DesignLabPage() {
  const [items, setItems] = useState<Task[]>(readTasks)
  const [search, setSearch] = useState('')
  const [filters, setFilters] = useState<Status[]>([])
  const [priorityFilters, setPriorityFilters] = useState<Priority[]>([])
  const [field, setField] = useState<Field>('dueDate')
  const [direction, setDirection] = useState<'asc' | 'desc'>('asc')
  const [selected, setSelected] = useState<string[]>([])
  const [editor, setEditor] = useState<Editor | null>(null)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [errors, setErrors] = useState<ErrorField>({})
  const [touched, setTouched] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(true)
  const [simulatedError, setSimulatedError] = useState(false)
  const [simulateEmpty, setSimulateEmpty] = useState(false)
  const [demoOpen, setDemoOpen] = useState(false)
  const [filterOpen, setFilterOpen] = useState(false)
  const [undo, setUndo] = useState<Task[] | null>(null)
  const [toastMessage, setToastMessage] = useState('')
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const titleRef = useRef<HTMLInputElement>(null)
  useEffect(() => {
    const id = setTimeout(() => setLoading(false), 600)
    return () => clearTimeout(id)
  }, [])
  useEffect(() => {
    try {
      localStorage.setItem(KEY, JSON.stringify(items))
    } catch {
      /* Private browsing or storage disabled. */
    }
  }, [items])
  const editorOpen = editor !== null
  useEffect(() => {
    if (editorOpen) titleRef.current?.focus()
  }, [editingId, editorOpen])
  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      const target = event.target as HTMLElement | null
      const typing =
        target?.matches('input,textarea,select,[contenteditable="true"]') ??
        false
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        searchRef.current?.focus()
      }
      if (event.key === 'Escape') {
        setEditor(null)
        setDemoOpen(false)
        setFilterOpen(false)
      }
      if (
        event.key.toLowerCase() === 'n' &&
        !typing &&
        !event.ctrlKey &&
        !event.metaKey &&
        !event.altKey
      ) {
        event.preventDefault()
        setEditingId(null)
        setEditor(blank())
        setErrors({})
        setTouched(new Set())
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])
  useEffect(
    () => () => {
      if (timer.current) clearTimeout(timer.current)
    },
    [],
  )
  const visible = useMemo(() => {
    if (simulateEmpty) return []
    const q = search.trim().toLowerCase()
    return items
      .filter(
        (t) =>
          (!q ||
            [
              t.id,
              t.title,
              t.description,
              t.project,
              t.assignee,
              t.status,
              t.priority,
            ].some((x) => x.toLowerCase().includes(q))) &&
          (!filters.length || filters.includes(t.status)) &&
          (!priorityFilters.length || priorityFilters.includes(t.priority)),
      )
      .sort((a, b) => {
        const av =
          field === 'priority' ? priorities.indexOf(a.priority) : a[field]
        const bv =
          field === 'priority' ? priorities.indexOf(b.priority) : b[field]
        const result =
          typeof av === 'number' && typeof bv === 'number'
            ? av - bv
            : String(av).localeCompare(String(bv))
        return direction === 'asc' ? result : -result
      })
  }, [items, search, filters, priorityFilters, field, direction, simulateEmpty])
  const metrics = {
    total: items.length,
    active: items.filter((x) => x.status === 'In progress').length,
    blocked: items.filter((x) => x.status === 'Blocked').length,
    completed: items.filter((x) => x.status === 'Done').length,
  }
  const activeFilters =
    filters.length + priorityFilters.length + (search.trim() ? 1 : 0)
  const selectedVisible = visible.filter((x) => selected.includes(x.id))
  const allChecked =
    visible.length > 0 && selectedVisible.length === visible.length
  const setSorting = (key: Field) => {
    if (field === key) setDirection((d) => (d === 'asc' ? 'desc' : 'asc'))
    else {
      setField(key)
      setDirection('asc')
    }
  }
  const clearFilters = () => {
    setFilters([])
    setPriorityFilters([])
    setSearch('')
    setSimulateEmpty(false)
  }
  const toggleFilter = <T extends Status | Priority>(
    value: T,
    list: T[],
    setter: (items: T[]) => void,
  ) =>
    setter(
      list.includes(value) ? list.filter((x) => x !== value) : [...list, value],
    )
  const openNew = () => {
    setEditingId(null)
    setEditor(blank())
    setErrors({})
    setTouched(new Set())
  }
  const openEdit = (item: Task) => {
    setEditingId(item.id)
    setEditor({ ...item })
    setErrors({})
    setTouched(new Set())
  }
  const onBlur = (name: keyof ErrorField) => {
    setTouched((old) => new Set([...old, name]))
    if (editor) setErrors(validate(editor))
  }
  const update = (key: keyof Editor, value: string) => {
    setEditor((old) => (old ? { ...old, [key]: value } : old))
    if (touched.has(key)) setErrors(validate({ ...editor!, [key]: value }))
  }
  const save = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!editor) return
    const issues = validate(editor)
    setErrors(issues)
    setTouched(new Set(['title', 'project', 'assignee', 'dueDate']))
    if (Object.keys(issues).length) return
    const record: Task = {
      ...editor,
      title: editor.title.trim(),
      project: editor.project.trim(),
      assignee: editor.assignee.trim(),
      id: editingId ?? `TSK-${String(Date.now()).slice(-8)}`,
    }
    setItems((old) =>
      editingId
        ? old.map((x) => (x.id === editingId ? record : x))
        : [record, ...old],
    )
    setEditor(null)
    setEditingId(null)
    setToastMessage(editingId ? 'Task updated' : 'Task created')
    setUndo(null)
  }
  const deleteItems = (ids: string[]) => {
    if (!ids.length) return
    const removed = items.filter((x) => ids.includes(x.id))
    setItems((old) => old.filter((x) => !ids.includes(x.id)))
    setSelected([])
    if (timer.current) clearTimeout(timer.current)
    setUndo(removed)
    setToastMessage(
      `${removed.length} task${removed.length === 1 ? '' : 's'} deleted`,
    )
    timer.current = setTimeout(() => {
      setUndo(null)
      setToastMessage('')
    }, 5000)
  }
  const restore = () => {
    if (!undo) return
    setItems((old) => [
      ...undo.filter((x) => !old.some((y) => y.id === x.id)),
      ...old,
    ])
    setUndo(null)
    setToastMessage('Deletion undone')
    if (timer.current) clearTimeout(timer.current)
  }
  const bulkStatus = (status: Status) => {
    setItems((old) =>
      old.map((x) => (selected.includes(x.id) ? { ...x, status } : x)),
    )
    setSelected([])
    setToastMessage('Selected tasks updated')
    setUndo(null)
  }
  const resetDemo = () => {
    setItems(seed)
    clearFilters()
    setSelected([])
    setSimulatedError(false)
    setLoading(false)
    setToastMessage('Sample records restored')
    setUndo(null)
  }
  const empty = visible.length === 0
  return (
    <div className="lab-page">
      <header className="lab-topbar">
        <div className="lab-brand">
          <span className="lab-brandmark">T</span>
          <strong>Taskflow</strong>
          <span> / UI lab</span>
        </div>
        <a href="/login">Back to sign in</a>
      </header>
      <div className="lab-body">
        <div className="lab-breadcrumb">
          WORKSPACE <span>/</span> PRODUCT DELIVERY
        </div>
        <header className="lab-heading">
          <div>
            <h1>Tasks</h1>
            <p>A clear overview of everything moving through the workspace.</p>
          </div>
          <button className="button primary" type="button" onClick={openNew}>
            {glyph('plus')} New task
          </button>
        </header>
        <div className="lab-summary" aria-label="Task summary">
          {[
            ['Total tasks', metrics.total, 'All tracked work'],
            ['In progress', metrics.active, 'Currently moving'],
            ['Blocked', metrics.blocked, 'Needs a decision'],
            ['Completed', metrics.completed, 'Ready to review'],
          ].map(([name, count, help]) => (
            <section key={name} className="lab-metric">
              <span>{name}</span>
              <strong>{count}</strong>
              <small>{help}</small>
            </section>
          ))}
        </div>
        <section className="lab-workspace">
          <div className="lab-section-title">
            <div>
              <h2>Task directory</h2>
              <p>Search, review and update your team’s work.</p>
            </div>
            <span>{visible.length} shown</span>
          </div>
          <div className="lab-toolbar">
            <label className="lab-search">
              {glyph('search')}
              <span className="sr-only">Search tasks</span>
              <input
                ref={searchRef}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search tasks, projects or people"
                type="search"
              />
              <kbd>⌘ K</kbd>
            </label>
            <div className="lab-filter-holder">
              <button
                className="button secondary"
                type="button"
                onClick={() => setFilterOpen((x) => !x)}
                aria-expanded={filterOpen}
              >
                {glyph('sliders')} Filters
                {activeFilters > 0 && (
                  <span className="lab-count">{activeFilters}</span>
                )}
              </button>
              {filterOpen && (
                <div
                  className="lab-filter-pop"
                  role="group"
                  aria-label="Task filters"
                >
                  <strong>Status</strong>
                  <div className="lab-filter-chips">
                    {statuses.map((x) => (
                      <button
                        key={x}
                        type="button"
                        aria-pressed={filters.includes(x)}
                        onClick={() => toggleFilter(x, filters, setFilters)}
                      >
                        {x}
                      </button>
                    ))}
                  </div>
                  <strong>Priority</strong>
                  <div className="lab-filter-chips">
                    {priorities.map((x) => (
                      <button
                        key={x}
                        type="button"
                        aria-pressed={priorityFilters.includes(x)}
                        onClick={() =>
                          toggleFilter(x, priorityFilters, setPriorityFilters)
                        }
                      >
                        {x}
                      </button>
                    ))}
                  </div>
                  <button
                    type="button"
                    className="lab-clear"
                    onClick={clearFilters}
                  >
                    Clear all filters
                  </button>
                </div>
              )}
            </div>
            <span className="lab-shortcuts">
              Press <kbd>N</kbd> to create
            </span>
          </div>
          {activeFilters > 0 && (
            <div className="lab-active-filters">
              {filters.map((x) => (
                <button
                  type="button"
                  key={x}
                  onClick={() => setFilters((v) => v.filter((y) => y !== x))}
                >
                  {x} ×
                </button>
              ))}
              {priorityFilters.map((x) => (
                <button
                  type="button"
                  key={x}
                  onClick={() =>
                    setPriorityFilters((v) => v.filter((y) => y !== x))
                  }
                >
                  {x} ×
                </button>
              ))}
              <button type="button" onClick={clearFilters}>
                Clear all
              </button>
            </div>
          )}
          {loading ? (
            <div
              className="lab-skeleton"
              role="status"
              aria-label="Loading tasks"
            >
              <div className="lab-skeleton-head" />
              {Array.from({ length: 7 }, (_, i) => (
                <div className="lab-skeleton-row" key={i}>
                  <i />
                  <i />
                  <i />
                  <i />
                </div>
              ))}
            </div>
          ) : simulatedError ? (
            <div className="lab-empty" role="alert">
              <div className="lab-empty-symbol">!</div>
              <h2>We couldn't load tasks</h2>
              <p>
                This is a simulated error state. Try again to return to the
                directory.
              </p>
              <button
                className="button primary"
                type="button"
                onClick={() => setSimulatedError(false)}
              >
                Try again
              </button>
            </div>
          ) : empty ? (
            <div className="lab-empty">
              <svg
                width="130"
                height="96"
                viewBox="0 0 130 96"
                fill="none"
                stroke="#8DABA2"
                strokeWidth="1.7"
                aria-hidden="true"
              >
                <rect x="22" y="15" width="86" height="64" rx="8" />
                <path d="M39 35h49M39 47h49M39 59h29" />
                <circle cx="98" cy="72" r="18" fill="#fff" />
                <path d="m110 84 10 10M92 72h12" />
              </svg>
              <h2>
                {activeFilters || simulateEmpty
                  ? 'No matching records found'
                  : 'Nothing here yet'}
              </h2>
              <p>
                {activeFilters || simulateEmpty
                  ? 'Try clearing filters or search to see more work.'
                  : 'Create your first task to start planning work.'}
              </p>
              <button
                className="button primary"
                type="button"
                onClick={
                  activeFilters || simulateEmpty ? clearFilters : openNew
                }
              >
                {activeFilters || simulateEmpty
                  ? 'Reset filters'
                  : 'Create first task'}
              </button>
            </div>
          ) : (
            <div className="lab-table-wrap">
              <table className="lab-table">
                <thead>
                  <tr>
                    <th>
                      <label>
                        <input
                          type="checkbox"
                          aria-label="Select all shown tasks"
                          checked={allChecked}
                          onChange={(e) =>
                            setSelected(
                              e.target.checked ? visible.map((t) => t.id) : [],
                            )
                          }
                        />
                      </label>
                    </th>
                    {(
                      [
                        ['title', 'Task'],
                        ['project', 'Project'],
                        ['status', 'Status'],
                        ['priority', 'Priority'],
                        ['dueDate', 'Due date'],
                      ] as [Field, string][]
                    ).map(([key, label]) => (
                      <th
                        key={key}
                        aria-sort={
                          field === key
                            ? direction === 'asc'
                              ? 'ascending'
                              : 'descending'
                            : 'none'
                        }
                      >
                        <button
                          type="button"
                          onClick={() => setSorting(key)}
                          aria-label={`Sort by ${label}`}
                        >
                          {label}
                          <span aria-hidden="true">
                            {field === key
                              ? direction === 'asc'
                                ? '↑'
                                : '↓'
                              : '↕'}
                          </span>
                        </button>
                      </th>
                    ))}
                    <th>Assignee</th>
                    <th>
                      <span className="sr-only">Actions</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {visible.map((item) => (
                    <tr
                      key={item.id}
                      className={
                        selected.includes(item.id) ? 'lab-selected' : ''
                      }
                    >
                      <td>
                        <input
                          type="checkbox"
                          aria-label={`Select ${item.title}`}
                          checked={selected.includes(item.id)}
                          onChange={(e) =>
                            setSelected((old) =>
                              e.target.checked
                                ? [...old, item.id]
                                : old.filter((x) => x !== item.id),
                            )
                          }
                        />
                      </td>
                      <td>
                        <button
                          type="button"
                          className="lab-open"
                          onClick={() => openEdit(item)}
                        >
                          <strong>{item.title}</strong>
                          <small>{item.id}</small>
                        </button>
                      </td>
                      <td>{item.project}</td>
                      <td>
                        <select
                          className={`lab-status lab-status-${item.status.toLowerCase().replaceAll(' ', '-')}`}
                          aria-label={`Status for ${item.title}`}
                          value={item.status}
                          onChange={(e) =>
                            setItems((old) =>
                              old.map((x) =>
                                x.id === item.id
                                  ? { ...x, status: e.target.value as Status }
                                  : x,
                              ),
                            )
                          }
                        >
                          {statuses.map((s) => (
                            <option key={s}>{s}</option>
                          ))}
                        </select>
                      </td>
                      <td>
                        <span
                          className={`lab-priority lab-priority-${item.priority.toLowerCase()}`}
                        >
                          {item.priority}
                        </span>
                      </td>
                      <td
                        className={
                          item.dueDate < day(0) && item.status !== 'Done'
                            ? 'lab-overdue'
                            : ''
                        }
                      >
                        {new Intl.DateTimeFormat(undefined, {
                          day: 'numeric',
                          month: 'short',
                          year: 'numeric',
                        }).format(new Date(`${item.dueDate}T12:00:00`))}
                      </td>
                      <td>{item.assignee}</td>
                      <td>
                        <button
                          type="button"
                          className="lab-icon-action"
                          aria-label={`Delete ${item.title}`}
                          title="Delete task"
                          onClick={() => deleteItems([item.id])}
                        >
                          {glyph('trash')}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>
      {selected.length > 0 && (
        <div
          className="lab-bulk-bar"
          role="toolbar"
          aria-label="Selected task actions"
        >
          <strong>{selected.length} selected</strong>
          <label>
            <span className="sr-only">Change selected task status</span>
            <select
              defaultValue=""
              onChange={(e) => {
                if (e.target.value) bulkStatus(e.target.value as Status)
                e.target.value = ''
              }}
            >
              <option value="">Change status…</option>
              {statuses.map((s) => (
                <option key={s}>{s}</option>
              ))}
            </select>
          </label>
          <button
            type="button"
            className="button danger"
            onClick={() => deleteItems(selected)}
          >
            {glyph('trash')} Delete
          </button>
          <button
            type="button"
            className="button secondary"
            onClick={() => setSelected([])}
          >
            Clear
          </button>
        </div>
      )}
      {toastMessage && (
        <div className="lab-toast" role="status">
          <span>{toastMessage}</span>
          {undo && (
            <button type="button" onClick={restore}>
              Undo
            </button>
          )}
          <button
            type="button"
            aria-label="Dismiss notification"
            onClick={() => {
              setToastMessage('')
              setUndo(null)
            }}
          >
            {glyph('close')}
          </button>
        </div>
      )}
      {editor && (
        <div className="lab-drawer-overlay" onMouseDown={() => setEditor(null)}>
          <section
            className="lab-drawer"
            role="dialog"
            aria-modal="true"
            aria-labelledby="lab-drawer-title"
            onMouseDown={(e) => e.stopPropagation()}
          >
            <header>
              <div>
                <span className="lab-overline">TASK DETAILS</span>
                <h2 id="lab-drawer-title">
                  {editingId ? 'Edit task' : 'Create task'}
                </h2>
              </div>
              <button
                type="button"
                className="icon-button"
                onClick={() => setEditor(null)}
                aria-label="Close task editor"
              >
                {glyph('close')}
              </button>
            </header>
            <form onSubmit={save} noValidate>
              <div className="lab-drawer-form">
                <label className="lab-field">
                  Title{' '}
                  <input
                    ref={titleRef}
                    name="title"
                    value={editor.title}
                    maxLength={121}
                    aria-invalid={Boolean(touched.has('title') && errors.title)}
                    onChange={(e) => update('title', e.target.value)}
                    onBlur={() => onBlur('title')}
                    placeholder="What needs to be done?"
                  />
                  {touched.has('title') && errors.title && (
                    <small role="alert">{errors.title}</small>
                  )}
                </label>
                <label className="lab-field">
                  Description{' '}
                  <textarea
                    name="description"
                    rows={4}
                    value={editor.description}
                    onChange={(e) => update('description', e.target.value)}
                    placeholder="Add context and expected outcome"
                  />
                </label>
                <label className="lab-field">
                  Project{' '}
                  <select
                    name="project"
                    value={editor.project}
                    onChange={(e) => update('project', e.target.value)}
                    onBlur={() => onBlur('project')}
                  >
                    <option value="">Select a project</option>
                    {[
                      'Platform',
                      'Product design',
                      'Quality',
                      'Accessibility',
                      'Documentation',
                    ].map((x) => (
                      <option key={x}>{x}</option>
                    ))}
                  </select>
                  {touched.has('project') && errors.project && (
                    <small role="alert">{errors.project}</small>
                  )}
                </label>
                <label className="lab-field">
                  Assignee{' '}
                  <input
                    name="assignee"
                    value={editor.assignee}
                    onChange={(e) => update('assignee', e.target.value)}
                    onBlur={() => onBlur('assignee')}
                    aria-invalid={Boolean(
                      touched.has('assignee') && errors.assignee,
                    )}
                    placeholder="Full name"
                  />
                  {touched.has('assignee') && errors.assignee && (
                    <small role="alert">{errors.assignee}</small>
                  )}
                </label>
                <div className="lab-two-cols">
                  <label className="lab-field">
                    Status{' '}
                    <select
                      name="status"
                      value={editor.status}
                      onChange={(e) => update('status', e.target.value)}
                    >
                      {statuses.map((x) => (
                        <option key={x}>{x}</option>
                      ))}
                    </select>
                  </label>
                  <label className="lab-field">
                    Priority{' '}
                    <select
                      name="priority"
                      value={editor.priority}
                      onChange={(e) => update('priority', e.target.value)}
                    >
                      {priorities.map((x) => (
                        <option key={x}>{x}</option>
                      ))}
                    </select>
                  </label>
                </div>
                <label className="lab-field">
                  Due date{' '}
                  <input
                    name="dueDate"
                    type="date"
                    value={editor.dueDate}
                    onBlur={() => onBlur('dueDate')}
                    onChange={(e) => update('dueDate', e.target.value)}
                    aria-invalid={Boolean(
                      touched.has('dueDate') && errors.dueDate,
                    )}
                  />
                  {touched.has('dueDate') && errors.dueDate && (
                    <small role="alert">{errors.dueDate}</small>
                  )}
                </label>
                <p className="lab-helper">
                  This interactive preview stores changes only in your browser.
                  It does not write to the API.
                </p>
              </div>
              <footer>
                <button
                  type="button"
                  className="button secondary"
                  onClick={() => setEditor(null)}
                >
                  Cancel
                </button>
                <button type="submit" className="button primary">
                  {editingId ? 'Save changes' : 'Create task'}
                </button>
              </footer>
            </form>
          </section>
        </div>
      )}
      <div className="lab-demo">
        <button
          className="lab-demo-trigger"
          type="button"
          aria-expanded={demoOpen}
          onClick={() => setDemoOpen((o) => !o)}
        >
          Demo controls {glyph('arrow')}
        </button>
        {demoOpen && (
          <div className="lab-demo-menu">
            <strong>Test UI states</strong>
            <button
              type="button"
              onClick={() => {
                setLoading(true)
                setDemoOpen(false)
                setTimeout(() => setLoading(false), 600)
              }}
            >
              Show 600ms skeleton
            </button>
            <button
              type="button"
              onClick={() => {
                setSimulateEmpty((v) => !v)
                setDemoOpen(false)
              }}
            >
              {simulateEmpty ? 'Leave empty state' : 'Show empty state'}
            </button>
            <button
              type="button"
              onClick={() => {
                setSimulatedError((v) => !v)
                setDemoOpen(false)
              }}
            >
              {simulatedError ? 'Leave error state' : 'Show error state'}
            </button>
            <button
              type="button"
              onClick={() => {
                resetDemo()
                setDemoOpen(false)
              }}
            >
              {glyph('refresh')} Restore sample data
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
