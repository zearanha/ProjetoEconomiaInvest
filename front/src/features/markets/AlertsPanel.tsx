import { useEffect,useState } from 'react'
import type { PriceAlert } from '../../shared/types/market'
export function AlertsPanel() {
  const [alerts, setAlerts] = useState<PriceAlert[]>(() => { try { return JSON.parse(localStorage.getItem('market-price-alerts') || '[]') } catch { return [] } })
  useEffect(() => { const sync = () => { try { setAlerts(JSON.parse(localStorage.getItem('market-price-alerts') || '[]')) } catch { setAlerts([]) } }; window.addEventListener('market-alerts-updated', sync); return () => window.removeEventListener('market-alerts-updated', sync) }, [])
  if (!alerts.length) return null
  return <div className="alerts-panel"><strong>Alertas de preço</strong>{alerts.map((alert) => <span key={alert.symbol}>{alert.symbol} {alert.direction === 'above' ? '≥' : '≤'} {alert.target.toLocaleString('pt-BR')} {alert.notified ? '· disparado' : ''}<button title={`Remover alerta de ${alert.symbol}`} onClick={() => { const next = alerts.filter((item) => item.symbol !== alert.symbol); localStorage.setItem('market-price-alerts', JSON.stringify(next)); setAlerts(next) }}>×</button></span>)}</div>
}

