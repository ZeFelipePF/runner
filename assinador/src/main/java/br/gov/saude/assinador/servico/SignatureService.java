package br.gov.saude.assinador.servico;

/**
 * Interface central para operacoes de assinatura digital.
 * Implementacoes: FakeSignatureService (simulacao), PKCS11SignatureService (dispositivo).
 *
 * Ambos os metodos lancam AssinadorException para parametros invalidos ou ausentes.
 * Os codigos de erro seguem o contrato definido em planejamento/contrato-fhir.md.
 */
public interface SignatureService {

    /**
     * Cria uma assinatura digital para a mensagem fornecida.
     *
     * @param message    conteudo a assinar, codificado em base64 — obrigatorio
     * @param privateKey identificador ou conteudo da chave privada — obrigatorio
     * @return assinatura gerada, codificada em base64
     * @throws AssinadorException PARAM_AUSENTE se message ou privateKey forem nulos/vazios;
     *                            PARAM_INVALIDO se message nao for base64 valido
     */
    String sign(String message, String privateKey) throws AssinadorException;

    /**
     * Valida uma assinatura digital.
     *
     * @param message   conteudo original, codificado em base64 — obrigatorio
     * @param signature assinatura a validar, codificada em base64 — obrigatorio
     * @param publicKey identificador ou conteudo da chave publica — obrigatorio
     * @return true se a assinatura for valida
     * @throws AssinadorException PARAM_AUSENTE se qualquer parametro for nulo/vazio;
     *                            PARAM_INVALIDO se message ou signature nao forem base64 validos
     */
    boolean validate(String message, String signature, String publicKey) throws AssinadorException;
}
