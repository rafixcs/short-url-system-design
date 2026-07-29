# Shorter URL Design System

API REST escrita em Go para criação de usuários e encurtamento de URLs. A
aplicação usa Chi como roteador HTTP e está organizada em camadas inspiradas em
Clean Architecture.

## Funcionalidades implementadas

- verificação de disponibilidade da API;
- criação e consulta de usuários;
- autenticação por e-mail ou nome de usuário com token JWT;
- proteção de endpoints por middleware `Authorization: Bearer`;
- armazenamento de senhas com bcrypt;
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
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) para tokens JWT
- [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
  para hash e verificação de senhas
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

A porta e a chave usada para assinar os JWTs podem ser configuradas pelas
variáveis `PORT` e `JWT_SECRET`:

```bash
PORT=9090 JWT_SECRET='use-a-long-random-secret' go run ./src/cmd
```

Quando `JWT_SECRET` não é informada, a aplicação usa uma chave apenas para
desenvolvimento. Defina obrigatoriamente uma chave forte fora do ambiente local.

## Rotas

| Método | Rota | Descrição | Sucesso |
| --- | --- | --- | --- |
| `GET` | `/status` | Verifica se a API está disponível | `200` |
| `POST` | `/api/v1/users/` | Cria um usuário | `201` |
| `POST` | `/api/v1/auth/login` | Autentica e retorna um JWT | `200` |
| `GET` | `/api/v1/users/{userID}` | Consulta um usuário (protegida) | `200` |
| `POST` | `/api/v1/urls/` | Cria uma URL curta (protegida) | `201` |
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

A senha é armazenada como hash bcrypt e não é incluída na resposta HTTP.

### Login

O login aceita o e-mail:

```bash
curl --request POST \
  --url http://localhost:8080/api/v1/auth/login \
  --header 'Content-Type: application/json' \
  --data '{
    "email": "rafael@example.com",
    "password": "change-me-123"
  }'
```

Ou o nome de usuário criado no campo `name`:

```json
{
  "user": "Rafael Silva",
  "password": "change-me-123"
}
```

Resposta `200 OK`:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_at": "2026-07-30T19:22:13-03:00"
}
```

O token usa HS256 e expira após 24 horas. Credenciais inválidas retornam
`401 Unauthorized`.

### Autorização

Envie o token nas rotas protegidas usando o header:

```text
Authorization: Bearer <access_token>
```

Headers ausentes, malformados, tokens inválidos ou expirados retornam
`401 Unauthorized`.

### Consultar usuário

Substitua `{userID}` pelo ID retornado na criação:

```bash
curl \
  --header 'Authorization: Bearer <access_token>' \
  http://localhost:8080/api/v1/users/688a9d8dd53426a463c454ab
```

Quando o usuário não existe, a API responde com `404 Not Found`.

### Criar URL curta

Use o ID de um usuário criado anteriormente:

```bash
curl --request POST \
  --url http://localhost:8080/api/v1/urls/ \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer <access_token>' \
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

Importe o arquivo no Postman e execute as requisições nesta ordem:

1. **Create User** salva o ID retornado em `user_id`;
2. **Login** salva o JWT retornado em `auth_token`;
3. **Get User** e **Create Short URL** enviam automaticamente o Bearer token.

As variáveis `base_url`, `user_id` e `auth_token` ficam no escopo da coleção.
`base_url` possui o valor padrão `http://localhost:8080`.

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

A configuração inicia `src/cmd` com `PORT=8080`, define uma `JWT_SECRET` de
desenvolvimento e permite o uso de breakpoints por meio do Delve.

## Verificação do projeto

Execute todos os testes e a análise estática:

```bash
go test ./...
go vet ./...
```

O teste E2E em `src/internal/server/router_e2e_test.go` inicia um servidor HTTP
temporário e valida o fluxo completo:

- status público;
- criação de usuário sem exposição da senha;
- bloqueio de rota protegida sem token;
- rejeição de credenciais inválidas;
- login por e-mail e por nome de usuário;
- consulta de usuário com Bearer token;
- criação de URL curta com Bearer token;
- redirecionamento `302` para a URL original.

Para executar apenas o teste E2E:

```bash
go test ./src/internal/server -run TestAuthenticationAndShortURLFlow -v
```

## Pipeline CI/CD e GitFlow

A pipeline do GitHub Actions está em:

```text
.github/workflows/gitflow-ci-cd.yml
```

Ela é executada em pushes para:

- `feature/*` e `feature_*`;
- `bugfix/*` e `bugfix_*`;
- `develop`;
- `release/*` e `release_*`;
- `hotfix/*` e `hotfix_*`;
- `main`;
- tags iniciadas por `v`, como `v1.0.0`.

Pull requests destinados a `develop` ou `main` também disparam a pipeline. A
política GitFlow aceita as seguintes transições:

| Destino | Origens permitidas |
| --- | --- |
| `develop` | `feature/*`, `bugfix/*`, `release/*` e `hotfix/*` |
| `main` | `release/*` e `hotfix/*` |

As variações com `_`, como a branch `feature_auth`, também são aceitas para
compatibilidade com as branches já usadas neste repositório.

### Etapas da pipeline

1. **GitFlow policy:** valida a origem e o destino do pull request.
2. **Quality:** verifica dependências, `gofmt`, `go vet`, race conditions, testes
   e cobertura.
3. **Build:** gera o binário Linux `shorter-url-api` e o disponibiliza como
   artefato por 14 dias.
4. **Container:** constrói o estágio `prd` de `infra/service.Dockerfile` sem
   publicar a imagem.
5. **Publish:** depois de todas as validações, publica a imagem no GitHub
   Container Registry somente para pushes em `main` e tags `v*`.

A imagem publicada segue este formato:

```text
ghcr.io/<owner>/<repository>:<tag>
```

Um push em `main` produz, entre outras, as tags `main`, `latest` e `sha-*`. Uma
tag Git `v1.2.3` produz tags semânticas como `1.2.3` e `1.2`.

O job de publicação usa o `GITHUB_TOKEN` fornecido pelo próprio Actions e pede
somente `contents: read` e `packages: write`. Confirme em **Settings → Actions →
General → Workflow permissions** que o GitHub Actions pode publicar packages.

Esta etapa de CD entrega uma imagem versionada no GHCR. Um deploy para servidor,
Kubernetes ou provedor de nuvem não foi incluído porque o projeto ainda não
define um ambiente de hospedagem ou credenciais de infraestrutura.

### Proteção recomendada das branches

No GitHub, configure regras para `develop` e `main` exigindo pull request e os
checks da pipeline antes do merge. Para produção, também é recomendável criar
um environment protegido e exigir aprovação antes da publicação ou de um futuro
job de deploy.

## Limitações atuais

- Os dados são armazenados somente em memória e são perdidos ao reiniciar a
  aplicação.
- O arquivo `mongo.go` ainda não implementa persistência no MongoDB.
- Ainda não há validação completa de nome, e-mail e senha.
- Não há refresh token, revogação de tokens ou autorização por papéis.
- A chave JWT padrão é adequada apenas ao desenvolvimento e deve ser substituída
  pela variável `JWT_SECRET` em outros ambientes.
- A criação de URL não verifica atualmente se o `user_id` informado existe.
- O fluxo principal possui cobertura E2E, mas ainda faltam testes unitários para
  os serviços, repositórios, handlers e cenários de borda.
