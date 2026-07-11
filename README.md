# fc-pos-expert-desafio-client-server-api

Projeto em Go com dois executáveis:

- `server/server.go`: expõe a rota `GET /cotacao` na porta `8080`
- `client/client.go`: consome o servidor e grava o valor em `cotacao.txt`

## Requisitos

- Go instalado

## Como executar

1. Inicie o servidor:

```bash
go run server/server.go
```

2. Em outro terminal, execute o cliente:

```bash
go run client/client.go
```

## Resultado

Após a execução do cliente, o arquivo `cotacao.txt` será criado na raiz do projeto com o valor da cotação retornada pelo servidor.
