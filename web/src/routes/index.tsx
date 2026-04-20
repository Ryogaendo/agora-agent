import { createFileRoute } from '@tanstack/react-router'
import { useState, useRef, useEffect } from 'react'
import { Send, Loader2 } from 'lucide-react'
import { ToolBadge } from '../components/ToolBadge'
import { StatusDot } from '../components/StatusDot'

export const Route = createFileRoute('/')({ component: Chat })

type Message =
  | { role: 'user'; text: string }
  | { role: 'agent'; text: string }
  | { role: 'tool'; name: string }
  | { role: 'status'; status: 'running' | 'idle' | 'error'; text?: string }

function Chat() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isRunning, setIsRunning] = useState(false)
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!input.trim() || isRunning) return

    const prompt = input.trim()
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', text: prompt }])
    setIsRunning(true)
    setMessages((prev) => [...prev, { role: 'status', status: 'running', text: 'Starting session...' }])

    try {
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt }),
      })

      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      if (!res.body) throw new Error('No response body')

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let agentText = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        for (const line of chunk.split('\n')) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6)
          if (data === '[DONE]') continue

          try {
            const event = JSON.parse(data)
            switch (event.type) {
              case 'message':
                agentText += event.text
                setMessages((prev) => {
                  const rest = prev.filter((m, i) => !(m.role === 'agent' && i === prev.length - 1))
                  return [...rest, { role: 'agent', text: agentText }]
                })
                break
              case 'tool_use':
                setMessages((prev) => [...prev, { role: 'tool', name: event.text }])
                break
              case 'done':
                setMessages((prev) => prev.filter((m) => m.role !== 'status'))
                break
              case 'error':
                setMessages((prev) => [...prev, { role: 'status', status: 'error', text: event.text }])
                break
            }
          } catch { /* skip */ }
        }
      }
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        { role: 'status', status: 'error', text: err instanceof Error ? err.message : 'Unknown error' },
      ])
    } finally {
      setIsRunning(false)
    }
  }

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <header className="flex items-center gap-2 px-6 py-3 border-b border-[var(--border)]">
        <span className="font-semibold text-base">Chat</span>
        {isRunning && (
          <span className="flex items-center gap-1 text-[13px] text-[var(--text-secondary)]">
            <StatusDot status="running" /> Running
          </span>
        )}
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-[720px] mx-auto flex flex-col gap-4">
          {messages.length === 0 && <EmptyState />}
          {messages.map((msg, i) => (
            <MessageBubble key={i} message={msg} />
          ))}
          <div ref={endRef} />
        </div>
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="px-6 py-4 border-t border-[var(--border)]">
        <div className="max-w-[720px] mx-auto flex gap-2">
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask agora-agent..."
            disabled={isRunning}
            className="flex-1 bg-[var(--bg-tertiary)] text-[var(--text-primary)] border border-[var(--border)] rounded-md px-3 py-2.5 text-sm font-sans outline-none transition-all duration-150 focus:border-[var(--accent-border)] focus:shadow-[0_0_0_2px_var(--accent-soft)] placeholder:text-[var(--text-muted)]"
          />
          <button
            type="submit"
            disabled={isRunning || !input.trim()}
            className="bg-[var(--accent)] text-[var(--text-inverse)] border-none rounded-md px-4 py-2.5 text-sm font-medium font-sans flex items-center gap-1 cursor-pointer transition-all duration-150 hover:bg-[var(--accent-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isRunning ? <Loader2 size={18} className="animate-spin" /> : <Send size={18} />}
          </button>
        </div>
      </form>
    </div>
  )
}

function MessageBubble({ message }: { message: Message }) {
  if (message.role === 'tool') {
    return (
      <div className="py-1">
        <ToolBadge name={message.name} />
      </div>
    )
  }

  if (message.role === 'status') {
    return (
      <div className="flex items-center gap-2 text-[13px] text-[var(--text-secondary)] py-2">
        <StatusDot status={message.status} />
        {message.text}
      </div>
    )
  }

  const isUser = message.role === 'user'

  return (
    <div
      className={`rounded-lg p-3.5 border ${
        isUser
          ? 'bg-[var(--accent-soft)] border-[var(--accent-border)]'
          : 'bg-[var(--bg-secondary)] border-[var(--border)]'
      }`}
    >
      <div
        className={`text-[11px] font-medium font-mono uppercase tracking-wider mb-1 ${
          isUser ? 'text-[var(--accent)]' : 'text-[var(--text-muted)]'
        }`}
      >
        {isUser ? 'You' : 'Agent'}
      </div>
      <div
        className={`whitespace-pre-wrap break-words ${
          isUser ? 'font-sans text-sm leading-relaxed' : 'font-mono text-[13px] leading-[1.55]'
        } text-[var(--text-primary)]`}
      >
        {message.text}
      </div>
    </div>
  )
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center flex-1 gap-4 py-20 text-[var(--text-muted)]">
      <span className="text-5xl opacity-30">&#x1F3DB;</span>
      <span className="font-mono text-lg font-semibold text-[var(--text-secondary)]">agora-agent</span>
      <span className="text-[13px] text-center max-w-[400px] leading-relaxed">
        Cross-repository analysis and shared knowledge management powered by Claude Managed Agents.
      </span>
      <div className="flex gap-2 flex-wrap justify-center mt-2">
        {[
          'Analyze auth flows across repos',
          "Summarize this week's PRs",
          'Compare monorepo structures',
        ].map((s) => (
          <span
            key={s}
            className="text-xs font-mono px-2 py-1 bg-[var(--bg-tertiary)] border border-[var(--border)] rounded text-[var(--text-secondary)] cursor-pointer hover:border-[var(--border-active)] transition-colors"
          >
            {s}
          </span>
        ))}
      </div>
    </div>
  )
}
