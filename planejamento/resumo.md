# Resumo Executivo e Checklist

## Visao Geral

O Sistema Runner e composto por:

1. **assinatura** - CLI multiplataforma que invoca o assinador.jar (modo local ou HTTP)
2. **assinador.jar** - Aplicacao Java: valida parametros FHIR, implementa `SignatureService`, simula assinatura digital
3. **simulador** - CLI multiplataforma que gerencia o ciclo de vida do simulador.jar do HubSaude
4. **Estado local** - Diretorio `~/.hubsaude/` com JDK, jars, state.json e config.json
5. **Provisionamento JDK** - Download automatico do JDK quando ausente
6. **Distribuicao** - Binarios para Windows/Linux/macOS via GitHub Releases com Cosign

---

## Checklist de Decisoes Pendentes

### Linguagem e Tooling

- [x] Linguagem do CLI: **Go** (`cobra` para parsing, `net/http` para cliente, `os/exec` para processos)
- [x] Build system do assinador.jar: **Maven com `mvnw`**
- [x] Framework HTTP do assinador.jar: **Javalin**
- [x] Versao do JDK: **JDK 21 LTS — Eclipse Temurin**
- [x] Distribuicao JDK para download: **Eclipse Temurin via api.adoptium.net**

### Parametros e Contratos

- [ ] Investigar parametros FHIR para `criar` e `validar` — ver `entregavel-assinador.md` e `entregavel-cli-assinatura.md`
- [ ] Definir forma de fornecer parametros (flags individuais, arquivo JSON, ou combinacao)
- [ ] Definir campos exatos de request/response do `SignatureController`

### Arquitetura

- [ ] Definir estrutura de pastas final do monorepo
- [x] Porta padrao do assinador.jar: **8088**, auto-deteccao no range +20 (8089, 8090...)
- [x] Endpoints do servidor: **`POST /sign`** e **`POST /validate`** (+ `/health`, `/shutdown`)
- [x] Formato de comunicacao: **JSON**
- [x] Binarios: **dois CLIs separados** (`assinatura` e `simulador`)
- [x] Modulo JDK: **duplicado** em cada CLI
- [x] Logging: `stdout` para resultado; `stderr` para progresso/erros; flags `--verbose`/`--quiet`

### CI/CD e Distribuicao

- [ ] Estrategia de empacotamento multiplataforma (.exe, .AppImage, .dmg) ou fallback para binario simples
- [ ] Configuracao do Cosign (OIDC provider, identidade)

---

## To-Do por Sprint

### Sprint 1 — Fundacao (17/03 - 30/03)
- [ ] Tomar todas as decisoes tecnicas acima
- [ ] Investigar parametros FHIR (criar e validar) e definir contrato
- [ ] Criar estrutura de projeto do CLI
- [ ] Criar estrutura de projeto do assinador.jar
- [ ] Implementar esqueleto do CLI com `--help` e subcomandos
- [ ] Implementar estrutura de `~/.hubsaude/` (estado-local)
- [ ] Configurar CI basico

### Sprint 2 — Assinador (31/03 - 13/04)
- [ ] Definir interface `SignatureService`
- [ ] Implementar `FakeSignatureService`
- [ ] Projetar esqueleto de `PKCS11SignatureService`
- [ ] Implementar `SignatureController` com `/sign`, `/validate`, `/health`, `/shutdown`
- [ ] Implementar validacao de parametros FHIR
- [ ] Implementar modo CLI do assinador.jar
- [ ] Testes unitarios (cobertura > 80%)

### Sprint 3 — Integracao (14/04 - 27/04)
- [ ] Implementar startup do CLI assinatura (ver startup.md)
- [ ] Implementar invocacao direta (modo local)
- [ ] Implementar invocacao HTTP (POST /sign, POST /validate)
- [ ] Implementar auto-deteccao de porta disponivel
- [ ] Implementar state.json (gravar/ler PID e porta)
- [ ] Implementar propagacao de erros
- [ ] Testes de integracao

### Sprint 4 — Simulador e JDK (28/04 - 11/05)
- [ ] Implementar startup do CLI simulador (ver startup.md)
- [ ] Implementar comandos: iniciar, parar, status
- [ ] Implementar flag --source para URL alternativa do simulador.jar
- [ ] Implementar download com verificacao de versao e cache
- [ ] Implementar auto-deteccao de porta para o simulador
- [ ] Implementar provisionamento automatico do JDK
- [ ] Testes multiplataforma

### Sprint 5 — Distribuicao (12/05 - 25/05)
- [ ] Cross-compilation para 3 plataformas
- [ ] Empacotamento (.exe, .AppImage, .dmg)
- [ ] Checksums SHA256
- [ ] Assinatura com Cosign
- [ ] GitHub Releases automaticas
- [ ] Testes de aceitacao

### Sprint 6 — Documentacao e Entrega (26/05 - 16/06)
- [ ] Manual do usuario
- [ ] Documentacao tecnica (incluindo SignatureService e PKCS#11)
- [ ] Guia de instalacao
- [ ] Exemplos de uso
- [ ] Release 1.0.0
