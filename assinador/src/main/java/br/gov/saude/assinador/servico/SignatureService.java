package br.gov.saude.assinador.servico;

/**
 * Interface central para operacoes de assinatura digital.
 * Implementacoes: FakeSignatureService (simulacao), PKCS11SignatureService (dispositivo).
 */
public interface SignatureService {

    /**
     * Cria uma assinatura digital para a mensagem fornecida.
     *
     * @param message    conteudo a assinar, codificado em base64
     * @param privateKey identificador ou conteudo da chave privada
     * @return assinatura gerada, codificada em base64
     */
    String sign(String message, String privateKey);

    /**
     * Valida uma assinatura digital.
     *
     * @param message   conteudo original, codificado em base64
     * @param signature assinatura a validar, codificada em base64
     * @param publicKey identificador ou conteudo da chave publica
     * @return true se a assinatura for valida
     */
    boolean validate(String message, String signature, String publicKey);
}
