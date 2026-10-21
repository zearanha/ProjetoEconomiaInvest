import { request } from './client'
import type { Quote, QuotePoint } from '../types/api'

export async function getQuotes(market: string) {
  return request<{ market: string; data: Quote[] }>(`/api/quotes?market=${encodeURIComponent(market)}`)
}
export async function getQuoteHistory(symbol: string, range: string) {
  const params = new URLSearchParams({ symbol, range })
  return request<{ symbol: string; data: QuotePoint[] }>(`/api/quotes/history?${params}`)
}
