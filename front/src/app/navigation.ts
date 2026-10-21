export type Section = 'overview' | 'stocks' | 'crypto' | 'fx' | 'fixed-income' | 'news'

export const menu: { id: Section; label: string; icon: string }[] = [
  { id: 'overview', label: 'Visão geral', icon: '▦' },
  { id: 'stocks', label: 'Ações', icon: '↗' },
  { id: 'crypto', label: 'Criptoativos', icon: '₿' },
  { id: 'fx', label: 'Câmbio', icon: '◎' },
  { id: 'fixed-income', label: 'Renda fixa', icon: '▤' },
  { id: 'news', label: 'Notícias', icon: '▧' },
]

export const marketBySection = { stocks: 'stocks', crypto: 'crypto', fx: 'fx' } as const

export const descriptions: Record<Section, string> = {
  overview: '',
  stocks: 'Cotações das principais ações brasileiras.',
  crypto: 'Ativos digitais e variação recente de preço.',
  fx: 'Câmbio e pares de moedas acompanhados.',
  'fixed-income': 'Indicadores e projeções de investimentos.',
  news: 'Últimas notícias econômicas e de mercado.',
}
