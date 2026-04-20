export function StatusDot({ status }: { status: 'running' | 'idle' | 'error' }) {
  const colors = {
    running: 'bg-[var(--warning)]',
    idle: 'bg-[var(--success)]',
    error: 'bg-[var(--error)]',
  }

  return (
    <span
      className={`inline-block w-2 h-2 rounded-full ${colors[status]} ${status === 'running' ? 'animate-[pulse_2s_infinite]' : ''}`}
    />
  )
}
