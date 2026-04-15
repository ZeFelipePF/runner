package br.gov.saude.assinador.validacao;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;

import java.util.Base64;
import java.util.List;
import java.util.Map;

/**
 * Validacao dos parametros FHIR de /sign e /validate.
 * Verifica presenca de campos obrigatorios, formato base64 e algoritmo suportado.
 * Nao valida semantica criptografica (ICP-Brasil, OCSP, CRL).
 */
public final class ValidadorFHIR {

    private static final Base64.Decoder BASE64 = Base64.getDecoder();

    private ValidadorFHIR() {}

    /** Valida payload de POST /sign. Lanca AssinadorException na primeira falha. */
    public static void validarCriacao(Map<String, Object> payload) {
        exigirPayload(payload);
        exigirTextoNaoVazio(payload, "bundle");
        exigirTextoNaoVazio(payload, "provenance");
        exigirObjeto(payload, "materialCriptografico");
        exigirListaNaoVazia(payload, "certificados");
        validarCertificados((List<?>) payload.get("certificados"));
        exigirInteiro(payload, "timestampReferencia");
        exigirEstrategiaTimestamp(payload);
        exigirTextoNaoVazio(payload, "politicaAssinatura");

        Object algoritmo = payload.get("algoritmo");
        if (algoritmo != null && !AlgoritmoSuportado.aceita(algoritmo.toString())) {
            throw new AssinadorException(Codigo.ALGORITMO_NAO_SUPORTADO,
                    "algoritmo '" + algoritmo + "' nao suportado; use " + AlgoritmoSuportado.SUPORTADOS);
        }
    }

    /** Valida payload de POST /validate. Lanca AssinadorException na primeira falha. */
    public static void validarVerificacao(Map<String, Object> payload) {
        exigirPayload(payload);
        exigirTextoNaoVazio(payload, "jws");
        exigirBase64(payload, "jws");
        exigirListaNaoVazia(payload, "trustStore");
        validarHashesSha256((List<?>) payload.get("trustStore"));
        exigirInteiro(payload, "minCertIssueDate");
        exigirInteiro(payload, "referenceTimestamp");
        exigirTextoNaoVazio(payload, "signaturePolicyId");
    }

    private static void exigirPayload(Map<String, Object> payload) {
        if (payload == null || payload.isEmpty()) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE, "payload vazio");
        }
    }

    private static void exigirTextoNaoVazio(Map<String, Object> payload, String campo) {
        Object v = payload.get(campo);
        if (!(v instanceof String s) || s.isBlank()) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE,
                    "campo '" + campo + "' obrigatorio");
        }
    }

    private static void exigirObjeto(Map<String, Object> payload, String campo) {
        Object v = payload.get(campo);
        if (!(v instanceof Map<?, ?> m) || m.isEmpty()) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE,
                    "campo '" + campo + "' obrigatorio");
        }
    }

    private static void exigirListaNaoVazia(Map<String, Object> payload, String campo) {
        Object v = payload.get(campo);
        if (!(v instanceof List<?> l) || l.isEmpty()) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE,
                    "campo '" + campo + "' obrigatorio");
        }
    }

    private static void exigirInteiro(Map<String, Object> payload, String campo) {
        Object v = payload.get(campo);
        if (!(v instanceof Number)) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE,
                    "campo '" + campo + "' obrigatorio (inteiro)");
        }
    }

    private static void exigirEstrategiaTimestamp(Map<String, Object> payload) {
        Object v = payload.get("estrategiaTimestamp");
        if (!(v instanceof String s) || s.isBlank()) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE,
                    "campo 'estrategiaTimestamp' obrigatorio");
        }
        if (!s.equals("iat") && !s.equals("tsa")) {
            throw new AssinadorException(Codigo.PARAM_INVALIDO,
                    "estrategiaTimestamp deve ser 'iat' ou 'tsa'");
        }
    }

    private static void exigirBase64(Map<String, Object> payload, String campo) {
        Object v = payload.get(campo);
        try {
            BASE64.decode(v.toString());
        } catch (IllegalArgumentException e) {
            throw new AssinadorException(Codigo.PARAM_INVALIDO,
                    "campo '" + campo + "' nao e base64 valido", e);
        }
    }

    private static void validarCertificados(List<?> certificados) {
        for (int i = 0; i < certificados.size(); i++) {
            Object c = certificados.get(i);
            if (!(c instanceof String s) || s.isBlank()) {
                throw new AssinadorException(Codigo.PARAM_INVALIDO,
                        "certificados[" + i + "] deve ser base64 nao vazio");
            }
            try {
                BASE64.decode(s);
            } catch (IllegalArgumentException e) {
                throw new AssinadorException(Codigo.PARAM_INVALIDO,
                        "certificados[" + i + "] nao e base64 valido", e);
            }
        }
    }

    private static void validarHashesSha256(List<?> hashes) {
        for (int i = 0; i < hashes.size(); i++) {
            Object h = hashes.get(i);
            if (!(h instanceof String s) || s.length() != 64 || !s.matches("[0-9a-fA-F]+")) {
                throw new AssinadorException(Codigo.PARAM_INVALIDO,
                        "trustStore[" + i + "] deve ser hash SHA-256 hex (64 chars)");
            }
        }
    }
}
