# Frontend do Painel de Mercado

## Requisitos

- Node.js 20.19+ ou 22.12+ e npm.
- API Go respondendo em `http://localhost:8080`.

O Vite encaminha `/api` e `/health` para a API local. O projeto usa polling: 30 segundos para cripto, 60 segundos para outros mercados e 5 minutos para taxas e notícias.

Para subir frontend, API e PostgreSQL em contêineres, use `docker compose up --build` na raiz do projeto e abra `http://localhost:3000`.

## Desenvolvimento (PowerShell)

Em um terminal, inicie a API conforme o README principal do projeto. Em outro:

```powershell
Set-Location D:\Projeto_Economia\front
npm install
npm run dev
```

Abra o endereço exibido pelo Vite, normalmente `http://localhost:5173`.

## Conferir

Com a API rodando, verifique as respostas REST:

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod "http://localhost:8080/api/quotes?market=crypto"
Invoke-RestMethod http://localhost:8080/api/fixed-income/rates
Invoke-RestMethod http://localhost:8080/api/news
```

Na interface, navegue entre mercados, salve um ativo pela estrela para vê-lo em “Acompanhe”, filtre notícias e envie uma simulação de renda fixa. Para validar o bundle de produção:

```powershell
npm run build
npm run preview
```

O conteúdo da lista de acompanhamento fica salvo no `localStorage` do navegador. O backend ainda não oferece histórico de preços, por isso não há gráfico histórico nesta etapa.
