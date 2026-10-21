# Backend

API REST em Go. Requer Go 1.27 ou superior.

## Executar (PowerShell)

```powershell
Set-Location D:\Projeto_Economia\back
go run ./cmd/api
```

Sem `DATABASE_URL`, as notícias ficam em memória. Para persistir em PostgreSQL, configure `DATABASE_URL` antes de iniciar. O Compose da raiz sobe a API e o banco.

O endpoint `GET /ws` publica mensagens JSON no formato `{ "type": "quotes", "market": "crypto", "data": [...], "updatedAt": "..." }`. Por padrão, a API consulta os três provedores a cada 30 segundos (`QUOTES_REFRESH_INTERVAL`) e envia também o último snapshot assim que um cliente se conecta. Os provedores são consultados por polling; isso não é um stream direto da Binance.

## Conferir os endpoints

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod "http://localhost:8080/api/quotes?market=crypto"
Invoke-RestMethod http://localhost:8080/api/fixed-income/rates
Invoke-RestMethod http://localhost:8080/api/news
```

Para inspecionar o WebSocket, abra o console do navegador e execute:

```javascript
const socket = new WebSocket("ws://localhost:8080/ws");
socket.onmessage = (event) => console.log(JSON.parse(event.data));
```

As rotas de simulação e filtros de notícias estão disponíveis em `/api/fixed-income/simulate` e `/api/news`.
