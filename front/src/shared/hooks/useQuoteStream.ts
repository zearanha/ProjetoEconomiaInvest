import { useEffect, useState } from 'react'
import type { Quote } from '../types/api'

export type Market = 'crypto' | 'fx' | 'stocks'
export type StreamStatus = 'connecting' | 'connected' | 'reconnecting'

type QuoteMessage = { type: 'quotes'; market: Market; data: Quote[] }
type QuoteMap = Partial<Record<Market, Quote[]>>

export function useQuoteStream() {
  const [quotes, setQuotes] = useState<QuoteMap>({})
  const [status, setStatus] = useState<StreamStatus>('connecting')

  useEffect(() => {
    let socket: WebSocket | null = null
    let reconnectTimer: number | undefined
    let stopped = false
    let attempts = 0

    function connect() {
      if (stopped) return
      setStatus(attempts === 0 ? 'connecting' : 'reconnecting')
      const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      socket = new WebSocket(`${scheme}//${window.location.host}/ws`)

      socket.onopen = () => {
        attempts = 0
        setStatus('connected')
      }
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as QuoteMessage
          if (message.type !== 'quotes' || !['crypto', 'fx', 'stocks'].includes(message.market) || !Array.isArray(message.data)) return
          setQuotes((current) => ({ ...current, [message.market]: message.data }))
        } catch {
          // Ignore malformed frames and keep the connection available.
        }
      }
      socket.onerror = () => socket?.close()
      socket.onclose = () => {
        if (stopped) return
        setStatus('reconnecting')
        const delay = Math.min(1_000 * 2 ** attempts, 30_000)
        attempts = Math.min(attempts + 1, 5)
        reconnectTimer = window.setTimeout(connect, delay)
      }
    }

    connect()
    return () => {
      stopped = true
      if (reconnectTimer !== undefined) window.clearTimeout(reconnectTimer)
      if (socket) {
        socket.onclose = null
        socket.close()
      }
    }
  }, [])

  return { quotes, status }
}
