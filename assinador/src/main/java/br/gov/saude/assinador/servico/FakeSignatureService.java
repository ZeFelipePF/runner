package br.gov.saude.assinador.servico;

import java.nio.charset.StandardCharsets;
import java.util.Base64;

/**
 * Implementacao simulada de SignatureService.
 *
 * Valida parametros obrigatorios e formato base64 e retorna valores pre-construidos.
 * Nao executa criptografia real — destinada a testes e demonstracao.
 */
public class FakeSignatureService implements SignatureService {

    /** Assinatura fixa retornada por sign() — base64 de "ASSINATURA-SIMULADA-v1". */
    public static final String ASSINATURA_SIMULADA =
            Base64.getEncoder().encodeToString("ASSINATURA-SIMULADA-v1".getBytes(StandardCharsets.UTF_8));

    private static final Base64.Decoder BASE64_DECODER = Base64.getDecoder();

    @Override
    public String sign(String message, String privateKey) throws AssinadorException {
        exigirPreenchido("message", message);
        exigirPreenchido("privateKey", privateKey);
        exigirBase64("message", message);
        return ASSINATURA_SIMULADA;
    }

    @Override
    public boolean validate(String message, String signature, String publicKey) throws AssinadorException {
        exigirPreenchido("message", message);
        exigirPreenchido("signature", signature);
        exigirPreenchido("publicKey", publicKey);
        exigirBase64("message", message);
        exigirBase64("signature", signature);
        return ASSINATURA_SIMULADA.equals(signature);
    }

    private static void exigirPreenchido(String nome, String valor) {
        if (valor == null || valor.isBlank()) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_AUSENTE,
                    "campo '" + nome + "' obrigatorio");
        }
    }

    private static void exigirBase64(String nome, String valor) {
        try {
            BASE64_DECODER.decode(valor);
        } catch (IllegalArgumentException e) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_INVALIDO,
                    "campo '" + nome + "' nao e base64 valido",
                    e);
        }
    }
}
