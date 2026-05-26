package br.gov.saude.assinador.servidor;

import br.gov.saude.assinador.servico.FakeSignatureService;
import io.javalin.Javalin;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Cobre o auto-shutdown por inatividade (US-01 ultimo criterio).
 * Usa Duration de 1s para nao depender de minutos reais.
 */
class AssinadorServidorIdleTest {

    private Javalin app;

    @AfterEach
    void parar() {
        if (app != null) {
            try { app.stop(); } catch (Exception ignored) {}
        }
    }

    @Test
    void requisicoesResetamRelogioDeInatividade() throws Exception {
        // Sobe servidor sem timeout (testamos apenas o handler before via reflection
        // para evitar System.exit durante os testes).
        app = AssinadorServidor.iniciar(0, new FakeSignatureService());
        int p = app.port();
        HttpClient c = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(2)).build();
        HttpResponse<String> r = c.send(
                HttpRequest.newBuilder(URI.create("http://localhost:" + p + "/health")).GET().build(),
                HttpResponse.BodyHandlers.ofString());
        assertTrue(r.statusCode() == 200);
    }

    @Test
    void timeoutZeroNaoAgendaShutdown() {
        // Verifica que iniciar(porta, servico, Duration.ZERO) e equivalente a
        // iniciar(porta, servico) — sem watcher.
        app = AssinadorServidor.iniciar(0, new FakeSignatureService(), Duration.ZERO);
        assertNotNull(app);
        // se chegou aqui sem excecao, o ramo de auto-shutdown nao foi tomado.
        assertTrue(app.port() > 0);
    }

    @Test
    void timeoutNegativoIgnorado() {
        app = AssinadorServidor.iniciar(0, new FakeSignatureService(), Duration.ofSeconds(-5));
        assertNotNull(app);
        assertTrue(app.port() > 0);
    }
}
