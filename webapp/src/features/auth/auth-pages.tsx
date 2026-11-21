import { type FormEvent, useEffect, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../../api/client'
import { errorMessage } from '../../api/errors'
import { useAuth } from './auth-context'
import { safeReturnUrl } from './safe-return'

function Field({
  id,
  label,
  type = 'text',
  autoComplete,
  minLength,
}: {
  id: string
  label: string
  type?: string
  autoComplete?: string
  minLength?: number
}) {
  return (
    <label className="field" htmlFor={id}>
      <span>{label}</span>
      <input
        id={id}
        name={id}
        type={type}
        autoComplete={autoComplete}
        minLength={minLength}
        required
      />
    </label>
  )
}

function SubmitButton({
  pending,
  children,
}: {
  pending: boolean
  children: string
}) {
  return (
    <button className="button primary full" type="submit" disabled={pending}>
      {pending ? 'Working…' : children}
    </button>
  )
}

export function LoginPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const [search] = useSearchParams()
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setPending(true)
    const data = new FormData(event.currentTarget)
    try {
      await auth.login({
        email: String(data.get('email')),
        password: String(data.get('password')),
      })
      navigate(safeReturnUrl(search.get('returnTo')), { replace: true })
    } catch (cause) {
      setError(errorMessage(cause))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="auth-card" aria-labelledby="login-title">
      <p className="eyebrow">Welcome back</p>
      <h1 id="login-title">Sign in to Taskflow</h1>
      <p className="auth-intro">Pick up where your team left off.</p>
      {auth.sessionExpired && (
        <div className="notice warning" role="alert">
          Your session expired. Sign in to continue.
        </div>
      )}
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      <form onSubmit={submit}>
        <Field
          id="email"
          label="Work email"
          type="email"
          autoComplete="email"
        />
        <div className="field-heading">
          <label htmlFor="password">Password</label>
          <Link to="/forgot-password">Forgot password?</Link>
        </div>
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
        />
        <SubmitButton pending={pending}>Sign in</SubmitButton>
      </form>
      <p className="auth-switch">
        New to Taskflow? <Link to="/register">Create an account</Link>
      </p>
    </section>
  )
}

export function RegisterPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setPending(true)
    const data = new FormData(event.currentTarget)
    try {
      await auth.register({
        name: String(data.get('name')),
        email: String(data.get('email')),
        password: String(data.get('password')),
      })
      navigate('/', { replace: true })
    } catch (cause) {
      setError(errorMessage(cause))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="auth-card" aria-labelledby="register-title">
      <p className="eyebrow">Start fresh</p>
      <h1 id="register-title">Create your account</h1>
      <p className="auth-intro">Set up your calm corner for focused work.</p>
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      <form onSubmit={submit}>
        <Field id="name" label="Full name" autoComplete="name" />
        <Field
          id="email"
          label="Work email"
          type="email"
          autoComplete="email"
        />
        <Field
          id="password"
          label="Password"
          type="password"
          autoComplete="new-password"
          minLength={8}
        />
        <p className="field-hint">
          Use at least 8 characters with a mix of letters and numbers.
        </p>
        <SubmitButton pending={pending}>Create account</SubmitButton>
      </form>
      <p className="auth-switch">
        Already have an account? <Link to="/login">Sign in</Link>
      </p>
    </section>
  )
}

export function ForgotPasswordPage() {
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
      await api.forgotPassword(
        String(new FormData(event.currentTarget).get('email')),
      )
      setSent(true)
    } catch (cause) {
      setError(errorMessage(cause))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="auth-card" aria-labelledby="forgot-title">
      <p className="eyebrow">Account recovery</p>
      <h1 id="forgot-title">Reset your password</h1>
      <p className="auth-intro">
        We’ll send instructions if an account matches that email.
      </p>
      {sent ? (
        <div className="success-panel" role="status">
          <span aria-hidden="true">✓</span>
          <h2>Check your inbox</h2>
          <p>Follow the secure link in the email to choose a new password.</p>
          <Link className="button secondary" to="/login">
            Back to sign in
          </Link>
        </div>
      ) : (
        <form onSubmit={submit}>
          {error && (
            <div className="notice error" role="alert">
              {error}
            </div>
          )}
          <Field
            id="email"
            label="Work email"
            type="email"
            autoComplete="email"
          />
          <SubmitButton pending={pending}>Send reset link</SubmitButton>
          <Link className="text-link centered" to="/login">
            Back to sign in
          </Link>
        </form>
      )}
    </section>
  )
}

export function ResetPasswordPage() {
  const [search] = useSearchParams()
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)
  const token = search.get('token') ?? ''

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const password = String(data.get('password'))
    if (password !== String(data.get('confirmPassword'))) {
      setError('Passwords do not match.')
      return
    }
    setPending(true)
    setError('')
    try {
      await api.resetPassword(token, password)
      navigate('/login?reset=success', { replace: true })
    } catch (cause) {
      setError(errorMessage(cause))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="auth-card" aria-labelledby="reset-title">
      <p className="eyebrow">Almost there</p>
      <h1 id="reset-title">Choose a new password</h1>
      {!token ? (
        <div className="notice error" role="alert">
          This reset link is missing its secure token.
        </div>
      ) : (
        <form onSubmit={submit}>
          {error && (
            <div className="notice error" role="alert">
              {error}
            </div>
          )}
          <Field
            id="password"
            label="New password"
            type="password"
            autoComplete="new-password"
            minLength={8}
          />
          <Field
            id="confirmPassword"
            label="Confirm password"
            type="password"
            autoComplete="new-password"
            minLength={8}
          />
          <SubmitButton pending={pending}>Update password</SubmitButton>
        </form>
      )}
    </section>
  )
}

export function VerifyEmailPage() {
  const [search] = useSearchParams()
  const token = search.get('token')
  const [state, setState] = useState<'working' | 'success' | 'error'>(
    token ? 'working' : 'error',
  )
  const [message, setMessage] = useState(
    token
      ? 'Verifying your email…'
      : 'This verification link is missing its secure token.',
  )

  useEffect(() => {
    if (!token) return
    const controller = new AbortController()
    api
      .verifyEmail(token, controller.signal)
      .then(() => {
        setState('success')
        setMessage('Your email is verified. You can now sign in.')
      })
      .catch((cause: unknown) => {
        if (!controller.signal.aborted) {
          setState('error')
          setMessage(errorMessage(cause))
        }
      })
    return () => controller.abort()
  }, [token])

  return (
    <section className="auth-card" aria-labelledby="verify-title">
      <p className="eyebrow">Email verification</p>
      <h1 id="verify-title">
        {state === 'success' ? 'You’re verified' : 'Confirming your email'}
      </h1>
      <div className={`status-panel ${state}`} role="status" aria-live="polite">
        <span aria-hidden="true">
          {state === 'working' ? '…' : state === 'success' ? '✓' : '!'}
        </span>
        <p>{message}</p>
      </div>
      {state !== 'working' && (
        <Link className="button primary full" to="/login">
          Continue to sign in
        </Link>
      )}
    </section>
  )
}

export function ResendVerificationPage() {
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
      const result = await api.resendVerification(
        String(new FormData(event.currentTarget).get('email')),
      )
      setMessage(result.message)
    } catch (cause) {
      setError(errorMessage(cause))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="auth-card" aria-labelledby="resend-title">
      <p className="eyebrow">Email verification</p>
      <h1 id="resend-title">Send a new link</h1>
      <p className="auth-intro">
        Verification links expire for your protection.
      </p>
      {message && (
        <div className="notice success" role="status">
          {message}
        </div>
      )}
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      <form onSubmit={submit}>
        <Field
          id="email"
          label="Work email"
          type="email"
          autoComplete="email"
        />
        <SubmitButton pending={pending}>Resend verification</SubmitButton>
      </form>
    </section>
  )
}
