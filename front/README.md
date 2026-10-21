# Frontend do Painel de Mercado

## Organização do código

O ponto de entrada `src/main.tsx` carrega o shell em `src/app/App.tsx`. O shell controla navegação e estado compartilhado; cada área mantém sua interface em um módulo de feature:

```text
src/
  app/                 shell e navegação
  features/
    markets/           cotações, gráfico e alertas
    overview/          resumo do painel
    fixed-income/      taxas e simulador
    news/              busca, filtros e notícias
  shared/
    api/               chamadas HTTP separadas por recurso
    hooks/             recursos periódicos e WebSocket
    lib/               formatação compartilhada
    types/             contratos da API
    ui/                tabela de cotações e elementos reutilizáveis
```

Novas páginas devem ficar em `features/`; estado e visual reutilizados entre áreas ficam em `shared/`. `src/api.ts` permanece como um barrel de compatibilidade para os módulos da API.

## Requisitos

- Node.js 20.19+ ou 22.12+ e npm.
- API Go respondendo em `http://localhost:8080`.

O Vite encaminha `/api` e `/health` para a API local e `/ws` para o WebSocket. As cotações usam o fluxo WebSocket com reconexão automática e polling REST como fallback (30 segundos para cripto, 60 segundos para outros mercados); taxas e notícias atualizam a cada 5 minutos.

Em produção, configure `VITE_API_BASE_URL` para a URL HTTPS pública da API. As chamadas REST e o WebSocket usam esse endereço. O `render.yaml` já fornece a variável ao compilar o site.

Para subir frontend, API e PostgreSQL em contêineres, use `docker compose up --build` na raiz do projeto e abra `http://localhost:3000`.

## Desenvolvimento (PowerShell)

Em um terminal, inicie a API conforme o README principal do projeto. Em outro:

```powershell
Set-Location D:\Projeto_Economia\front
npm ci
npm run dev
```

Abra o endereço exibido pelo Vite, normalmente `http://localhost:5173`.

O cabeçalho mostra “WebSocket conectado” quando `/ws` está ativo e “Reconectando…” durante uma queda. As cotações exibidas pelo último snapshot permanecem na tela enquanto a conexão se recupera; o polling REST serve como fallback.

## Conferir

Com a API rodando, verifique as respostas REST:

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod "http://localhost:8080/api/quotes?market=crypto"
Invoke-RestMethod http://localhost:8080/api/fixed-income/rates
Invoke-RestMethod http://localhost:8080/api/news
```

Na interface, navegue entre mercados, busque ativos por símbolo ou nome, ordene por nome/preço/variação/volume, filtre altas, baixas e acompanhados. A tabela mostra 20 ativos por página; use os controles no rodapé para navegar pelas demais páginas. Busca, filtros e ordenação são aplicados antes da paginação. Use as estrelas para montar sua lista e as setas na visão geral para organizar a ordem. O botão de alerta (♢) salva um limite para o ativo; ao atingir o preço, a interface avisa e pode enviar uma notificação do navegador.

Ao selecionar um ativo, o gráfico permite consultar 1 dia, 1 semana ou 1 mês. Ele usa amostras coletadas pela API uma vez por minuto; o histórico começa a se formar após a inicialização e só sobrevive a reinícios quando PostgreSQL está configurado. O painel de mercado mostra a disponibilidade do provedor, quantidade de ativos e horário da última atualização.

Para validar o bundle de produção:

```powershell
npm run build
npm run preview
```

O conteúdo da lista de acompanhamento fica salvo no `localStorage` do navegador. O backend ainda não oferece histórico de preços, por isso não há gráfico histórico nesta etapa.
