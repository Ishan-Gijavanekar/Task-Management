import './notifications.css'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import type { NotificationPreferences } from '../../api/types'

const preferenceOptions = [
  {
    key: 'mentions',
    label: 'Mentions',
    description: 'Notify me when someone mentions me in a task comment.',
  },
  {
    key: 'taskUpdates',
    label: 'Task updates',
    description: 'Notify me when tasks I follow are changed.',
  },
  {
    key: 'emailEnabled',
    label: 'Email notifications',
    description: 'Also send enabled notifications to my email address.',
  },
] as const

export function NotificationsPage() {
  const queryClient = useQueryClient()
  const notificationKey = ['notifications']
  const preferenceKey = ['notification-preferences']
  const notifications = useQuery({
    queryKey: notificationKey,
    queryFn: ({ signal }) => api.listNotifications(false, signal),
  })
  const preferences = useQuery({
    queryKey: preferenceKey,
    queryFn: ({ signal }) => api.getNotificationPreferences(signal),
  })
  const markRead = useMutation({
    mutationFn: (id?: string) => api.markNotificationsRead(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: notificationKey }),
  })
  const savePreferences = useMutation({
    mutationFn: api.updateNotificationPreferences,
    onMutate: async (next) => {
      await queryClient.cancelQueries({ queryKey: preferenceKey })
      const previous =
        queryClient.getQueryData<NotificationPreferences>(preferenceKey)
      queryClient.setQueryData(preferenceKey, next)
      return previous
    },
    onError: (_, __, previous) =>
      queryClient.setQueryData(preferenceKey, previous),
    onSuccess: (updated) => queryClient.setQueryData(preferenceKey, updated),
  })
  const items = notifications.data?.notifications ?? []
  const unreadCount = items.filter((item) => !item.readAt).length

  return (
    <div className="page notifications-page">
      <header className="page-header notifications-header">
        <div>
          <p className="eyebrow">Updates</p>
          <h1>Notifications</h1>
          <p>Mentions and task signals that need your attention.</p>
        </div>
        <div className="notification-actions">
          <span>{unreadCount ? `${unreadCount} unread` : 'All caught up'}</span>
          <button
            className="button secondary"
            type="button"
            disabled={!unreadCount || markRead.isPending}
            onClick={() => markRead.mutate(undefined)}
          >
            {markRead.isPending ? 'Marking…' : 'Mark all read'}
          </button>
        </div>
      </header>
      <div className="notifications-layout">
        <section
          className="panel notification-inbox"
          aria-labelledby="inbox-title"
        >
          <div className="panel-heading notification-section-heading">
            <div>
              <h2 id="inbox-title">Inbox</h2>
              <p>Recent activity across your workspace.</p>
            </div>
          </div>
          {markRead.isError && (
            <div className="notice error" role="alert">
              {errorMessage(markRead.error)}
            </div>
          )}
          {notifications.isLoading ? (
            <div className="notification-loading">
              <span className="spinner" />
              <span>Loading notifications…</span>
            </div>
          ) : notifications.isError ? (
            <div className="notice error" role="alert">
              {errorMessage(notifications.error)}
            </div>
          ) : items.length ? (
            <div className="notification-list">
              {items.map((item) => (
                <Link
                  className={item.readAt ? '' : 'unread'}
                  key={item.id}
                  to={`/tasks?task=${item.taskId}`}
                  onClick={() => !item.readAt && markRead.mutate(item.id)}
                >
                  <span className="notification-icon" aria-hidden="true">
                    {item.type.toLocaleLowerCase().includes('mention')
                      ? '@'
                      : '✓'}
                  </span>
                  <span className="notification-copy">
                    <strong>{item.message}</strong>
                    <span>
                      {item.type.replaceAll('_', ' ')}
                      <span aria-hidden="true"> · </span>
                      <time dateTime={item.createdAt}>
                        {new Intl.DateTimeFormat(undefined, {
                          dateStyle: 'medium',
                          timeStyle: 'short',
                        }).format(new Date(item.createdAt))}
                      </time>
                    </span>
                  </span>
                  {!item.readAt && (
                    <span className="unread-dot">
                      <span className="sr-only">Unread</span>
                    </span>
                  )}
                </Link>
              ))}
            </div>
          ) : (
            <div className="notification-empty">
              <span aria-hidden="true">✓</span>
              <h2>You’re all caught up</h2>
              <p>New mentions and task updates will appear here.</p>
            </div>
          )}
        </section>
        <section
          className="panel preference-panel"
          aria-labelledby="preferences-title"
        >
          <div className="notification-section-heading">
            <h2 id="preferences-title">Preferences</h2>
            <p>Choose how you want to stay informed.</p>
          </div>
          {preferences.isLoading ? (
            <div className="notification-loading">
              <span className="spinner" />
              <span>Loading preferences…</span>
            </div>
          ) : preferences.isError ? (
            <div className="notice error" role="alert">
              {errorMessage(preferences.error)}
            </div>
          ) : preferences.data ? (
            <div className="preference-list">
              {preferenceOptions.map(({ key, label, description }) => (
                <label key={key}>
                  <span>
                    <strong>{label}</strong>
                    <small>{description}</small>
                  </span>
                  <input
                    type="checkbox"
                    role="switch"
                    checked={preferences.data[key]}
                    disabled={savePreferences.isPending}
                    onChange={(event) =>
                      savePreferences.mutate({
                        ...preferences.data,
                        [key]: event.target.checked,
                      })
                    }
                  />
                </label>
              ))}
            </div>
          ) : null}
          {savePreferences.isError && (
            <div className="notice error" role="alert">
              {errorMessage(savePreferences.error)}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
