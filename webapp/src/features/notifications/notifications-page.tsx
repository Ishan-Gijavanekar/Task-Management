import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../../api/client'
import { LoadingScreen } from '../../ui/states'

export function NotificationsPage() {
  const queryClient = useQueryClient()
  const notifications = useQuery({
    queryKey: ['notifications'],
    queryFn: ({ signal }) => api.listNotifications(false, signal),
  })
  const preferences = useQuery({
    queryKey: ['notification-preferences'],
    queryFn: ({ signal }) => api.getNotificationPreferences(signal),
  })
  const markRead = useMutation({
    mutationFn: (id?: string) => api.markNotificationsRead(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  })
  const savePreferences = useMutation({
    mutationFn: api.updateNotificationPreferences,
    onSuccess: (updated) =>
      queryClient.setQueryData(['notification-preferences'], updated),
  })

  return (
    <div className="page narrow-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Updates</p>
          <h1>Notifications</h1>
          <p>Mentions and task signals that need your attention.</p>
        </div>
        <button
          className="button secondary"
          type="button"
          onClick={() => markRead.mutate(undefined)}
        >
          Mark all read
        </button>
      </header>
      {notifications.isLoading ? (
        <LoadingScreen label="Loading notifications" />
      ) : (
        <section className="panel notification-list">
          {notifications.data?.notifications.length ? (
            notifications.data.notifications.map((item) => (
              <Link
                className={item.readAt ? '' : 'unread'}
                key={item.id}
                to={`/tasks?task=${item.taskId}`}
                onClick={() => !item.readAt && markRead.mutate(item.id)}
              >
                <span>{item.message}</span>
                <time dateTime={item.createdAt}>
                  {new Intl.DateTimeFormat(undefined, {
                    dateStyle: 'medium',
                    timeStyle: 'short',
                  }).format(new Date(item.createdAt))}
                </time>
              </Link>
            ))
          ) : (
            <p className="muted">You’re all caught up.</p>
          )}
        </section>
      )}
      {preferences.data && (
        <section className="panel preference-panel">
          <h2>Preferences</h2>
          {(['mentions', 'taskUpdates', 'emailEnabled'] as const).map((key) => (
            <label key={key}>
              <input
                type="checkbox"
                checked={preferences.data[key]}
                onChange={(event) =>
                  savePreferences.mutate({
                    ...preferences.data,
                    [key]: event.target.checked,
                  })
                }
              />
              <span>{key.replace(/([A-Z])/g, ' $1')}</span>
            </label>
          ))}
        </section>
      )}
    </div>
  )
}
