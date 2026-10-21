import { request } from './client'
import type { Rate, Simulation } from '../types/api'

export async function getRates() { return request<{ data: Rate[] }>('/api/fixed-income/rates') }
export async function simulate(input: { amount: number; termDays: number; cdbPercentCDI: number; lciPercentCDI: number; lcaPercentCDI: number }) {
  return request<Simulation>('/api/fixed-income/simulate', { method: 'POST', body: JSON.stringify(input) })
}
