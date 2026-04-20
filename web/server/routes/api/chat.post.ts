import { defineEventHandler, readBody, setResponseHeaders } from 'vinxi/http'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const apiKey = process.env.ANTHROPIC_API_KEY

  if (!apiKey) {
    return new Response(JSON.stringify({ error: 'ANTHROPIC_API_KEY not set' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  // Read agent config
  const configPath = `${process.env.HOME}/.agora-agent.json`
  let config: { agent_id?: string; environment_id?: string } = {}
  try {
    const { readFile } = await import('node:fs/promises')
    const raw = await readFile(configPath, 'utf-8')
    config = JSON.parse(raw)
  } catch {
    return new Response(JSON.stringify({ error: 'Run `agora-agent setup` first' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  if (!config.agent_id || !config.environment_id) {
    return new Response(JSON.stringify({ error: 'Agent not configured' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const headers = {
    'x-api-key': apiKey,
    'anthropic-version': '2023-06-01',
    'anthropic-beta': 'managed-agents-2026-04-01',
    'content-type': 'application/json',
  }

  // Create session
  const sessionRes = await fetch('https://api.anthropic.com/v1/sessions', {
    method: 'POST',
    headers,
    body: JSON.stringify({
      agent: config.agent_id,
      environment_id: config.environment_id,
    }),
  })

  if (!sessionRes.ok) {
    const err = await sessionRes.text()
    return new Response(JSON.stringify({ error: `Session failed: ${err}` }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const session = (await sessionRes.json()) as { id: string }

  // Send message
  await fetch(`https://api.anthropic.com/v1/sessions/${session.id}/events`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      events: [
        {
          type: 'user.message',
          content: [{ type: 'text', text: body.prompt }],
        },
      ],
    }),
  })

  // Stream
  const streamRes = await fetch(
    `https://api.anthropic.com/v1/sessions/${session.id}/stream`,
    { headers: { ...headers, Accept: 'text/event-stream' } },
  )

  if (!streamRes.ok || !streamRes.body) {
    return new Response(JSON.stringify({ error: 'Stream failed' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const { readable, writable } = new TransformStream()
  const writer = writable.getWriter()
  const encoder = new TextEncoder()

  ;(async () => {
    const reader = streamRes.body!.getReader()
    const decoder = new TextDecoder()

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        for (const line of chunk.split('\n')) {
          if (!line.startsWith('data: ')) continue
          try {
            const ev = JSON.parse(line.slice(6))
            let out: { type: string; text?: string } | null = null

            switch (ev.type) {
              case 'agent.message':
                if (ev.content) {
                  const text = ev.content
                    .filter((b: { type: string }) => b.type === 'text')
                    .map((b: { text: string }) => b.text)
                    .join('')
                  if (text) out = { type: 'message', text }
                }
                break
              case 'agent.tool_use':
                out = { type: 'tool_use', text: ev.name }
                break
              case 'session.status_idle':
                out = { type: 'done' }
                break
              case 'session.status_terminated':
                out = { type: 'error', text: 'Session terminated' }
                break
            }

            if (out) {
              await writer.write(encoder.encode(`data: ${JSON.stringify(out)}\n\n`))
            }
          } catch {
            /* skip */
          }
        }
      }
    } finally {
      await writer.write(encoder.encode('data: [DONE]\n\n'))
      await writer.close()
    }
  })()

  setResponseHeaders(event, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
  })

  return readable
})
