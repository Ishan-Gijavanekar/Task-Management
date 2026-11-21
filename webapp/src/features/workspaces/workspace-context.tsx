/* eslint-disable react-refresh/only-export-components -- Provider and hook form one workspace module. */
import { useQuery } from '@tanstack/react-query'
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { api } from '../../api/client'
import type { Workspace } from '../../api/types'
import { useAuth } from '../auth/auth-context'

export interface WorkspaceReference {
  id: string
  name: string
}

interface WorkspaceContextValue {
  workspace: Workspace | null
  workspaces: WorkspaceReference[]
  loading: boolean
  select(id: string): void
  remember(workspace: Workspace): void
  forget(id: string): void
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null)

function readReferences(key: string) {
  try {
    const value = JSON.parse(localStorage.getItem(key) ?? '[]') as unknown
    if (!Array.isArray(value)) return []
    return value.filter(
      (item): item is WorkspaceReference =>
        typeof item === 'object' &&
        item !== null &&
        typeof (item as WorkspaceReference).id === 'string' &&
        typeof (item as WorkspaceReference).name === 'string',
    )
  } catch {
    return []
  }
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const storageKey = `taskflow.workspaces.${user?.id ?? 'anonymous'}`
  const selectedKey = `${storageKey}.selected`
  const [workspaces, setWorkspaces] = useState<WorkspaceReference[]>(() =>
    readReferences(storageKey),
  )
  const [selectedId, setSelectedId] = useState(
    () => localStorage.getItem(selectedKey) ?? workspaces[0]?.id ?? '',
  )

  const persist = useCallback(
    (items: WorkspaceReference[]) => {
      setWorkspaces(items)
      localStorage.setItem(storageKey, JSON.stringify(items))
    },
    [storageKey],
  )

  const selected = useQuery({
    queryKey: ['workspace', selectedId],
    queryFn: ({ signal }) => api.getWorkspace(selectedId, signal),
    enabled: Boolean(selectedId),
  })

  const value = useMemo<WorkspaceContextValue>(
    () => ({
      workspace: selected.data ?? null,
      workspaces,
      loading: selected.isLoading,
      select: (id) => {
        setSelectedId(id)
        localStorage.setItem(selectedKey, id)
      },
      remember: (workspace) => {
        persist([
          { id: workspace.id, name: workspace.name },
          ...workspaces.filter((item) => item.id !== workspace.id),
        ])
        setSelectedId(workspace.id)
        localStorage.setItem(selectedKey, workspace.id)
      },
      forget: (id) => {
        const next = workspaces.filter((item) => item.id !== id)
        persist(next)
        if (id === selectedId) {
          const nextId = next[0]?.id ?? ''
          setSelectedId(nextId)
          if (nextId) localStorage.setItem(selectedKey, nextId)
          else localStorage.removeItem(selectedKey)
        }
      },
    }),
    [
      persist,
      selected.data,
      selected.isLoading,
      selectedId,
      selectedKey,
      workspaces,
    ],
  )

  return (
    <WorkspaceContext.Provider value={value}>
      {children}
    </WorkspaceContext.Provider>
  )
}

export function useWorkspace() {
  const value = useContext(WorkspaceContext)
  if (!value)
    throw new Error('useWorkspace must be used inside WorkspaceProvider')
  return value
}
