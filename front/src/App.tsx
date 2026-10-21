import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react'
import { getNews, getQuotes, getRates, simulate, type NewsItem, type Quote, type Rate, type Simulation } from './api'

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
  const quoteState = {
    stocks: useResource(() => getQuotes('stocks'), [], 60_000),
    crypto: useResource(() => getQuotes('crypto'), [], 30_000),
    fx: useResource(() => getQuotes('fx'), [], 60_000),
  }
  const ratesState = useResource(getRates, [], 5 * 60_000)
  const newsState = useResource(() => getNews(newsCategory, newsSearch, newsPage), [newsCategory, newsSearch, newsPage], 5 * 60_000)
  useEffect(() => { localStorage.setItem('market-watchlist', JSON.stringify(watchlist)) }, [watchlist])
  const allQuotes = useMemo(() => Object.values(quoteState).flatMap((state) => state.data?.data ?? []), [quoteState.stocks.data, quoteState.crypto.data, quoteState.fx.data])
  const toggleWatch = (symbol: string) => setWatchlist((current) => current.includes(symbol) ? current.filter((item) => item !== symbol) : [...current, symbol])
  const title = menu.find((item) => item.id === section)?.label ?? 'Visão geral'

  return <div className="app-shell">
    <aside className="sidebar">
      <a className="brand" href="#inicio" onClick={(e) => { e.preventDefault(); setSection('overview') }}><span className="brand-mark">M</span><span>mercado<span className="brand-light">.painel</span></span></a>
      <div className="side-caption">MENU PRINCIPAL</div>
      <nav>{menu.map((item) => <button key={item.id} className={`nav-item ${section === item.id ? 'active' : ''}`} onClick={() => setSection(item.id)}><span className="nav-icon">{item.icon}</span>{item.label}{item.id === 'news' && <span className="nav-dot" />}</button>)}</nav>
      <div className="sidebar-bottom"><span className="live-dot" /> Dados atualizados automaticamente <small>Intervalo de 30 a 60 segundos</small></div>
    </aside>
    <main className="main-content">
      <header className="topbar"><div className="breadcrumb">Mercado <span>/</span> <strong>{title}</strong></div><div className="top-actions"><span className="market-status"><i /> Atualização automática</span><span className="avatar">JM</span></div></header>
      <div className="page-wrap">
        <div className="page-heading"><div><div className="eyebrow">{todayLabel.toLocaleUpperCase('pt-BR')}</div><h1>{section === 'overview' ? 'Painel de mercado' : title}</h1><p>{section === 'overview' ? 'Acompanhe os principais indicadores do mercado.' : descriptions[section]}</p></div><button className="refresh-button" onClick={() => { Object.values(quoteState).forEach((state) => void state.refresh()); void ratesState.refresh(); void newsState.refresh() }}>↻ <span>Atualizar</span></button></div>
        {section === 'overview' && <Overview quotes={allQuotes} quoteState={quoteState} rates={ratesState.data?.data ?? []} ratesError={ratesState.error} news={newsState.data?.data ?? []} newsError={newsState.error} watchlist={watchlist} toggleWatch={toggleWatch} go={setSection} />}
        {section in markets && <MarketPanel market={markets[section]} state={quoteState[section as keyof typeof quoteState]} watchlist={watchlist} toggleWatch={toggleWatch} />}
        {section === 'fixed-income' && <FixedIncome rates={ratesState.data?.data ?? []} error={ratesState.error} loading={ratesState.loading} />}
        {section === 'news' && <NewsPanel state={newsState} category={newsCategory} setCategory={(value) => { setNewsCategory(value); setNewsPage(1) }} query={newsSearch} setQuery={(value) => { setNewsSearch(value); setNewsPage(1) }} page={newsPage} setPage={setNewsPage} />}
        <footer>Fontes de mercado conectadas · Valores informativos e sujeitos a atualização dos provedores.</footer>
      </div>
    </main>
  </div>
}

const descriptions: Record<Section, string> = { overview: '', stocks: 'Cotações das principais ações brasileiras.', crypto: 'Ativos digitais e variação recente de preço.', fx: 'Câmbio e pares de moedas acompanhados.', 'fixed-income': 'Indicadores e projeções de investimentos.', news: 'Últimas notícias econômicas e de mercado.' }

type States = Record<string, ReturnType<typeof useResource<{ market: string; data: Quote[] }>>>
function Overview({ quotes, quoteState, rates, ratesError, news, newsError, watchlist, toggleWatch, go }: { quotes: Quote[]; quoteState: States; rates: Rate[]; ratesError: string; news: NewsItem[]; newsError: string; watchlist: string[]; toggleWatch: (symbol: string) => void; go: (section: Section) => void }) {
  const watched = quotes.filter((quote) => watchlist.includes(quote.symbol))
  const indices = rates.filter((rate) => ['SELIC', 'CDI', 'IPCA'].includes(rate.indicator))
  return <>
    <section className="stat-grid">
      <StatCard label="Ibovespa" value={bestQuote(quotes, /IBOV/i)?.price} change={bestQuote(quotes, /IBOV/i)?.changePct} suffix=" pts" />
      <StatCard label="Bitcoin" value={bestQuote(quotes, /BTC/i)?.price} change={bestQuote(quotes, /BTC/i)?.changePct} currency="USD" />
      <StatCard label="Dólar comercial" value={bestQuote(quotes, /USD|DOLAR/i)?.price} change={bestQuote(quotes, /USD|DOLAR/i)?.changePct} currency="BRL" />
      <div className="stat-card rate-card"><div className="card-label">TAXAS DE REFERÊNCIA</div>{indices.length ? indices.slice(0, 3).map((rate) => <div className="mini-rate" key={rate.indicator}><span>{rate.indicator}</span><b>{rate.value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}%</b></div>) : <div className="empty-inline">{ratesError || 'Carregando indicadores…'}</div>}</div>
    </section>
    <div className="content-grid"><section className="panel quotes-panel"><PanelTitle title="Acompanhe" subtitle="Ativos e moedas em destaque" action="Ver mercados" onAction={() => go('stocks')} />{watched.length ? <QuoteTable quotes={watched} watchlist={watchlist} toggleWatch={toggleWatch} /> : <div className="empty-state"><div className="empty-icon">☆</div><strong>Sua lista começa aqui</strong><span>Salve um ativo clicando na estrela ao lado da cotação.</span><button className="text-button" onClick={() => go('stocks')}>Explorar mercados <span>→</span></button></div>}</section>
      <section className="panel indicators-panel"><PanelTitle title="Indicadores" subtitle="Renda fixa brasileira" action="Ver detalhes" onAction={() => go('fixed-income')} />{rates.length ? rates.map((rate) => <div className="indicator-row" key={rate.indicator}><span className="indicator-icon">{rate.indicator.slice(0, 1)}</span><div className="indicator-name"><b>{rate.indicator}</b><small>{rate.source} · {dateTime.format(new Date(rate.date))}</small></div><strong>{rate.value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}{rate.unit}</strong></div>) : <EmptyMessage error={ratesError} />}</section></div>
    <section className="panel news-preview"><PanelTitle title="Em destaque" subtitle="Notícias recentes" action="Todas as notícias" onAction={() => go('news')} />{newsError ? <EmptyMessage error={newsError} /> : news.length ? <div className="news-grid">{news.slice(0, 3).map((item) => <NewsCard key={item.id} item={item} />)}</div> : <EmptyMessage error="As notícias ainda estão sendo carregadas." />}</section>
    {Object.values(quoteState).some((state) => state.error) && <div className="notice">Algumas fontes de cotações estão indisponíveis. Os demais dados continuam acessíveis.</div>}
  </>
}

function MarketPanel({ market, state, watchlist, toggleWatch }: { market: string; state: ReturnType<typeof useResource<{ market: string; data: Quote[] }>>; watchlist: string[]; toggleWatch: (symbol: string) => void }) {
  const names: Record<string, string> = { stocks: 'Ações', crypto: 'Criptoativos', fx: 'Câmbio' }
  return <section className="panel market-panel"><PanelTitle title={`Cotações de ${names[market]}`} subtitle="Atualizadas automaticamente" action="Atualização em até 60s" /><QuoteTable quotes={state.data?.data ?? []} watchlist={watchlist} toggleWatch={toggleWatch} loading={state.loading} error={state.error} /></section>
}

function QuoteTable({ quotes, watchlist, toggleWatch, loading, error }: { quotes: Quote[]; watchlist: string[]; toggleWatch: (symbol: string) => void; loading?: boolean; error?: string }) {
  if (error && !quotes.length) return <EmptyMessage error={error} />
  if (loading && !quotes.length) return <EmptyMessage error="Buscando cotações…" />
  if (!quotes.length) return <EmptyMessage error="Nenhuma cotação disponível neste momento." />
  return <div className="table-scroll"><table><thead><tr><th>ATIVO</th><th>PREÇO</th><th>VARIAÇÃO</th><th>MÁX / MÍN</th><th>ATUALIZADO</th><th /></tr></thead><tbody>{quotes.map((quote) => <tr key={quote.symbol}><td><div className="asset-cell"><span className={`coin-icon ${quote.market}`}>{quote.symbol.slice(0, 1)}</span><span><b>{quote.symbol}</b><small>{quote.name}</small></span></div></td><td className="price-cell">{formatPrice(quote)}</td><td><span className={`change-pill ${quote.changePct >= 0 ? 'positive' : 'negative'}`}>{quote.changePct >= 0 ? '+' : ''}{quote.changePct.toFixed(2)}%</span></td><td className="muted-cell">{quote.high || quote.low ? `${formatPrice(quote, quote.high)} / ${formatPrice(quote, quote.low)}` : '—'}</td><td className="muted-cell">{dateTime.format(new Date(quote.updatedAt))}</td><td><button className={`star-button ${watchlist.includes(quote.symbol) ? 'saved' : ''}`} title="Salvar na lista" onClick={() => toggleWatch(quote.symbol)}>{watchlist.includes(quote.symbol) ? '★' : '☆'}</button></td></tr>)}</tbody></table></div>
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
