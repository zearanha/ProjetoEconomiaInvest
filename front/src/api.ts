export type Quote = {
  symbol: string; name: string; market: string; currency: string
  price: number; change: number; changePct: number; volume?: number
  high?: number; low?: number; updatedAt: string
}
export type Rate = { indicator: string; value: number; unit: string; date: string; source: string }
export type NewsItem = {
  id: string; title: string; source: string; publishedAt: string
  summary: string; url: string; category: string
}
export type NewsPage = { data: NewsItem[]; page: number; pageSize: number; total: number }
export type SimulationResult = {
  product: string; grossProfit: number; tax: number; netProfit: number
  netAmount: number; incomeTaxPercent: number
}
export type Simulation = {
  cdiAnnualPercent: number; estimatedBusinessDays: number
  results: SimulationResult[]; assumption: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, headers: { 'Content-Type': 'application/json', ...init?.headers } })
  if (!response.ok) {
    let message = `Falha na API (${response.status})`
    try { message = (await response.json()).error ?? message } catch { /* resposta sem JSON */ }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

export async function getQuotes(market: string) {
  return request<{ market: string; data: Quote[] }>(`/api/quotes?market=${encodeURIComponent(market)}`)
}
export async function getRates() { return request<{ data: Rate[] }>('/api/fixed-income/rates') }
export async function getNews(category: string, q: string, page: number) {
  const params = new URLSearchParams({ page: String(page) })
  if (category) params.set('category', category)
  if (q) params.set('q', q)
  return request<NewsPage>(`/api/news?${params}`)
}
export async function simulate(input: { amount: number; termDays: number; cdbPercentCDI: number; lciPercentCDI: number; lcaPercentCDI: number }) {
  return request<Simulation>('/api/fixed-income/simulate', { method: 'POST', body: JSON.stringify(input) })
}
