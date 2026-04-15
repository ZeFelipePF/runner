package br.gov.saude.assinador.validacao;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class ValidadorFHIRTest {

    private Map<String, Object> payloadCriacaoValido() {
        Map<String, Object> p = new HashMap<>();
        p.put("bundle", "{}");
        p.put("provenance", "{}");
        p.put("materialCriptografico", Map.of("tipo", "PEM", "chavePrivada", "---"));
        p.put("certificados", List.of("Y2VydA=="));
        p.put("timestampReferencia", 1_700_000_000L);
        p.put("estrategiaTimestamp", "iat");
        p.put("politicaAssinatura", "urn:policy:v1");
        return p;
    }

    private Map<String, Object> payloadVerificacaoValido() {
        Map<String, Object> p = new HashMap<>();
        p.put("jws", "anNv");
        p.put("trustStore", List.of("a".repeat(64)));
        p.put("minCertIssueDate", 1_600_000_000L);
        p.put("referenceTimestamp", 1_700_000_000L);
        p.put("signaturePolicyId", "urn:policy:v1");
        return p;
    }

    @Test
    void criacaoAceitaPayloadValido() {
        assertDoesNotThrow(() -> ValidadorFHIR.validarCriacao(payloadCriacaoValido()));
    }

    @Test
    void criacaoRejeitaBundleAusente() {
        Map<String, Object> p = payloadCriacaoValido();
        p.remove("bundle");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarCriacao(p));
        assertEquals(Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void criacaoRejeitaCertificadoNaoBase64() {
        Map<String, Object> p = payloadCriacaoValido();
        p.put("certificados", List.of("@@@"));
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarCriacao(p));
        assertEquals(Codigo.PARAM_INVALIDO, ex.getCodigo());
    }

    @Test
    void criacaoRejeitaEstrategiaTimestampInvalida() {
        Map<String, Object> p = payloadCriacaoValido();
        p.put("estrategiaTimestamp", "zzz");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarCriacao(p));
        assertEquals(Codigo.PARAM_INVALIDO, ex.getCodigo());
    }

    @Test
    void criacaoRejeitaAlgoritmoNaoSuportado() {
        Map<String, Object> p = payloadCriacaoValido();
        p.put("algoritmo", "HS256");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarCriacao(p));
        assertEquals(Codigo.ALGORITMO_NAO_SUPORTADO, ex.getCodigo());
    }

    @Test
    void criacaoAceitaAlgoritmoSuportado() {
        Map<String, Object> p = payloadCriacaoValido();
        p.put("algoritmo", "RS256");
        assertDoesNotThrow(() -> ValidadorFHIR.validarCriacao(p));
    }

    @Test
    void verificacaoAceitaPayloadValido() {
        assertDoesNotThrow(() -> ValidadorFHIR.validarVerificacao(payloadVerificacaoValido()));
    }

    @Test
    void verificacaoRejeitaJwsAusente() {
        Map<String, Object> p = payloadVerificacaoValido();
        p.remove("jws");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarVerificacao(p));
        assertEquals(Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void verificacaoRejeitaHashTrustStoreInvalido() {
        Map<String, Object> p = payloadVerificacaoValido();
        p.put("trustStore", List.of("xyz"));
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarVerificacao(p));
        assertEquals(Codigo.PARAM_INVALIDO, ex.getCodigo());
    }

    @Test
    void verificacaoRejeitaJwsNaoBase64() {
        Map<String, Object> p = payloadVerificacaoValido();
        p.put("jws", "@@@");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> ValidadorFHIR.validarVerificacao(p));
        assertEquals(Codigo.PARAM_INVALIDO, ex.getCodigo());
    }
}
