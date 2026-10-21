import { useEffect,useMemo,useState } from 'react'
import { getQuoteHistory } from '../../shared/api/quotes'
import type { Quote,QuotePoint } from '../../shared/types/api'
import type { ResourceState } from '../../shared/hooks/useResource'
import { dateTime } from '../../shared/lib/format'
import { QuoteTable } from '../../shared/ui/QuoteTable'
import { PanelTitle } from '../../shared/ui/MarketWidgets'
import { AlertsPanel } from './AlertsPanel'
import { QuoteChart } from './QuoteChart'
export function MarketPage({ market, state, watchlist, toggleWatch }: { market: string; state: ResourceState<{ market: string; data: Quote[] }>; watchlist: string[]; toggleWatch: (symbol: string) => void }) {
  const names: Record<string, string> = { stocks: 'Ações', crypto: 'Criptoativos', fx: 'Câmbio' }
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('name')
  const [direction, setDirection] = useState('all')
  const [favoritesOnly, setFavoritesOnly] = useState(false)
  const [selected, setSelected] = useState('')
  const [range, setRange] = useState('1d')
  const [history, setHistory] = useState<QuotePoint[]>([])
  const quotes = state.data?.data ?? []
  const filtered = useMemo(() => quotes.filter((q) => `${q.symbol} ${q.name}`.toLowerCase().includes(query.toLowerCase()) && (direction === 'all' || (direction === 'up' ? q.changePct >= 0 : q.changePct < 0)) && (!favoritesOnly || watchlist.includes(q.symbol))).sort((a, b) => {
    const key = sort.replace('-desc', '')
    const value = key === 'price' ? a.price - b.price : key === 'change' ? a.changePct - b.changePct : key === 'volume' ? (a.volume ?? 0) - (b.volume ?? 0) : a.name.localeCompare(b.name)
    return sort.endsWith('-desc') ? -value : value
  }), [quotes, query, sort, direction, favoritesOnly, watchlist])
  useEffect(() => {
    if (!selected && quotes.length) setSelected(quotes[0].symbol)
    if (selected && !quotes.some((quote) => quote.symbol === selected)) setSelected(quotes[0]?.symbol ?? '')
  }, [quotes, selected])
  useEffect(() => { if (!selected) return; let active = true; getQuoteHistory(selected, range).then((result) => { if (active) setHistory(result.data) }).catch(() => { if (active) setHistory([]) }); return () => { active = false } }, [selected, range, state.data])
  const lastUpdated = quotes.reduce((latest, quote) => Math.max(latest, new Date(quote.updatedAt).getTime()), 0)
  return <section className="panel market-panel"><PanelTitle title={`Cotações de ${names[market]}`} subtitle={state.error ? `Fonte indisponível · ${state.error}` : `Fonte ativa · atualizado ${lastUpdated ? dateTime.format(lastUpdated) : 'aguardando dados'}`} action="Atualização automática" />
    <div className="market-controls"><input aria-label="Buscar ativo" placeholder="Buscar por nome ou símbolo" value={query} onChange={(e) => setQuery(e.target.value)} /><select aria-label="Ordenar ativos" value={sort} onChange={(e) => setSort(e.target.value)}><option value="name">Nome (A–Z)</option><option value="name-desc">Nome (Z–A)</option><option value="price">Preço (menor)</option><option value="price-desc">Preço (maior)</option><option value="change">Variação (menor)</option><option value="change-desc">Variação (maior)</option><option value="volume">Volume (menor)</option><option value="volume-desc">Volume (maior)</option></select><select aria-label="Filtrar variação" value={direction} onChange={(e) => setDirection(e.target.value)}><option value="all">Todas as variações</option><option value="up">Em alta</option><option value="down">Em baixa</option></select><label className="favorite-filter"><input type="checkbox" checked={favoritesOnly} onChange={(e) => setFavoritesOnly(e.target.checked)} /> Acompanhados</label></div>
    {selected && <div className="chart-panel"><div className="chart-heading"><strong>Histórico · {selected}</strong><div>{[['1d','1 dia'],['1w','1 semana'],['1m','1 mês']].map(([value,label]) => <button key={value} className={range === value ? 'selected' : ''} onClick={() => setRange(value)}>{label}</button>)}</div></div><QuoteChart points={history} /></div>}
    <AlertsPanel />
    <QuoteTable quotes={filtered} watchlist={watchlist} toggleWatch={toggleWatch} loading={state.loading} error={state.error} />
    {!state.error && <div className="source-health"><span className="live-dot" /> Provedor respondendo · {quotes.length} ativos recebidos</div>}
  </section>
}

