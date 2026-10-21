import { useEffect,useMemo,useState } from 'react'
import { getQuotes } from '../shared/api/quotes'
import { getRates } from '../shared/api/fixedIncome'
import { getNews } from '../shared/api/news'
import { useQuoteStream } from '../shared/hooks/useQuoteStream'
import { FixedIncomePage } from '../features/fixed-income/FixedIncomePage'
import { MarketPage } from '../features/markets/MarketPage'
import { NewsPage } from '../features/news/NewsPage'
import { Overview } from '../features/overview/Overview'
import { useResource } from '../shared/hooks/useResource'
import { todayLabel } from '../shared/lib/format'
import { descriptions,marketBySection,menu,type Section } from './navigation'
export default function App(){
 const [section,setSection]=useState<Section>('overview'),[watchlist,setWatchlist]=useState<string[]>(()=>{try{return JSON.parse(localStorage.getItem('market-watchlist')||'[]')}catch{return []}}),[newsCategory,setNewsCategory]=useState(''),[newsSearch,setNewsSearch]=useState(''),[newsPage,setNewsPage]=useState(1)
 const stream=useQuoteStream(),quoteState={stocks:useResource(()=>getQuotes('stocks'),[],60_000),crypto:useResource(()=>getQuotes('crypto'),[],30_000),fx:useResource(()=>getQuotes('fx'),[],60_000)},ratesState=useResource(getRates,[],300_000),newsState=useResource(()=>getNews(newsCategory,newsSearch,newsPage),[newsCategory,newsSearch,newsPage],300_000)
 const displayQuoteState={stocks:{...quoteState.stocks,data:stream.quotes.stocks?{market:'stocks',data:stream.quotes.stocks}:quoteState.stocks.data},crypto:{...quoteState.crypto,data:stream.quotes.crypto?{market:'crypto',data:stream.quotes.crypto}:quoteState.crypto.data},fx:{...quoteState.fx,data:stream.quotes.fx?{market:'fx',data:stream.quotes.fx}:quoteState.fx.data}}
 useEffect(()=>{localStorage.setItem('market-watchlist',JSON.stringify(watchlist))},[watchlist]);const allQuotes=useMemo(()=>Object.values(displayQuoteState).flatMap(state=>state.data?.data??[]),[displayQuoteState.stocks.data,displayQuoteState.crypto.data,displayQuoteState.fx.data])
 const toggleWatch=(symbol:string)=>setWatchlist(items=>items.includes(symbol)?items.filter(item=>item!==symbol):[...items,symbol]),moveWatch=(symbol:string,offset:number)=>setWatchlist(items=>{const from=items.indexOf(symbol),to=from+offset;if(from<0||to<0||to>=items.length)return items;const next=[...items];[next[from],next[to]]=[next[to],next[from]];return next}),title=menu.find(item=>item.id===section)?.label??'Vis�o geral'
 const refreshAll=()=>{Object.values(quoteState).forEach(state=>void state.refresh());void ratesState.refresh();void newsState.refresh()}
 return <div className="app-shell"><aside className="sidebar"><a className="brand" href="#inicio" onClick={e=>{e.preventDefault();setSection('overview')}}><span className="brand-mark">M</span><span>mercado<span className="brand-light">.painel</span></span></a><div className="side-caption">MENU PRINCIPAL</div><nav>{menu.map(item=><button key={item.id} className={`nav-item ${section===item.id?'active':''}`} onClick={()=>setSection(item.id)}><span className="nav-icon">{item.icon}</span>{item.label}{item.id==='news'&&<span className="nav-dot"/>}</button>)}</nav><div className="sidebar-bottom"><span className="live-dot"/> Dados atualizados automaticamente <small>Intervalo de 30 a 60 segundos</small></div></aside>
 <main className="main-content"><header className="topbar"><div className="breadcrumb">Mercado <span>/</span> <strong>{title}</strong></div><div className="top-actions"><span className={`market-status status-${stream.status}`}><i/>{stream.status==='connected'?'WebSocket conectado':stream.status==='reconnecting'?'Reconectando�':'Conectando�'}</span><span className="avatar">JM</span></div></header><div className="page-wrap"><div className="page-heading"><div><div className="eyebrow">{todayLabel.toLocaleUpperCase('pt-BR')}</div><h1>{section==='overview'?'Painel de mercado':title}</h1><p>{section==='overview'?'Acompanhe os principais indicadores do mercado.':descriptions[section]}</p></div><button className="refresh-button" onClick={refreshAll}>? <span>Atualizar</span></button></div>
 {section==='overview'&&<Overview quotes={allQuotes} quoteState={displayQuoteState} rates={ratesState.data?.data??[]} ratesError={ratesState.error} news={newsState.data?.data??[]} newsError={newsState.error} watchlist={watchlist} toggleWatch={toggleWatch} moveWatch={moveWatch} go={setSection}/>}
 {section in marketBySection&&<MarketPage market={marketBySection[section as keyof typeof marketBySection]} state={displayQuoteState[section as keyof typeof displayQuoteState]} watchlist={watchlist} toggleWatch={toggleWatch}/>}
 {section==='fixed-income'&&<FixedIncomePage rates={ratesState.data?.data??[]} error={ratesState.error} loading={ratesState.loading}/>}
 {section==='news'&&<NewsPage state={newsState} category={newsCategory} setCategory={value=>{setNewsCategory(value);setNewsPage(1)}} query={newsSearch} setQuery={value=>{setNewsSearch(value);setNewsPage(1)}} page={newsPage} setPage={setNewsPage}/>}<footer>Fontes de mercado conectadas � Valores informativos e sujeitos a atualiza��o dos provedores.</footer></div></main></div>
}
