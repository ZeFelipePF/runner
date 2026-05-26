# Requisitos Não-Funcionais (NFRs) — Sistema Runner

Documento que responde à crítica do `runnerProfessor/docs/planejamento.md`:

> "Há ênfase em requisitos funcionais, nenhum requisito não funcional foi
> fornecido. Por exemplo, se observamos a norma ISO 25010, veremos que não há
> requisitos de desempenho, consumo de recursos, portabilidade e outros."

Os NFRs abaixo seguem a [ISO/IEC 25010:2011](https://iso25000.com/index.php/en/iso-25000-standards/iso-25010)
(características de qualidade) e são escritos no formato **SMART** (Specific,
Measurable, Achievable, Relevant, Time-bound) — cada NFR tem uma métrica
verificável.

---

## 1. Performance (Desempenho)

### NFR-PERF-01 — Cold start (modo local)
| Atributo | Valor |
|---|---|
| **Métrica** | Tempo entre `assinatura criar` ser invocado e a resposta JSON ser emitida |
| **Alvo** | < 3 segundos em hardware razoável (CPU ≥ 2 cores, RAM ≥ 4GB, SSD) |
| **Justificativa** | Cold start envolve `os/exec` → JVM startup + carga do jar. Aceitável para uso esporádico. |
| **Como medir** | `time assinatura criar --modo local --payload pedido.json` |

### NFR-PERF-02 — Warm start (modo HTTP)
| Atributo | Valor |
|---|---|
| **Métrica** | Latência p95 de `POST /sign` com servidor já rodando |
| **Alvo** | < 200ms |
| **Justificativa** | Modo servidor elimina overhead da JVM; usuário percebe resposta imediata. |
| **Como medir** | `ab -n 100 -c 4 -p pedido.json http://localhost:8088/sign` (ou `wrk`) |

### NFR-PERF-03 — Startup do simulador
| Atributo | Valor |
|---|---|
| **Métrica** | Tempo entre `simulador iniciar` e `/actuator/health` retornar 200 |
| **Alvo** | < 30s incluindo download do JRE em primeira execução; < 10s nas subsequentes |
| **Justificativa** | Dependência do JVM startup do Spring + carregamento do jar. |

---

## 2. Portabilidade

### NFR-PORT-01 — Plataformas suportadas
| Atributo | Valor |
|---|---|
| **Alvo** | Linux amd64, Windows amd64, macOS amd64 |
| **Verificação** | Job `acceptance` em matrix 3 SOs no `ci.yml` deve passar em toda a `main` |

### NFR-PORT-02 — Provisionamento sem root
| Atributo | Valor |
|---|---|
| **Alvo** | JRE 21 baixado e extraído em `~/.hubsaude/jdk/` — sem `sudo`/admin |
| **Justificativa** | Usuário sem privilégios deve conseguir usar o sistema. |

### NFR-PORT-03 — Binário autocontido
| Atributo | Valor |
|---|---|
| **Alvo** | Binários Go estaticamente lincados, sem dependências de bibliotecas C externas |
| **Verificação** | `ldd assinatura` (Linux) retorna "not a dynamic executable" |

---

## 3. Segurança

### NFR-SEG-01 — Assinatura de artefatos
| Atributo | Valor |
|---|---|
| **Alvo** | 100% dos artefatos da release têm `.sig` + `.pem` Cosign (Sigstore OIDC) |
| **Verificação** | `cosign verify-blob --certificate <art>.pem --signature <art>.sig --certificate-identity-regexp '^https://github.com/.*/runner/.*' --certificate-oidc-issuer https://token.actions.githubusercontent.com <art>` retorna "Verified OK" |

### NFR-SEG-02 — Isolamento da chave privada (PKCS#11)
| Atributo | Valor |
|---|---|
| **Alvo** | Quando provider PKCS#11 ativo, chave privada **nunca** trafega na memória do CLI ou do jar — operações ocorrem dentro do dispositivo |
| **Justificativa** | Garantia básica de tokens A3 ICP-Brasil. |

### NFR-SEG-03 — Sem segredos em código ou logs
| Atributo | Valor |
|---|---|
| **Alvo** | PINs PKCS#11 lidos apenas de stdin ou arquivo; nunca em argv. Logs nunca incluem chave privada ou PIN. |
| **Verificação** | `grep -r "PIN\|password\|senha" assinador/src` mostra apenas referências em nomes de variáveis |

### NFR-SEG-04 — Validação de integridade de downloads
| Atributo | Valor |
|---|---|
| **Alvo** | Toda artefato baixado (JRE, simulador.jar) tem SHA-256 verificado contra valor esperado |
| **Verificação** | Tests `TestBaixarAdoptium_ChecksumDivergente` e `TestDownload_RevalidaChecksum` |

---

## 4. Confiabilidade

### NFR-CONF-01 — Auto-detecção de porta livre
| Atributo | Valor |
|---|---|
| **Alvo** | Se porta padrão (`8088` assinador / `8443` simulador) estiver ocupada, tentar as próximas até +20 |
| **Verificação** | Test `TestEscolherPorta_AutoDetectaProxima` |

### NFR-CONF-02 — Recuperação de PID stale
| Atributo | Valor |
|---|---|
| **Alvo** | Se `state.json` registra PID não-vivo, considerar inativo e iniciar nova instância |
| **Verificação** | Test `TestIniciar_DescartaPidObsoleto` |

### NFR-CONF-03 — Shutdown graceful
| Atributo | Valor |
|---|---|
| **Alvo** | `assinatura servidor parar` tenta HTTP shutdown → SIGTERM → SIGKILL nesta ordem; reporta método usado |
| **Verificação** | Tests em `processo_test.go` (parar nao-rodando, stale) |

### NFR-CONF-04 — Idempotência de start
| Atributo | Valor |
|---|---|
| **Alvo** | `servidor iniciar` chamado duas vezes não cria duas instâncias — detecta via `state.json + PID + /health` |
| **Verificação** | Test `TestIniciar_ReusaInstanciaExistente` |

---

## 5. Manutenibilidade

### NFR-MAN-01 — Cobertura de testes
| Atributo | Valor |
|---|---|
| **Alvo** | Go: > 80% nos pacotes `internal/*`. Java: > 80% (excluindo `PKCS11SignatureService` quando sem SoftHSM2). |
| **Verificação** | `go test -cover ./...` e relatório JaCoCo |

### NFR-MAN-02 — CI matrix 3 SOs
| Atributo | Valor |
|---|---|
| **Alvo** | Toda PR para `main` executa build + test + acceptance em Ubuntu/Windows/macOS |
| **Verificação** | `.github/workflows/ci.yml` |

### NFR-MAN-03 — Logging estruturado
| Atributo | Valor |
|---|---|
| **Alvo** | Go usa `slog`; Java usa Logback + Logstash encoder (JSON). Mensagens incluem timestamp, nível, módulo. |

---

## 6. Observabilidade

### NFR-OBS-01 — OpenTelemetry bridge
| Atributo | Valor |
|---|---|
| **Alvo** | Logs Go e Java exportáveis via OTel quando env `OTEL_EXPORTER_*` definidas |
| **Justificativa** | Permite integração futura com Grafana/Loki/Tempo sem mudanças de código |

### NFR-OBS-02 — Endpoint /health e /api/info
| Atributo | Valor |
|---|---|
| **Alvo** | `assinador.jar` expõe `/health` (200 quando vivo) e o simulador é consultado em `/api/info` pelo CLI |
| **Verificação** | Testes `SignatureControllerTest` e `TestConsultar_ApiInfoPopulado` |

---

## 7. Usabilidade

### NFR-USA-01 — Help integrado
| Atributo | Valor |
|---|---|
| **Alvo** | Todos os comandos e subcomandos têm `--help` listando flags com descrição e defaults |
| **Verificação** | `assinatura --help`, `assinatura criar --help`, etc. (testes de aceitação) |

### NFR-USA-02 — Mensagens de erro estruturadas
| Atributo | Valor |
|---|---|
| **Alvo** | Erros emitem JSON em stderr com `{error: <CODIGO>, message: <descricao>}`; exit code mapeado por código (PARAM_AUSENTE=2, ...) |
| **Verificação** | Tests em `cmd/exit_test.go` + `MapeadorErroTest.java` |

### NFR-USA-03 — Separação stdout/stderr
| Atributo | Valor |
|---|---|
| **Alvo** | stdout = resultado pipe-friendly (JSON); stderr = progresso/erros operacionais; flag `--quiet` redireciona progresso para `io.Discard` |
| **Justificativa** | Permite uso em scripts (`assinatura criar ... | jq .`). |

---

## 8. Compatibilidade

### NFR-COMP-01 — Versionamento semântico
| Atributo | Valor |
|---|---|
| **Alvo** | Releases seguem SemVer (`v<MAJOR>.<MINOR>.<PATCH>`); breaking changes só em major bumps |

### NFR-COMP-02 — Backward compatibility de `state.json`
| Atributo | Valor |
|---|---|
| **Alvo** | Versões futuras leem `state.json` antigos sem quebrar; campos novos são opcionais |

---

## Rastreabilidade

| US (spec) | NFRs relacionados |
|-----------|-------------------|
| US-01 | NFR-PERF-01, NFR-PERF-02, NFR-CONF-01, NFR-CONF-03, NFR-CONF-04, NFR-USA-01..03 |
| US-02 | NFR-SEG-02, NFR-SEG-03, NFR-USA-02 |
| US-03 | NFR-PERF-03, NFR-CONF-01, NFR-OBS-02, NFR-PORT-01 |
| US-04 | NFR-PORT-01, NFR-PORT-02, NFR-SEG-04 |
| US-05 | NFR-SEG-01, NFR-PORT-01, NFR-PORT-03, NFR-COMP-01 |

---

## Como validar todos os NFRs

```bash
# Performance: rodar manualmente com `time`
# Portabilidade: CI matrix 3 SOs
# Segurança: cosign verify (NFR-SEG-01); testes unitários (SEG-04)
# Confiabilidade: go test ./... -tags=integration
# Manutenibilidade: go test -cover; mvn test
# Observabilidade: testes funcionais; instrumentação manual
# Usabilidade: testes de aceitação (-tags=acceptance)
```
