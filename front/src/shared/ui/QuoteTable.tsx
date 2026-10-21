import { useEffect,useState } from 'react'
import type { Quote } from '../types/api'
import { dateTime,formatPrice } from '../lib/format'
import type { PriceAlert } from '../types/market'
import { EmptyMessage } from './MarketWidgets'
export function QuoteTable({ quotes, watchlist, toggleWatch, moveWatch, loading, error }: { quotes: Quote[]; watchlist: string[]; toggleWatch: (symbol: string) => void; moveWatch?: (symbol:string, offset:number) => void; loading?: boolean; error?: string }) {
  const [, setAlertRefresh] = useState(0)
  useEffect(() => {
    let alerts: { symbol: string; target: number; direction: 'above'|'below'; notified?: boolean }[] = []
    try { alerts = JSON.parse(localStorage.getItem('market-price-alerts') || '[]') } catch { /* ignore invalid saved state */ }
    const fired: string[] = []
    const next = alerts.map((alert) => {
      const quote = quotes.find((item) => item.symbol === alert.symbol)
      const hit = quote && (alert.direction === 'above' ? quote.price >= alert.target : quote.price <= alert.target)
      if (hit && !alert.notified) {
        fired.push(`${alert.symbol} atingiu ${formatPrice(quote!, alert.target)}`)
        if ('Notification' in window && Notification.permission === 'granted') new Notification('Alerta de preço', { body: fired[fired.length - 1] })
        return { ...alert, notified: true }
      }
      return alert
    })
    if (fired.length) { localStorage.setItem('market-price-alerts', JSON.stringify(next)); setAlertRefresh((n) => n + 1); window.dispatchEvent(new Event('market-alerts-updated')); if (!('Notification' in window) || Notification.permission !== 'granted') window.alert(fired.join('\n')) }
  }, [quotes])
  if (error && !quotes.length) return <EmptyMessage error={error} />
  if (loading && !quotes.length) return <EmptyMessage error="Buscando cotações…" />
  if (!quotes.length) return <EmptyMessage error="Nenhuma cotação disponível neste momento." />
  return <div className="table-scroll"><table><thead><tr><th>ATIVO</th><th>PREÇO</th><th>VARIAÇÃO</th><th>VOLUME</th><th>MÁX / MÍN</th><th>ATUALIZADO</th><th /></tr></thead><tbody>{quotes.map((quote) => <tr key={quote.symbol}><td><div className="asset-cell"><span className={`coin-icon ${quote.market}`}>{quote.symbol.slice(0, 1)}</span><span><b>{quote.symbol}</b><small>{quote.name}</small></span></div></td><td className="price-cell">{formatPrice(quote)}</td><td><span className={`change-pill ${quote.changePct >= 0 ? 'positive' : 'negative'}`}>{quote.changePct >= 0 ? '+' : ''}{quote.changePct.toFixed(2)}%</span></td><td className="muted-cell">{quote.volume ? quote.volume.toLocaleString('pt-BR', { notation: 'compact', maximumFractionDigits: 2 }) : '—'}</td><td className="muted-cell">{quote.high || quote.low ? `${formatPrice(quote, quote.high)} / ${formatPrice(quote, quote.low)}` : '—'}</td><td className="muted-cell">{dateTime.format(new Date(quote.updatedAt))}</td><td><button className={`star-button ${watchlist.includes(quote.symbol) ? 'saved' : ''}`} title="Salvar na lista" onClick={() => toggleWatch(quote.symbol)}>{watchlist.includes(quote.symbol) ? '★' : '☆'}</button>{moveWatch && <><button className="order-button" title="Mover para cima" onClick={() => moveWatch(quote.symbol,-1)}>↑</button><button className="order-button" title="Mover para baixo" onClick={() => moveWatch(quote.symbol,1)}>↓</button></>}<button className="alert-button" title="Criar alerta de preço" onClick={async () => { const target = window.prompt(`Disparar alerta quando ${quote.symbol} chegar a qual preço?`); if (!target || !Number.isFinite(Number(target)) || Number(target) <= 0) return; const direction = Number(target) >= quote.price ? 'above' : 'below'; let alerts: PriceAlert[] = []; try { alerts = JSON.parse(localStorage.getItem('market-price-alerts') || '[]') } catch { /* reset invalid state */ }; alerts = [...alerts.filter((a) => a.symbol !== quote.symbol), { symbol:quote.symbol, target:Number(target), direction, notified:false }]; localStorage.setItem('market-price-alerts', JSON.stringify(alerts)); window.dispatchEvent(new Event('market-alerts-updated')); if ('Notification' in window && Notification.permission === 'default') await Notification.requestPermission(); window.alert(`Alerta criado: ${quote.symbol} ${direction === 'above' ? 'acima de' : 'abaixo de'} ${target}`) }}>♢</button></td></tr>)}</tbody></table></div>
}
