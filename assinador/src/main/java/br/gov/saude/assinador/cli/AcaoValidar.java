package br.gov.saude.assinador.cli;

import br.gov.saude.assinador.servico.SignatureService;
import br.gov.saude.assinador.validacao.ValidadorFHIR;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Orquestra a operacao de validacao: valida payload, delega ao SignatureService
 * e monta a resposta no formato OperationOutcome FHIR R4.
 */
public final class AcaoValidar {

    private final SignatureService service;

    public AcaoValidar(SignatureService service) {
        this.service = service;
    }

    public Map<String, Object> executar(Map<String, Object> payload) {
        ValidadorFHIR.validarVerificacao(payload);

        String jws = (String) payload.get("jws");
        String politica = (String) payload.get("signaturePolicyId");
        boolean valido = service.validate(jws, jws, politica);

        return montarOperationOutcome(valido, politica);
    }

    private static Map<String, Object> montarOperationOutcome(boolean valido, String politica) {
        Map<String, Object> issue = new LinkedHashMap<>();
        if (valido) {
            issue.put("severity", "information");
            issue.put("code", "informational");
            issue.put("details", Map.of(
                    "coding", List.of(Map.of("code", "VALIDATION.SUCCESS")),
                    "text", "Assinatura digital validada com sucesso"));
            issue.put("diagnostics", "Politica: " + politica);
        } else {
            issue.put("severity", "error");
            issue.put("code", "invalid");
            issue.put("details", Map.of(
                    "coding", List.of(Map.of("code", "VALIDATION.SIGNATURE-VERIFICATION-FAILED")),
                    "text", "Assinatura nao corresponde ao valor esperado"));
            issue.put("diagnostics", "Politica: " + politica);
        }

        Map<String, Object> outcome = new LinkedHashMap<>();
        outcome.put("resourceType", "OperationOutcome");
        outcome.put("issue", List.of(issue));
        return outcome;
    }
}
