import type { QuotePoint } from '../../shared/types/api'

export function QuoteChart({ points }: { points: QuotePoint[] }) {
  if (!points.length) return <div className="chart-empty">O histórico começa a ser formado conforme novas cotações chegam.</div>
  const values = points.map((point) => point.price)
  const min = Math.min(...values), max = Math.max(...values), span = max - min || Math.abs(max) * .01 || 1
  const coordinates = points.map((point, index) => `${(index / Math.max(points.length - 1, 1)) * 100},${90 - ((point.price - min) / span) * 80}`).join(' ')
  return <div className="chart-wrap"><svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="Gráfico histórico de preço"><polyline points={coordinates} fill="none" stroke="#5ae0a0" strokeWidth="1.4" vectorEffect="non-scaling-stroke" /></svg><div className="chart-labels"><span>{min.toLocaleString('pt-BR', { maximumFractionDigits: 4 })}</span><span>{max.toLocaleString('pt-BR', { maximumFractionDigits: 4 })}</span></div></div>
}
