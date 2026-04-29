package br.gov.saude.assinador.servidor;

import br.gov.saude.assinador.servico.FakeSignatureService;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.javalin.Javalin;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.http.HttpResponse.BodyHandlers;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Base64;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

class SignatureControllerTest {

    private static final ObjectMapper JSON = new ObjectMapper();
    private Javalin app;
    private HttpClient client;
    private int porta;

    @BeforeEach
    void iniciar() {
        app = AssinadorServidor.iniciar(0, new FakeSignatureService());
        porta = app.port();
        client = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(2)).build();
    }

    @AfterEach
    void parar() {
        if (app != null) app.stop();
    }

    @Test
    void healthRetornaUp() throws Exception {
        HttpResponse<String> r = get("/health");
        assertEquals(200, r.statusCode());
        Map<String, Object> corpo = parse(r.body());
        assertEquals("UP", corpo.get("status"));
        assertNotNull(corpo.get("iniciadoEm"));
    }

    @Test
    void signRetornaSignatureFhir() throws Exception {
        String payload = JSON.writeValueAsString(payloadValidoCriacao());
        HttpResponse<String> r = post("/sign", payload);
        assertEquals(200, r.statusCode());
        Map<String, Object> corpo = parse(r.body());
        assertEquals("Signature", corpo.get("resourceType"));
        assertNotNull(corpo.get("data"));
    }

    @Test
    void signComPayloadInvalidoRetorna400() throws Exception {
        HttpResponse<String> r = post("/sign", "{}");
        assertEquals(400, r.statusCode());
        Map<String, Object> corpo = parse(r.body());
        assertEquals("PARAM_AUSENTE", corpo.get("error"));
    }

    @Test
    void signComJsonMalformadoRetorna400() throws Exception {
        HttpResponse<String> r = post("/sign", "{ invalido");
        assertEquals(400, r.statusCode());
        Map<String, Object> corpo = parse(r.body());
        assertEquals("PARAM_INVALIDO", corpo.get("error"));
    }

    @Test
    void validateComAssinaturaCorretaRetornaSucesso() throws Exception {
        String b64 = Base64.getEncoder().encodeToString("dummy".getBytes(StandardCharsets.UTF_8));
        Map<String, Object> payload = Map.of(
                "jws", b64,
                "trustStore", List.of("a".repeat(64)),
                "minCertIssueDate", 0,
                "referenceTimestamp", 0,
                "signaturePolicyId", "urn:test:1.0");
        HttpResponse<String> r = post("/validate", JSON.writeValueAsString(payload));
        assertEquals(200, r.statusCode());
        Map<String, Object> corpo = parse(r.body());
        assertEquals("OperationOutcome", corpo.get("resourceType"));
    }

    @Test
    void shutdownRetornaConfirmacao() throws Exception {
        HttpResponse<String> r = post("/shutdown", "");
        assertEquals(200, r.statusCode());
        assertTrue(r.body().contains("SHUTTING_DOWN"));
    }

    private HttpResponse<String> get(String caminho) throws Exception {
        HttpRequest req = HttpRequest.newBuilder(URI.create("http://localhost:" + porta + caminho))
                .timeout(Duration.ofSeconds(2)).GET().build();
        return client.send(req, BodyHandlers.ofString());
    }

    private HttpResponse<String> post(String caminho, String corpo) throws Exception {
        HttpRequest req = HttpRequest.newBuilder(URI.create("http://localhost:" + porta + caminho))
                .timeout(Duration.ofSeconds(2))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(corpo)).build();
        return client.send(req, BodyHandlers.ofString());
    }

    private Map<String, Object> parse(String corpo) throws Exception {
        return JSON.readValue(corpo, new TypeReference<>() {});
    }

    private static Map<String, Object> payloadValidoCriacao() {
        String b64 = Base64.getEncoder().encodeToString("x".getBytes(StandardCharsets.UTF_8));
        return Map.of(
                "bundle", "{\"resourceType\":\"Bundle\"}",
                "provenance", "{\"resourceType\":\"Provenance\"}",
                "materialCriptografico", Map.of("tipo", "PEM", "chavePrivada", "----PEM----"),
                "certificados", List.of(b64),
                "timestampReferencia", 1700000000,
                "estrategiaTimestamp", "iat",
                "politicaAssinatura", "urn:test:1.0");
    }
}
