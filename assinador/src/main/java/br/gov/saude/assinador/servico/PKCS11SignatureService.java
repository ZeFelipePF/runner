package br.gov.saude.assinador.servico;

import java.nio.file.Path;

/**
 * Esqueleto de SignatureService apoiado em dispositivo criptografico via SunPKCS11.
 *
 * A chave privada permanece no dispositivo — os parametros privateKey/publicKey da
 * interface sao usados apenas como alias/identificador dentro do KeyStore PKCS#11.
 * A autenticacao e feita com o PIN fornecido na construcao.
 *
 * Este servico nao e funcional sem um token/smartcard fisico e o driver nativo do
 * fabricante. Os metodos lancam AssinadorException enquanto a integracao nao for
 * concluida.
 *
 * Fluxo previsto de inicializacao:
 *   Provider p = Security.getProvider("SunPKCS11").configure(configPath);
 *   Security.addProvider(p);
 *   KeyStore ks = KeyStore.getInstance("PKCS11", p);
 *   ks.load(null, pin);
 */
public class PKCS11SignatureService implements SignatureService {

    private final Path configPath;
    private final char[] pin;

    public PKCS11SignatureService(Path configPath, char[] pin) {
        if (configPath == null) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_AUSENTE,
                    "configPath obrigatorio");
        }
        if (pin == null || pin.length == 0) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_AUSENTE,
                    "pin obrigatorio");
        }
        this.configPath = configPath;
        this.pin = pin.clone();
    }

    @Override
    public String sign(String message, String privateKeyAlias) throws AssinadorException {
        throw new AssinadorException(
                AssinadorException.Codigo.ERRO_INTERNO,
                "PKCS11SignatureService nao implementado: dispositivo criptografico indisponivel");
    }

    @Override
    public boolean validate(String message, String signature, String publicKeyAlias) throws AssinadorException {
        throw new AssinadorException(
                AssinadorException.Codigo.ERRO_INTERNO,
                "PKCS11SignatureService nao implementado: dispositivo criptografico indisponivel");
    }

    Path getConfigPath() {
        return configPath;
    }

    char[] getPin() {
        return pin.clone();
    }
}
