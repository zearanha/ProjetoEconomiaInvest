import type { NewsItem } from '../../shared/types/api'
import { dateTime } from '../../shared/lib/format'
export function NewsRow({ item }: { item: NewsItem }) { return <article className="news-row"><div className="news-category">{item.category.replace('-', ' ')}</div><div className="news-copy"><a href={item.url} target="_blank" rel="noreferrer">{item.title} ↗</a><p>{item.summary}</p><small>{item.source} · {dateTime.format(new Date(item.publishedAt))}</small></div></article> }
export function NewsCard({ item }: { item: NewsItem }) { return <article className="news-card"><div className="news-category">{item.category.replace('-', ' ')}</div><a href={item.url} target="_blank" rel="noreferrer">{item.title} ↗</a><p>{item.summary}</p><small>{item.source} · {dateTime.format(new Date(item.publishedAt))}</small></article> }

