import { request } from './client'
import type { NewsPage } from '../types/api'

export async function getNews(category: string, q: string, page: number) {
  const params = new URLSearchParams({ page: String(page) })
  if (category) params.set('category', category)
  if (q) params.set('q', q)
  return request<NewsPage>(`/api/news?${params}`)
}
