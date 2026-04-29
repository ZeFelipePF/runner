package br.gov.saude.assinador.cli;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import br.gov.saude.assinador.servico.FakeSignatureService;
import br.gov.saude.assinador.servico.SignatureService;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.*;

class AssinadorCliTest {

    private final ObjectMapper json = new ObjectMapper();
    private ByteArrayOutputStream outBuf;
    private ByteArrayOutputStream errBuf;
    private PrintStream out;
    private PrintStream err;

    @BeforeEach
    void setUp() {
        outBuf = new ByteArrayOutputStream();
        errBuf = new ByteArrayOutputStream();
        out = new PrintStream(outBuf, true, StandardCharsets.UTF_8);
        err = new PrintStream(errBuf, true, StandardCharsets.UTF_8);
    }

    private AssinadorCli cli(InputStream stdin, SignatureService svc) {
        return new AssinadorCli(svc, stdin, out, err);
    }

    private AssinadorCli cli(InputStream stdin) {
        return cli(stdin, new FakeSignatureService());
    }

    private static InputStream stdin(String s) {
        return new ByteArrayInputStream(s.getBytes(StandardCharsets.UTF_8));
    }

    private static InputStream stdinVazio() {
        return new ByteArrayInputStream(new byte[0]);
    }

    private static String payloadSignValido() {
        return """
                {
                  "bundle": "{}",
                  "provenance": "{}",
                  "materialCriptografico": {"tipo": "PEM", "chavePrivada": "k"},
                  "certificados": ["Y2VydA=="],
                  "timestampReferencia": 1700000000,
                  "estrategiaTimestamp": "iat",
                  "politicaAssinatura": "urn:policy:v1"
                }
                """;
    }

    private static String payloadValidateValido(String jws) {
        return """
                {
                  "jws": "%s",
                  "trustStore": ["%s"],
                  "minCertIssueDate": 1600000000,
                  "referenceTimestamp": 1700000000,
                  "signaturePolicyId": "urn:policy:v1"
                }
                """.formatted(jws, "a".repeat(64));
    }

    private String stdout() { return outBuf.toString(StandardCharsets.UTF_8); }
    private String stderr() { return errBuf.toString(StandardCharsets.UTF_8); }

    @Test
    void semArgsExibeAjuda() {
        int code = cli(stdinVazio()).executar(new String[]{});
        assertEquals(0, code);
        assertTrue(stdout().contains("Uso: assinador"));
        assertEquals("", stderr());
    }

    @Test
    void flagHelpExibeAjuda() {
        int code = cli(stdinVazio()).executar(new String[]{"-h"});
        assertEquals(0, code);
        assertTrue(stdout().contains("Comandos:"));
    }

    @Test
    void flagHelpLongoExibeAjuda() {
        int code = cli(stdinVazio()).executar(new String[]{"--help"});
        assertEquals(0, code);
        assertTrue(stdout().contains("Comandos:"));
    }

    @Test
    void comandoHelpExibeAjuda() {
        int code = cli(stdinVazio()).executar(new String[]{"help"});
        assertEquals(0, code);
        assertTrue(stdout().contains("Comandos:"));
    }

    @Test
    void comandoDesconhecidoRetornaParamInvalido() throws Exception {
        int code = cli(stdin(payloadSignValido())).executar(new String[]{"foo"});
        assertEquals(3, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("PARAM_INVALIDO", err.get("error").asText());
    }

    @Test
    void signComStdinValidoRetornaSignature() throws Exception {
        int code = cli(stdin(payloadSignValido())).executar(new String[]{"sign"});
        assertEquals(0, code);
        JsonNode r = json.readTree(stdout());
        assertEquals("Signature", r.get("resourceType").asText());
        assertEquals(FakeSignatureService.ASSINATURA_SIMULADA, r.get("data").asText());
    }

    @Test
    void signComInputArquivoFunciona(@TempDir Path tmp) throws Exception {
        Path arquivo = tmp.resolve("payload.json");
        Files.writeString(arquivo, payloadSignValido());
        int code = cli(stdinVazio()).executar(new String[]{"sign", "--input", arquivo.toString()});
        assertEquals(0, code);
        JsonNode r = json.readTree(stdout());
        assertEquals("Signature", r.get("resourceType").asText());
    }

    @Test
    void signComInputDashLeStdin() throws Exception {
        int code = cli(stdin(payloadSignValido())).executar(new String[]{"sign", "--input", "-"});
        assertEquals(0, code);
        JsonNode r = json.readTree(stdout());
        assertEquals("Signature", r.get("resourceType").asText());
    }

    @Test
    void signSemBundleFalhaComParamAusente() throws Exception {
        String payload = payloadSignValido().replace("\"bundle\": \"{}\",", "");
        int code = cli(stdin(payload)).executar(new String[]{"sign"});
        assertEquals(2, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("PARAM_AUSENTE", err.get("error").asText());
    }

    @Test
    void signComAlgoritmoNaoSuportado() throws Exception {
        String payload = payloadSignValido().replace(
                "\"politicaAssinatura\": \"urn:policy:v1\"",
                "\"politicaAssinatura\": \"urn:policy:v1\", \"algoritmo\": \"HS256\"");
        int code = cli(stdin(payload)).executar(new String[]{"sign"});
        assertEquals(4, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("ALGORITMO_NAO_SUPORTADO", err.get("error").asText());
    }

    @Test
    void validateComStdinAceitaAssinaturaSimulada() throws Exception {
        String p = payloadValidateValido(FakeSignatureService.ASSINATURA_SIMULADA);
        int code = cli(stdin(p)).executar(new String[]{"validate"});
        assertEquals(0, code);
        JsonNode r = json.readTree(stdout());
        assertEquals("OperationOutcome", r.get("resourceType").asText());
        assertEquals("information", r.get("issue").get(0).get("severity").asText());
    }

    @Test
    void validateRejeitaAssinaturaDiferente() throws Exception {
        int code = cli(stdin(payloadValidateValido("YW55"))).executar(new String[]{"validate"});
        assertEquals(0, code);
        JsonNode r = json.readTree(stdout());
        assertEquals("error", r.get("issue").get(0).get("severity").asText());
    }

    @Test
    void inputArquivoInexistenteFalha() {
        int code = cli(stdinVazio()).executar(new String[]{"sign", "--input", "nao-existe.json"});
        assertEquals(3, code);
        assertTrue(stderr().contains("PARAM_INVALIDO"));
    }

    @Test
    void inputSemValorFalha() throws Exception {
        int code = cli(stdinVazio()).executar(new String[]{"sign", "--input"});
        assertEquals(2, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("PARAM_AUSENTE", err.get("error").asText());
    }

    @Test
    void payloadVazioStdinFalha() throws Exception {
        int code = cli(stdinVazio()).executar(new String[]{"sign"});
        assertEquals(2, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("PARAM_AUSENTE", err.get("error").asText());
    }

    @Test
    void jsonMalformadoFalhaComParamInvalido() throws Exception {
        int code = cli(stdin("{ nao e json valido")).executar(new String[]{"sign"});
        assertEquals(3, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("PARAM_INVALIDO", err.get("error").asText());
    }

    @Test
    void erroInternoEhCapturado() throws Exception {
        SignatureService quebrado = new SignatureService() {
            @Override public String sign(String m, String k) { throw new IllegalStateException("boom"); }
            @Override public boolean validate(String m, String s, String k) { return false; }
        };
        int code = cli(stdin(payloadSignValido()), quebrado).executar(new String[]{"sign"});
        assertEquals(1, code);
        JsonNode err = json.readTree(stderr());
        assertEquals("ERRO_INTERNO", err.get("error").asText());
        assertTrue(err.get("message").asText().contains("boom"));
    }

    @Test
    void mainPropagaErroPorExit() {
        AssinadorException ex = assertThrows(AssinadorException.class, () -> {
            throw new AssinadorException(Codigo.PARAM_AUSENTE, "test");
        });
        assertEquals(Codigo.PARAM_AUSENTE, ex.getCodigo());
    }
}
