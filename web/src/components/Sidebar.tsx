import { Link, useRouterState } from '@tanstack/react-router'
import { MessageSquare, FolderOpen, Zap, Settings } from 'lucide-react'

export function Sidebar() {
  return (
    <nav className="w-60 shrink-0 border-r border-[var(--border)] bg-[var(--bg-secondary)] flex flex-col p-4 gap-1">
      <div className="px-3 mb-6">
        <span className="font-mono font-semibold text-[15px] text-[var(--text-primary)]">
          agora
        </span>
        <span className="font-mono text-[15px] text-[var(--text-muted)]">
          -agent
        </span>
      </div>

      <SidebarLink to="/" icon={<MessageSquare size={18} />} label="Chat" />
      <SidebarLink to="/agora" icon={<FolderOpen size={18} />} label="Agora" />
      <SidebarLink to="/skills" icon={<Zap size={18} />} label="Skills" />

      <div className="flex-1" />

      <SidebarLink to="/settings" icon={<Settings size={18} />} label="Settings" />
    </nav>
  )
}

function SidebarLink({ to, icon, label }: { to: string; icon: React.ReactNode; label: string }) {
  const router = useRouterState()
  const isActive = router.location.pathname === to

  return (
    <Link
      to={to}
      className={`
        flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium no-underline
        transition-all duration-150
        ${isActive
          ? 'bg-[var(--accent-soft)] text-[var(--accent)] border-l-2 border-l-[var(--accent)]'
          : 'text-[var(--text-secondary)] border-l-2 border-l-transparent hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]'
        }
      `}
    >
      {icon}
      {label}
    </Link>
  )
}
