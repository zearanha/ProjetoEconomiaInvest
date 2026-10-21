import type { Quote } from '../types/api'
export const brl=new Intl.NumberFormat('pt-BR',{style:'currency',currency:'BRL'}), usd=new Intl.NumberFormat('pt-BR',{style:'currency',currency:'USD'}), dateTime=new Intl.DateTimeFormat('pt-BR',{dateStyle:'short',timeStyle:'short'}), todayLabel=new Intl.DateTimeFormat('pt-BR',{weekday:'long',day:'numeric',month:'long'}).format(new Date())
export function formatPrice(quote:Quote,value=quote.price){return (quote.currency==='BRL'?brl:quote.currency==='USD'||quote.currency==='USDT'?usd:new Intl.NumberFormat('pt-BR',{maximumFractionDigits:4})).format(value)}
export function bestQuote(quotes:Quote[],pattern:RegExp){return quotes.find(q=>pattern.test(`${q.symbol} ${q.name}`))}
