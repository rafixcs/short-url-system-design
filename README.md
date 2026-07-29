# Shorter URL Design System

API REST escrita em Go para criação de usuários e encurtamento de URLs. A
aplicação usa Chi como roteador HTTP e está organizada em camadas inspiradas em
Clean Architecture.

## Funcionalidades implementadas

- verificação de disponibilidade da API;
- criação e consulta de usuários;
- criação de códigos curtos com seis caracteres;
- validação básica da URL original;
- redirecionamento do código curto para a URL original;
- repositório concorrente em memória protegido por `sync.RWMutex`;
- injeção de dependências entre repositório, serviços e handlers;
- middlewares de log, request ID, recuperação de panic e CORS;
- coleção Postman com variáveis e testes automáticos;
- configuração de debug para VS Code e Delve.

## Tecnologias

- Go 1.26.5
- [Chi](https://github.com/go-chi/chi) para roteamento HTTP
- [go-chi/cors](https://github.com/go-chi/cors) para CORS
- MongoDB Driver, reservado para a futura implementação do repositório MongoDB

## Arquitetura

O fluxo das dependências é:

```text
HTTP request
    ↓
Router (Chi)
    ↓
Handler HTTP
    ↓
Service (regra de negócio)
    ↓
Repository interface
    ↓
In-memory repository
```

As interfaces são definidas na camada de domínio. A composição concreta das
dependências acontece no servidor:

```text
src/
├── cmd/
│   └── main.go                         # ponto de entrada
├── internal/
│   ├── domain/                         # modelos e contratos
│   ├── service/                        # casos de uso
│   ├── infrastructure/
│   │   ├── http/                       # handlers e DTOs HTTP
│   │   └── repository/                 # implementações de persistência
│   └── server/                         # servidor, composição e rotas
└── pkg/
    ├── env/                            # leitura de variáveis de ambiente
    └── utils/                          # utilitários HTTP
```

O servidor cria uma única instância do repositório em memória e a compartilha
entre os serviços de usuários e URLs. Os handlers recebem os serviços por
construtor e dependem das interfaces do domínio.

## Executando localmente

### Pré-requisitos

- Go 1.26.5 ou versão compatível com o `go.mod`.

Baixe as dependências e execute a aplicação:

```bash
go mod download
go run ./src/cmd
```

Por padrão, a API estará disponível em:

```text
http://localhost:8080
```

A porta pode ser alterada com a variável `PORT`:

```bash
PORT=9090 go run ./src/cmd
```

## Rotas

| Método | Rota | Descrição | Sucesso |
| --- | --- | --- | --- |
| `GET` | `/status` | Verifica se a API está disponível | `200` |
| `POST` | `/api/v1/users/` | Cria um usuário | `201` |
| `GET` | `/api/v1/users/{userID}` | Consulta um usuário | `200` |
| `POST` | `/api/v1/urls/` | Cria uma URL curta | `201` |
| `GET` | `/{shortURL}` | Redireciona para a URL original | `302` |

As barras finais nas rotas de criação fazem parte das rotas atualmente
registradas no Chi.

### Status da API

```bash
curl http://localhost:8080/status
```

Resposta:

```json
{
  "status": "its alive!"
}
```

### Criar usuário

```bash
curl --request POST \
  --url http://localhost:8080/api/v1/users/ \
  --header 'Content-Type: application/json' \
  --data '{
    "name": "Rafael Silva",
    "email": "rafael@example.com",
    "password": "change-me-123"
  }'
```

Resposta `201 Created`:

```json
{
  "id": "688a9d8dd53426a463c454ab",
  "name": "Rafael Silva",
  "email": "rafael@example.com"
}
```

A senha não é incluída na resposta HTTP.

### Consultar usuário

Substitua `{userID}` pelo ID retornado na criação:

```bash
curl http://localhost:8080/api/v1/users/688a9d8dd53426a463c454ab
```

Quando o usuário não existe, a API responde com `404 Not Found`.

### Criar URL curta

Use o ID de um usuário criado anteriormente:

```bash
curl --request POST \
  --url http://localhost:8080/api/v1/urls/ \
  --header 'Content-Type: application/json' \
  --data '{
    "user_id": "688a9d8dd53426a463c454ab",
    "long_url": "https://example.com/articles/clean-architecture"
  }'
```

Resposta `201 Created`:

```json
{
  "user_id": "688a9d8dd53426a463c454ab",
  "long_url": "https://example.com/articles/clean-architecture",
  "shor_url": "B2OXy0"
}
```

> O campo de resposta está atualmente implementado como `shor_url`. A correção
> para `short_url` deve ser tratada como uma alteração de contrato da API.

A URL original deve ser absoluta e conter esquema e host, por exemplo
`https://example.com/page`.

### Acessar a URL curta

Abra o endereço no navegador:

```text
http://localhost:8080/B2OXy0
```

Ou inspecione o redirecionamento sem segui-lo:

```bash
curl --include http://localhost:8080/B2OXy0
```

A resposta será `302 Found`, com a URL original no cabeçalho `Location`. Para
fazer o `curl` seguir o redirecionamento, use:

```bash
curl --location http://localhost:8080/B2OXy0
```

## Postman

A coleção está disponível em:

```text
postman/shorter-url-api.postman_collection.json
```

Importe o arquivo no Postman e execute primeiro a requisição **Create User**.
O script da coleção salva automaticamente o ID retornado na variável
`user_id`, utilizada pelas requisições **Get User** e **Create Short URL**.

A variável `base_url` possui o valor padrão `http://localhost:8080`.

## Debug no VS Code

O projeto contém a configuração:

```text
.vscode/launch.json
```

Para iniciar o debug:

1. instale a extensão oficial de Go no VS Code;
2. abra **Run and Debug**;
3. selecione **Debug Shorter URL API**;
4. pressione `F5`.

A configuração inicia `src/cmd` com `PORT=8080` e permite o uso de breakpoints
por meio do Delve.

## Verificação do projeto

Execute todos os testes e a análise estática:

```bash
go test ./...
go vet ./...
```

Ainda não existem arquivos de teste automatizado no repositório; atualmente,
esses comandos verificam a compilação e problemas detectáveis pelo `go vet`.

## Limitações atuais

- Os dados são armazenados somente em memória e são perdidos ao reiniciar a
  aplicação.
- O arquivo `mongo.go` ainda não implementa persistência no MongoDB.
- A senha ainda é armazenada sem hash no modelo em memória. Isso não é adequado
  para produção; uma implementação real deve aplicar um algoritmo próprio para
  senhas, como bcrypt ou Argon2.
- Ainda não há validação completa de nome, e-mail e senha.
- Ainda não há autenticação ou autorização.
- A criação de URL não verifica atualmente se o `user_id` informado existe.
- Ainda não há testes unitários ou de integração.
