import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/skills')({ component: SkillsPage })

function SkillsPage() {
  return (
    <div className="p-6 max-w-[960px] mx-auto">
      <h1 className="text-2xl font-semibold mb-6">Skills</h1>
      <p className="text-[13px] text-[var(--text-secondary)]">
        Organon skill management. Coming soon.
      </p>
    </div>
  )
}
