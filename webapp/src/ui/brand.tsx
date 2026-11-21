import { Link } from 'react-router-dom'

export function Brand() {
  return (
    <Link className="brand" to="/" aria-label="Taskflow home">
      <span className="brand-mark" aria-hidden="true">
        T
      </span>
      <span>Taskflow</span>
    </Link>
  )
}
