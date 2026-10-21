import type { NewsItem,Quote,Rate } from '../../shared/types/api'
import type { Section } from '../../app/navigation'
import type { ResourceState } from '../../shared/hooks/useResource'
import { dateTime } from '../../shared/lib/format'
import { bestQuote } from '../../shared/lib/format'
import { EmptyMessage,PanelTitle,StatCard } from '../../shared/ui/MarketWidgets'
import { QuoteTable } from '../../shared/ui/QuoteTable'
import { NewsCard } from '../news/NewsCards'
type States=Record<string,ResourceState<{market:string;data:Quote[]}>>
export function Overview({ quotes, quoteState, rates, ratesError, news, newsError, watchlist, toggleWatch, moveWatch, go }: { quotes: Quote[]; quoteState: States; rates: Rate[]; ratesError: string; news: NewsItem[]; newsError: string; watchlist: string[]; toggleWatch: (symbol: string) => void; moveWatch: (symbol:string, offset:number) => void; go: (section: Section) => void }) {
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
