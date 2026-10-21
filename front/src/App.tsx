import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react'
import { getNews, getQuotes, getQuoteHistory, getRates, simulate, type NewsItem, type Quote, type QuotePoint, type Rate, type Simulation } from './api'
import { useQuoteStream } from './hooks/useQuoteStream'

type Section = 'overview' | 'stocks' | 'crypto' | 'fx' | 'fixed-income' | 'news'
const menu: { id: Section; label: string; icon: string }[] = [
  { id: 'overview', label: 'Visão geral', icon: '◫' }, { id: 'stocks', label: 'Ações', icon: '↗' },
  { id: 'crypto', label: 'Criptoativos', icon: '₿' }, { id: 'fx', label: 'Câmbio', icon: '◎' },
  { id: 'fixed-income', label: 'Renda fixa', icon: '▤' }, { id: 'news', label: 'Notícias', icon: '▧' },
]
const markets: Record<string, string> = { stocks: 'stocks', crypto: 'crypto', fx: 'fx' }
const brl = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' })
const usd = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'USD' })
const dateTime = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
const todayLabel = new Intl.DateTimeFormat('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' }).format(new Date())

function useResource<T>(load: () => Promise<T>, deps: unknown[] = [], interval = 60_000) {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const refresh = useCallback(async () => {
    try { setError(''); setData(await load()) } catch (e) { setError(e instanceof Error ? e.message : 'Falha ao carregar dados') }
    finally { setLoading(false) }
  // load is memoized at call sites; deps define when the request changes.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)
  useEffect(() => { void refresh(); const timer = window.setInterval(() => void refresh(), interval); return () => window.clearInterval(timer) }, [refresh, interval])
  return { data, error, loading, refresh }
}

function App() {
  const [section, setSection] = useState<Section>('overview')
  const [watchlist, setWatchlist] = useState<string[]>(() => { try { return JSON.parse(localStorage.getItem('market-watchlist') || '[]') } catch { return [] } })
  const [newsCategory, setNewsCategory] = useState('')
  const [newsSearch, setNewsSearch] = useState('')
  const [newsPage, setNewsPage] = useState(1)
  const stream = useQuoteStream()
  const quoteState = {
    stocks: useResource(() => getQuotes('stocks'), [], 60_000),
    crypto: useResource(() => getQuotes('crypto'), [], 30_000),
    fx: useResource(() => getQuotes('fx'), [], 60_000),
  }
  const ratesState = useResource(getRates, [], 5 * 60_000)
  const newsState = useResource(() => getNews(newsCategory, newsSearch, newsPage), [newsCategory, newsSearch, newsPage], 5 * 60_000)
  const displayQuoteState = {
    stocks: { ...quoteState.stocks, data: stream.quotes.stocks ? { market: 'stocks', data: stream.quotes.stocks } : quoteState.stocks.data },
    crypto: { ...quoteState.crypto, data: stream.quotes.crypto ? { market: 'crypto', data: stream.quotes.crypto } : quoteState.crypto.data },
    fx: { ...quoteState.fx, data: stream.quotes.fx ? { market: 'fx', data: stream.quotes.fx } : quoteState.fx.data },
  }
  useEffect(() => { localStorage.setItem('market-watchlist', JSON.stringify(watchlist)) }, [watchlist])
  const allQuotes = useMemo(() => Object.values(displayQuoteState).flatMap((state) => state.data?.data ?? []), [displayQuoteState.stocks.data, displayQuoteState.crypto.data, displayQuoteState.fx.data])
  const toggleWatch = (symbol: string) => setWatchlist((current) => current.includes(symbol) ? current.filter((item) => item !== symbol) : [...current, symbol])
  const moveWatch = (symbol: string, offset: number) => setWatchlist((current) => { const from = current.indexOf(symbol), to = from + offset; if (from < 0 || to < 0 || to >= current.length) return current; const next = [...current]; [next[from], next[to]] = [next[to], next[from]]; return next })
  const title = menu.find((item) => item.id === section)?.label ?? 'Visão geral'

  return <div className="app-shell">
    <aside className="sidebar">
      <a className="brand" href="#inicio" onClick={(e) => { e.preventDefault(); setSection('overview') }}><span className="brand-mark">M</span><span>mercado<span className="brand-light">.painel</span></span></a>
      <div className="side-caption">MENU PRINCIPAL</div>
      <nav>{menu.map((item) => <button key={item.id} className={`nav-item ${section === item.id ? 'active' : ''}`} onClick={() => setSection(item.id)}><span className="nav-icon">{item.icon}</span>{item.label}{item.id === 'news' && <span className="nav-dot" />}</button>)}</nav>
      <div className="sidebar-bottom"><span className="live-dot" /> Dados atualizados automaticamente <small>Intervalo de 30 a 60 segundos</small></div>
    </aside>
    <main className="main-content">
      <header className="topbar"><div className="breadcrumb">Mercado <span>/</span> <strong>{title}</strong></div><div className="top-actions"><span className={`market-status status-${stream.status}`}><i /> {stream.status === 'connected' ? 'WebSocket conectado' : stream.status === 'reconnecting' ? 'Reconectando…' : 'Conectando…'}</span><span className="avatar">JM</span></div></header>
      <div className="page-wrap">
        <div className="page-heading"><div><div className="eyebrow">{todayLabel.toLocaleUpperCase('pt-BR')}</div><h1>{section === 'overview' ? 'Painel de mercado' : title}</h1><p>{section === 'overview' ? 'Acompanhe os principais indicadores do mercado.' : descriptions[section]}</p></div><button className="refresh-button" onClick={() => { Object.values(quoteState).forEach((state) => void state.refresh()); void ratesState.refresh(); void newsState.refresh() }}>↻ <span>Atualizar</span></button></div>
        {section === 'overview' && <Overview quotes={allQuotes} quoteState={displayQuoteState} rates={ratesState.data?.data ?? []} ratesError={ratesState.error} news={newsState.data?.data ?? []} newsError={newsState.error} watchlist={watchlist} toggleWatch={toggleWatch} moveWatch={moveWatch} go={setSection} />}
        {section in markets && <MarketPanel market={markets[section]} state={displayQuoteState[section as keyof typeof displayQuoteState]} watchlist={watchlist} toggleWatch={toggleWatch} />}
        {section === 'fixed-income' && <FixedIncome rates={ratesState.data?.data ?? []} error={ratesState.error} loading={ratesState.loading} />}
        {section === 'news' && <NewsPanel state={newsState} category={newsCategory} setCategory={(value) => { setNewsCategory(value); setNewsPage(1) }} query={newsSearch} setQuery={(value) => { setNewsSearch(value); setNewsPage(1) }} page={newsPage} setPage={setNewsPage} />}
        <footer>Fontes de mercado conectadas · Valores informativos e sujeitos a atualização dos provedores.</footer>
      </div>
    </main>
  </div>
}

const descriptions: Record<Section, string> = { overview: '', stocks: 'Cotações das principais ações brasileiras.', crypto: 'Ativos digitais e variação recente de preço.', fx: 'Câmbio e pares de moedas acompanhados.', 'fixed-income': 'Indicadores e projeções de investimentos.', news: 'Últimas notícias econômicas e de mercado.' }

type States = Record<string, ReturnType<typeof useResource<{ market: string; data: Quote[] }>>>
function Overview({ quotes, quoteState, rates, ratesError, news, newsError, watchlist, toggleWatch, moveWatch, go }: { quotes: Quote[]; quoteState: States; rates: Rate[]; ratesError: string; news: NewsItem[]; newsError: string; watchlist: string[]; toggleWatch: (symbol: string) => void; moveWatch: (symbol:string, offset:number) => void; go: (section: Section) => void }) {
  const watched = watchlist.flatMap((symbol) => quotes.filter((quote) => quote.symbol === symbol))
  const indices = rates.filter((rate) => ['SELIC', 'CDI', 'IPCA'].includes(rate.indicator))
  return <>
    <section className="stat-grid">
      <StatCard label="Ibovespa" value={bestQuote(quotes, /IBOV/i)?.price} change={bestQuote(quotes, /IBOV/i)?.changePct} suffix=" pts" />
      <StatCard label="Bitcoin" value={bestQuote(quotes, /BTC/i)?.price} change={bestQuote(quotes, /BTC/i)?.changePct} currency="USD" />
      <StatCard label="Dólar comercial" value={bestQuote(quotes, /USD|DOLAR/i)?.price} change={bestQuote(quotes, /USD|DOLAR/i)?.changePct} currency="BRL" />
      <div className="stat-card rate-card"><div className="card-label">TAXAS DE REFERÊNCIA</div>{indices.length ? indices.slice(0, 3).map((rate) => <div className="mini-rate" key={rate.indicator}><span>{rate.indicator}</span><b>{rate.value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}%</b></div>) : <div className="empty-inline">{ratesError || 'Carregando indicadores…'}</div>}</div>
    </section>
    <div className="content-grid"><section className="panel quotes-panel"><PanelTitle title="Acompanhe" subtitle="Ativos e moedas em destaque" action="Ver mercados" onAction={() => go('stocks')} />{watched.length ? <QuoteTable quotes={watched} watchlist={watchlist} toggleWatch={toggleWatch} moveWatch={moveWatch} /> : <div className="empty-state"><div className="empty-icon">☆</div><strong>Sua lista começa aqui</strong><span>Salve um ativo clicando na estrela ao lado da cotação.</span><button className="text-button" onClick={() => go('stocks')}>Explorar mercados <span>→</span></button></div>}</section>
      <section className="panel indicators-panel"><PanelTitle title="Indicadores" subtitle="Renda fixa brasileira" action="Ver detalhes" onAction={() => go('fixed-income')} />{rates.length ? rates.map((rate) => <div className="indicator-row" key={rate.indicator}><span className="indicator-icon">{rate.indicator.slice(0, 1)}</span><div className="indicator-name"><b>{rate.indicator}</b><small>{rate.source} · {dateTime.format(new Date(rate.date))}</small></div><strong>{rate.value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}{rate.unit}</strong></div>) : <EmptyMessage error={ratesError} />}</section></div>
    <section className="panel news-preview"><PanelTitle title="Em destaque" subtitle="Notícias recentes" action="Todas as notícias" onAction={() => go('news')} />{newsError ? <EmptyMessage error={newsError} /> : news.length ? <div className="news-grid">{news.slice(0, 3).map((item) => <NewsCard key={item.id} item={item} />)}</div> : <EmptyMessage error="As notícias ainda estão sendo carregadas." />}</section>
    {Object.values(quoteState).some((state) => state.error) && <div className="notice">Algumas fontes de cotações estão indisponíveis. Os demais dados continuam acessíveis.</div>}
  </>
}

function MarketPanel({ market, state, watchlist, toggleWatch }: { market: string; state: ReturnType<typeof useResource<{ market: string; data: Quote[] }>>; watchlist: string[]; toggleWatch: (symbol: string) => void }) {
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

type PriceAlert = { symbol:string; target:number; direction:'above'|'below'; notified?:boolean }
function AlertsPanel() {
  const [alerts, setAlerts] = useState<PriceAlert[]>(() => { try { return JSON.parse(localStorage.getItem('market-price-alerts') || '[]') } catch { return [] } })
  useEffect(() => { const sync = () => { try { setAlerts(JSON.parse(localStorage.getItem('market-price-alerts') || '[]')) } catch { setAlerts([]) } }; window.addEventListener('market-alerts-updated', sync); return () => window.removeEventListener('market-alerts-updated', sync) }, [])
  if (!alerts.length) return null
  return <div className="alerts-panel"><strong>Alertas de preço</strong>{alerts.map((alert) => <span key={alert.symbol}>{alert.symbol} {alert.direction === 'above' ? '≥' : '≤'} {alert.target.toLocaleString('pt-BR')} {alert.notified ? '· disparado' : ''}<button title={`Remover alerta de ${alert.symbol}`} onClick={() => { const next = alerts.filter((item) => item.symbol !== alert.symbol); localStorage.setItem('market-price-alerts', JSON.stringify(next)); setAlerts(next) }}>×</button></span>)}</div>
}

function QuoteChart({ points }: { points: QuotePoint[] }) {
  if (!points.length) return <div className="chart-empty">O histórico começa a ser formado conforme novas cotações chegam.</div>
  const values = points.map((point) => point.price)
  const min = Math.min(...values), max = Math.max(...values), span = max - min || Math.abs(max) * .01 || 1
  const coordinates = points.map((point, index) => `${(index / Math.max(points.length - 1, 1)) * 100},${90 - ((point.price - min) / span) * 80}`).join(' ')
  return <div className="chart-wrap"><svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="Gráfico histórico de preço"><polyline points={coordinates} fill="none" stroke="#5ae0a0" strokeWidth="1.4" vectorEffect="non-scaling-stroke" /></svg><div className="chart-labels"><span>{min.toLocaleString('pt-BR', { maximumFractionDigits: 4 })}</span><span>{max.toLocaleString('pt-BR', { maximumFractionDigits: 4 })}</span></div></div>
}

function QuoteTable({ quotes, watchlist, toggleWatch, moveWatch, loading, error }: { quotes: Quote[]; watchlist: string[]; toggleWatch: (symbol: string) => void; moveWatch?: (symbol:string, offset:number) => void; loading?: boolean; error?: string }) {
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
function formatPrice(quote: Quote, value = quote.price) { return (quote.currency === 'BRL' ? brl : quote.currency === 'USD' || quote.currency === 'USDT' ? usd : new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 4 })).format(value) }
function bestQuote(quotes: Quote[], pattern: RegExp) { return quotes.find((quote) => pattern.test(`${quote.symbol} ${quote.name}`)) }
function StatCard({ label, value, change, currency, suffix }: { label: string; value?: number; change?: number; currency?: string; suffix?: string }) { return <div className="stat-card"><div className="stat-top"><span className="card-label">{label}</span><span className="sparkline">⌁</span></div><div className="stat-value">{value == null ? '—' : currency === 'USD' ? usd.format(value) : currency === 'BRL' ? brl.format(value) : `${value.toLocaleString('pt-BR')}${suffix ?? ''}`}</div><div className={`stat-change ${change == null ? '' : change >= 0 ? 'up' : 'down'}`}>{change == null ? 'Aguardando fonte' : `${change >= 0 ? '↑' : '↓'} ${Math.abs(change).toFixed(2)}%`} <span>no período informado</span></div></div> }
function PanelTitle({ title, subtitle, action, onAction }: { title: string; subtitle: string; action?: string; onAction?: () => void }) { return <div className="panel-title"><div><h2>{title}</h2><p>{subtitle}</p></div>{action && <button className="panel-action" onClick={onAction}>{action} <span>→</span></button>}</div> }
function EmptyMessage({ error }: { error: string }) { return <div className="empty-message"><span className="empty-icon">⌁</span><span>{error}</span></div> }

function FixedIncome({ rates, error, loading }: { rates: Rate[]; error: string; loading: boolean }) {
  const [amount, setAmount] = useState('10000')
  const [termDays, setTermDays] = useState('365')
  const [cdb, setCdb] = useState('110')
  const [lci, setLci] = useState('90')
  const [lca, setLca] = useState('90')
  const [result, setResult] = useState<Simulation | null>(null)
  const [simulationError, setSimulationError] = useState('')
  const [busy, setBusy] = useState(false)
  async function submit(event: FormEvent) { event.preventDefault(); setBusy(true); setSimulationError(''); try { setResult(await simulate({ amount: Number(amount), termDays: Number(termDays), cdbPercentCDI: Number(cdb), lciPercentCDI: Number(lci), lcaPercentCDI: Number(lca) })) } catch (e) { setSimulationError(e instanceof Error ? e.message : 'Não foi possível simular.') } finally { setBusy(false) } }
  return <><div className="rate-cards">{['SELIC', 'CDI', 'IPCA'].map((name) => { const rate = rates.find((item) => item.indicator === name); return <div className="panel rate-detail" key={name}><span className="card-label">{name}</span><strong>{rate ? `${rate.value.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}${rate.unit}` : loading ? '…' : '—'}</strong><small>{rate ? `Fonte: ${rate.source}` : error || 'Indicador indisponível'}</small></div> })}</div><div className="panel simulator"><PanelTitle title="Simulador de renda fixa" subtitle="Compare uma projeção de CDB, LCI e LCA com a taxa CDI disponível." /><form className="sim-form" onSubmit={submit}><label>Valor aplicado <div className="input-prefix"><span>R$</span><input type="number" min="1" step="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} required /></div></label><label>Prazo em dias<input type="number" min="1" max="36500" value={termDays} onChange={(e) => setTermDays(e.target.value)} required /></label><label>CDB (% do CDI)<input type="number" min="1" max="500" value={cdb} onChange={(e) => setCdb(e.target.value)} required /></label><label>LCI (% do CDI)<input type="number" min="1" max="500" value={lci} onChange={(e) => setLci(e.target.value)} required /></label><label>LCA (% do CDI)<input type="number" min="1" max="500" value={lca} onChange={(e) => setLca(e.target.value)} required /></label><button className="primary-button" disabled={busy}>{busy ? 'Calculando…' : 'Calcular projeção'}</button></form>{simulationError && <div className="form-error">{simulationError}</div>}{result && <div className="simulation-result"><div className="result-heading"><div><span className="card-label">PROJEÇÃO LÍQUIDA</span><p>CDI considerado: {result.cdiAnnualPercent.toLocaleString('pt-BR')}% a.a. · {result.estimatedBusinessDays} dias úteis estimados</p></div></div><div className="result-grid">{result.results.map((item) => <div className="result-card" key={item.product}><span>{item.product} <small>{item.incomeTaxPercent ? `IR ${item.incomeTaxPercent}%` : 'Isento de IR'}</small></span><b>{brl.format(item.netAmount)}</b><small>Rendimento líquido {brl.format(item.netProfit)}</small></div>)}</div><p className="assumption">{result.assumption}</p></div>}</div></>
}

function NewsPanel({ state, category, setCategory, query, setQuery, page, setPage }: { state: ReturnType<typeof useResource<import('./api').NewsPage>>; category: string; setCategory: (value: string) => void; query: string; setQuery: (value: string) => void; page: number; setPage: (value: number) => void }) {
  const categories = [['', 'Todas'], ['brasil', 'Brasil'], ['global', 'Global'], ['crypto', 'Cripto'], ['fixed-income', 'Renda fixa'], ['stocks', 'Ações']]
  const maxPage = Math.max(1, Math.ceil((state.data?.total ?? 0) / (state.data?.pageSize ?? 20)))
  return <section className="panel news-panel"><PanelTitle title="Notícias de mercado" subtitle="Acompanhe os assuntos que movimentam os mercados." /><div className="news-controls"><div className="category-tabs">{categories.map(([value, label]) => <button key={value} className={category === value ? 'selected' : ''} onClick={() => setCategory(value)}>{label}</button>)}</div><label className="search-box"><span>⌕</span><input placeholder="Buscar notícias" value={query} onChange={(e) => setQuery(e.target.value)} /></label></div>{state.error ? <EmptyMessage error={state.error} /> : state.loading && !state.data ? <EmptyMessage error="Carregando notícias…" /> : state.data?.data.length ? <div className="news-list">{state.data.data.map((item) => <NewsRow key={item.id} item={item} />)}</div> : <EmptyMessage error="Nenhuma notícia encontrada com estes filtros." />}<div className="pagination"><span>{state.data?.total ?? 0} notícias encontradas</span><div><button disabled={page <= 1} onClick={() => setPage(page - 1)}>← Anterior</button><span>Página {page} de {maxPage}</span><button disabled={page >= maxPage} onClick={() => setPage(page + 1)}>Próxima →</button></div></div></section>
}
function NewsRow({ item }: { item: NewsItem }) { return <article className="news-row"><div className="news-category">{item.category.replace('-', ' ')}</div><div className="news-copy"><a href={item.url} target="_blank" rel="noreferrer">{item.title} ↗</a><p>{item.summary}</p><small>{item.source} · {dateTime.format(new Date(item.publishedAt))}</small></div></article> }
function NewsCard({ item }: { item: NewsItem }) { return <article className="news-card"><div className="news-category">{item.category.replace('-', ' ')}</div><a href={item.url} target="_blank" rel="noreferrer">{item.title} ↗</a><p>{item.summary}</p><small>{item.source} · {dateTime.format(new Date(item.publishedAt))}</small></article> }

export default App
