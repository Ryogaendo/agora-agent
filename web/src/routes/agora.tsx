import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/agora')({ component: AgoraPage })

function AgoraPage() {
  return (
    <div className="p-6 max-w-[960px] mx-auto">
      <h1 className="text-2xl font-semibold mb-6">Agora</h1>
      <p className="text-[13px] text-[var(--text-secondary)]">
        Shared knowledge store. Coming soon.
      </p>
    </div>
  )
}
