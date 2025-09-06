# Sistema de Temperatura por CEP com OTEL(Open Telemetry) e Zipkin

Dois microserviços que trabalham em conjunto para fornecer informações de temperatura com base em um CEP fornecido. utiliza OpenTelemetry para rastreamento distribuído e Zipkin para visualização dos traces.

## Serviços

### Serviço A (Input Service)
- Responsável por receber e validar o CEP
- Porta: 8080
- Endpoint: POST /
- Payload: `{ "cep": "29902555" }`

### Serviço B (Temperature Service)
- Responsável por buscar o endereço e a temperatura
- Porta: 8081
- Endpoint: POST /temperature
- Integração com ViaCEP e WeatherAPI

## Requisitos

- Docker e Docker Compose
- Chave de API do WeatherAPI (https://www.weatherapi.com/)

## Como Executar

1. Clone o repositório
2. Crie um arquivo `.env` na raiz do projeto com sua chave do WeatherAPI:
   ```
   WEATHER_API_KEY=sua_chave_aqui
   ```

3. Execute o projeto com Docker Compose:
   ```bash
   docker-compose up -d
   ```

servicos:
   - http://localhost:8080 - servico A (input-service)
   - http://localhost:8081 - servico B (temperature-service)
   - http://localhost:9411 - Zipkin

## Como usar:

### Requisição Válida
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'

```

Resposta de Sucesso (200 OK):
```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

#### CEP Inválido (422 Unprocessable Entity)
```json
{
  "error": "invalid zipcode"
}
```

#### CEP Não Encontrado (404 Not Found)
```json
{
  "error": "can not find zipcode"
}
```

## Monitoramento

O projeto utiliza OpenTelemetry para rastreamento distribuído e Zipkin para visualização dos traces. Você pode acessar o Zipkin em http://localhost:9411 para visualizar os traces dos serviços.
