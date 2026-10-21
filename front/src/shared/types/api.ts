export type Quote = {
  symbol: string; name: string; market: string; currency: string
  price: number; change: number; changePct: number; volume?: number
  high?: number; low?: number; updatedAt: string
}
export type QuotePoint = { time: string; price: number }
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
