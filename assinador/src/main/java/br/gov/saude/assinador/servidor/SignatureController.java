package br.gov.saude.assinador.servidor;

import br.gov.saude.assinador.cli.AcaoAssinar;
import br.gov.saude.assinador.cli.AcaoValidar;
import br.gov.saude.assinador.erro.MapeadorErro;
import br.gov.saude.assinador.erro.RespostaErro;
import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import br.gov.saude.assinador.servico.SignatureService;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.javalin.Javalin;
import io.javalin.http.Context;
import io.javalin.http.HttpStatus;

import java.time.Instant;
import java.util.Map;

/**
 * Expoe operacoes de assinatura via HTTP usando Javalin.
 * Endpoints: POST /sign, POST /validate, GET /health, POST /shutdown.
 */
public final class SignatureController {

    private static final TypeReference<Map<String, Object>> TIPO_PAYLOAD = new TypeReference<>() {};

    private final SignatureService servico;
    private final Runnable aoDesligar;
    private final ObjectMapper json = new ObjectMapper();
    private final Instant inicio = Instant.now();

    public SignatureController(SignatureService servico) {
        // Producao: POST /shutdown encerra o processo (o CLI depende disso para
        // parar o assinador.jar em modo servidor).
        this(servico, () -> System.exit(0));
    }

    /**
     * Permite injetar a acao de desligamento. Testes passam um no-op para
     * verificar o endpoint {@code /shutdown} sem encerrar a JVM do runner de testes.
     */
    public SignatureController(SignatureService servico, Runnable aoDesligar) {
        this.servico = servico;
        this.aoDesligar = aoDesligar;
    }

    public void registrar(Javalin app) {
        app.post("/sign", this::sign);
        app.post("/validate", this::validate);
        app.get("/health", this::health);
        app.post("/shutdown", this::shutdown);

        app.exception(AssinadorException.class, (ex, ctx) -> responderErro(ctx, ex));
        app.exception(Exception.class, (ex, ctx) -> {
            AssinadorException convertido = new AssinadorException(
                    Codigo.ERRO_INTERNO,
                    ex.getMessage() == null ? "erro interno" : ex.getMessage());
            responderErro(ctx, convertido);
        });
    }

    private void sign(Context ctx) {
        Map<String, Object> payload = lerPayload(ctx);
        Map<String, Object> resposta = new AcaoAssinar(servico).executar(payload);
        ctx.contentType("application/json").json(resposta);
    }

    private void validate(Context ctx) {
        Map<String, Object> payload = lerPayload(ctx);
        Map<String, Object> resposta = new AcaoValidar(servico).executar(payload);
        ctx.contentType("application/json").json(resposta);
    }

    private void health(Context ctx) {
        ctx.json(Map.of(
                "status", "UP",
                "iniciadoEm", inicio.toString(),
                "uptimeSegundos", Instant.now().getEpochSecond() - inicio.getEpochSecond()));
    }

    private void shutdown(Context ctx) {
        ctx.json(Map.of("status", "SHUTTING_DOWN"));
        Thread t = new Thread(() -> {
            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
            aoDesligar.run();
        }, "assinador-shutdown");
        t.setDaemon(true);
        t.start();
    }

    private Map<String, Object> lerPayload(Context ctx) {
        String body = ctx.body();
        if (body == null || body.isBlank()) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE, "corpo da requisicao vazio");
        }
        try {
            Map<String, Object> payload = json.readValue(body, TIPO_PAYLOAD);
            if (payload == null) {
                throw new AssinadorException(Codigo.PARAM_AUSENTE, "payload vazio");
            }
            return payload;
        } catch (com.fasterxml.jackson.core.JsonProcessingException ex) {
            throw new AssinadorException(Codigo.PARAM_INVALIDO,
                    "JSON invalido: " + ex.getOriginalMessage());
        }
    }

    private void responderErro(Context ctx, AssinadorException ex) {
        Codigo codigo = ex.getCodigo() == null ? Codigo.ERRO_INTERNO : ex.getCodigo();
        int status = MapeadorErro.httpStatus(codigo);
        ctx.status(HttpStatus.forStatus(status));
        ctx.contentType("application/json").json(RespostaErro.de(ex));
    }
}
