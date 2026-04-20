import { Terminal } from 'lucide-react'

export function ToolBadge({ name }: { name: string }) {
  return (
    <span className="inline-flex items-center gap-1 bg-[var(--bg-tertiary)] border border-[var(--border)] rounded px-2 py-1 font-mono text-xs text-[var(--text-secondary)]">
      <Terminal size={14} className="text-[var(--info)]" />
      {name}
    </span>
  )
}
