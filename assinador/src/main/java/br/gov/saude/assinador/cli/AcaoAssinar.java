package br.gov.saude.assinador.cli;

import br.gov.saude.assinador.servico.SignatureService;
import br.gov.saude.assinador.validacao.ValidadorFHIR;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Base64;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Orquestra a operacao de criacao de assinatura: valida payload FHIR, delega ao
 * SignatureService e monta a resposta no formato Signature FHIR R4.
 *
 * A representacao de saida segue o exemplo de planejamento/contrato-fhir.md
 * (resourceType=Signature, sigFormat=application/jose).
 */
public final class AcaoAssinar {

    private final SignatureService service;

    public AcaoAssinar(SignatureService service) {
        this.service = service;
    }

    public Map<String, Object> executar(Map<String, Object> payload) {
        ValidadorFHIR.validarCriacao(payload);

        String mensagemCanonica = canonicalizar(payload);
        String chavePrivada = extrairChavePrivada(payload);
        String assinatura = service.sign(mensagemCanonica, chavePrivada);

        return montarSignature(payload, assinatura);
    }

    private static String canonicalizar(Map<String, Object> payload) {
        String bundle = (String) payload.get("bundle");
        String provenance = (String) payload.get("provenance");
        String concat = bundle + "|" + provenance;
        return Base64.getEncoder().encodeToString(concat.getBytes(StandardCharsets.UTF_8));
    }

    @SuppressWarnings("unchecked")
    private static String extrairChavePrivada(Map<String, Object> payload) {
        Map<String, Object> material = (Map<String, Object>) payload.get("materialCriptografico");
        Object tipo = material.get("tipo");
        if ("PEM".equals(tipo) && material.get("chavePrivada") instanceof String s) return s;
        if ("PKCS12".equals(tipo) && material.get("alias") instanceof String s) return s;
        if (("SMARTCARD".equals(tipo) || "TOKEN".equals(tipo))
                && material.get("identificador") instanceof String s) return s;
        if ("REMOTE".equals(tipo) && material.get("enderecoServico") instanceof String s) return s;
        return tipo == null ? "desconhecido" : tipo.toString();
    }

    private static Map<String, Object> montarSignature(Map<String, Object> payload, String assinatura) {
        Map<String, Object> signature = new LinkedHashMap<>();
        signature.put("resourceType", "Signature");
        signature.put("type", List.of(Map.of(
                "system", "urn:iso-astm:E1762-95:2013",
                "code", "1.2.840.10065.1.12.1.1")));
        signature.put("when", Instant.ofEpochSecond(((Number) payload.get("timestampReferencia")).longValue()).toString());
        signature.put("sigFormat", "application/jose");
        signature.put("targetFormat", "application/octet-stream");
        signature.put("data", assinatura);
        return signature;
    }
}
