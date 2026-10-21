# Backend

API REST em Go. Requer Go 1.27 ou superior.

## Executar (PowerShell)

```powershell
Set-Location D:\Projeto_Economia\back
go run ./cmd/api
```

Sem `DATABASE_URL`, notícias e histórico de cotações ficam em memória e são perdidos ao reiniciar a API. Para persistir ambos em PostgreSQL, configure `DATABASE_URL` antes de iniciar. O Compose da raiz sobe a API e o banco.

O servidor usa `PORT` quando `HTTP_ADDR` não está definido (Render fornece essa variável e exige que o serviço escute nela). No deploy Render, `FRONTEND_ORIGIN` deve conter a URL HTTPS do site estático; ela libera chamadas REST e o WebSocket apenas para esse frontend.

O endpoint `GET /ws` publica mensagens JSON no formato `{ "type": "quotes", "market": "crypto", "data": [...], "updatedAt": "..." }`. Por padrão, a API consulta os três provedores a cada 30 segundos (`QUOTES_REFRESH_INTERVAL`) e envia também o último snapshot assim que um cliente se conecta. Os provedores são consultados por polling; isso não é um stream direto da Binance.

O provedor Binance lista por padrão 20 pares contra USDT. Você pode substituir a lista com `CRYPTO_SYMBOLS`, por exemplo `BTCUSDT,ETHUSDT,DOGEUSDT`. Para ações, a API pública da brapi fornece 50 ações ordenadas por volume sem chave. Defina `STOCK_SYMBOLS` para pedir tickers específicos; essa seleção usa a chave `BRAPI_TOKEN` quando configurada.

## Conferir os endpoints

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod "http://localhost:8080/api/quotes?market=crypto"
Invoke-RestMethod "http://localhost:8080/api/quotes/history?symbol=BTCUSDT&range=1d"
Invoke-RestMethod http://localhost:8080/api/fixed-income/rates
Invoke-RestMethod http://localhost:8080/api/news
```

Para inspecionar o WebSocket, abra o console do navegador e execute:

```javascript
const socket = new WebSocket("ws://localhost:8080/ws");
socket.onmessage = (event) => console.log(JSON.parse(event.data));
```

As rotas de simulação e filtros de notícias estão disponíveis em `/api/fixed-income/simulate` e `/api/news`.

O histórico é gravado uma vez por minuto para cada ativo recebido. Consulte períodos `1d`, `1w` ou `1m`; os pontos começam a acumular depois que a API é iniciada e não são importados dos provedores.
