# Painel de Mercado

O repositório está dividido em duas aplicações independentes:

- `back/`: API REST em Go.
- `front/`: interface React + TypeScript com Vite.

## Executar localmente

Inicie a API em um terminal:

```powershell
Set-Location D:\Projeto_Economia\back
go run ./cmd/api
```

Sem `DATABASE_URL`, a API mantém as notícias em memória. Em outro terminal, inicie a interface:

```powershell
Set-Location D:\Projeto_Economia\front
npm ci
npm run dev
```

O Vite encaminha chamadas `/api` para `http://localhost:8080`. Consulte `back/README.md` e `front/README.md` para configuração e validação.

## Docker Compose

Na raiz do projeto:

```powershell
docker compose up --build
```

O Compose inicia PostgreSQL, API e frontend. Abra `http://localhost:3000`; as chamadas `/api`, `/health` e `/ws` passam pelo proxy do frontend até a API. A API continua exposta em `http://localhost:8080` para chamadas diretas. O WebSocket envia snapshots de cotações consultados periodicamente nos provedores.

O padrão inclui 20 pares de cripto contra USDT e 50 ações ordenadas por volume via lista pública da brapi. `CRYPTO_SYMBOLS` e `STOCK_SYMBOLS` permitem personalizar as listas; a chave `BRAPI_TOKEN` é usada para buscar tickers específicos.
