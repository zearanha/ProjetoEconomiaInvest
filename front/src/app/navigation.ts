export type Section='overview'|'stocks'|'crypto'|'fx'|'fixed-income'|'news'
export const menu:{id:Section;label:string;icon:string}[]=[{id:'overview',label:'Vis�o geral',icon:'?'},{id:'stocks',label:'A��es',icon:'?'},{id:'crypto',label:'Criptoativos',icon:'?'},{id:'fx',label:'C�mbio',icon:'?'},{id:'fixed-income',label:'Renda fixa',icon:'?'},{id:'news',label:'Not�cias',icon:'?'}]
export const marketBySection={stocks:'stocks',crypto:'crypto',fx:'fx'} as const
export const descriptions:Record<Section,string>={overview:'',stocks:'Cota��es das principais a��es brasileiras.',crypto:'Ativos digitais e varia��o recente de pre�o.',fx:'C�mbio e pares de moedas acompanhados.','fixed-income':'Indicadores e proje��es de investimentos.',news:'�ltimas not�cias econ�micas e de mercado.'}
