package br.gov.saude.assinador.cli;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import br.gov.saude.assinador.servico.FakeSignatureService;
import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class AcaoValidarTest {

    private final AcaoValidar acao = new AcaoValidar(new FakeSignatureService());

    private Map<String, Object> payloadVerificacaoValido(String jws) {
        Map<String, Object> p = new HashMap<>();
        p.put("jws", jws);
        p.put("trustStore", List.of("a".repeat(64)));
        p.put("minCertIssueDate", 1_600_000_000L);
        p.put("referenceTimestamp", 1_700_000_000L);
        p.put("signaturePolicyId", "urn:policy:v1");
        return p;
    }

    @Test
    @SuppressWarnings("unchecked")
    void validateAceitaAssinaturaSimulada() {
        Map<String, Object> p = payloadVerificacaoValido(FakeSignatureService.ASSINATURA_SIMULADA);
        Map<String, Object> r = acao.executar(p);
        assertEquals("OperationOutcome", r.get("resourceType"));
        List<Map<String, Object>> issues = (List<Map<String, Object>>) r.get("issue");
        assertEquals("information", issues.get(0).get("severity"));
        assertEquals("informational", issues.get(0).get("code"));
    }

    @Test
    @SuppressWarnings("unchecked")
    void validateRejeitaAssinaturaDiferente() {
        Map<String, Object> r = acao.executar(payloadVerificacaoValido("anNvLW91dHJv"));
        List<Map<String, Object>> issues = (List<Map<String, Object>>) r.get("issue");
        assertEquals("error", issues.get(0).get("severity"));
        assertEquals("invalid", issues.get(0).get("code"));
    }

    @Test
    void validatePropagaErroDeValidacao() {
        Map<String, Object> p = payloadVerificacaoValido("anNv");
        p.remove("jws");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> acao.executar(p));
        assertEquals(Codigo.PARAM_AUSENTE, ex.getCodigo());
    }
}
