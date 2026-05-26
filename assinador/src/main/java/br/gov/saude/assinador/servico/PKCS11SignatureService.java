package br.gov.saude.assinador.servico;

import java.nio.file.Path;
import java.security.GeneralSecurityException;
import java.security.KeyStore;
import java.security.PrivateKey;
import java.security.Provider;
import java.security.ProviderException;
import java.security.PublicKey;
import java.security.Security;
import java.security.Signature;
import java.security.cert.Certificate;
import java.util.Base64;

/**
 * SignatureService apoiado em dispositivo criptografico via SunPKCS11.
 *
 * A chave privada nunca sai do dispositivo: os parametros {@code privateKey} /
 * {@code publicKey} da interface {@link SignatureService} sao usados aqui como
 * <strong>alias</strong> dentro do KeyStore PKCS#11. A autenticacao usa o PIN
 * fornecido na construcao.
 *
 * <h3>Setup para desenvolvimento (SoftHSM2)</h3>
 * Veja {@code docs/tecnico.md} §5.2 — "Setup do SoftHSM2".
 *
 * <h3>Tratamento de erros (alinhado com {@code cmd/exit.go} no CLI Go)</h3>
 * <ul>
 *   <li>{@code DISPOSITIVO_INDISPONIVEL} — driver PKCS#11 nao carrega (HTTP 503, exit 6)</li>
 *   <li>{@code PIN_INVALIDO} — PIN incorreto (HTTP 401, exit 6)</li>
 *   <li>{@code PARAM_INVALIDO} — alias inexistente (HTTP 400, exit 3)</li>
 * </ul>
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
        if (message == null) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_AUSENTE, "message obrigatoria");
        }
        if (privateKeyAlias == null || privateKeyAlias.isBlank()) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_AUSENTE, "alias da chave privada obrigatorio");
        }
        try {
            KeyStore ks = abrirKeyStore();
            PrivateKey chave = (PrivateKey) ks.getKey(privateKeyAlias, null);
            if (chave == null) {
                throw new AssinadorException(
                        AssinadorException.Codigo.PARAM_INVALIDO,
                        "alias '" + privateKeyAlias + "' nao encontrado no dispositivo");
            }
            Signature sig = Signature.getInstance("SHA256withRSA");
            sig.initSign(chave);
            sig.update(message.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            return Base64.getEncoder().encodeToString(sig.sign());
        } catch (AssinadorException ex) {
            throw ex;
        } catch (GeneralSecurityException ex) {
            throw mapearExcecaoCriptografica(ex);
        }
    }

    @Override
    public boolean validate(String message, String signature, String publicKeyAlias) throws AssinadorException {
        if (message == null || signature == null || publicKeyAlias == null || publicKeyAlias.isBlank()) {
            throw new AssinadorException(
                    AssinadorException.Codigo.PARAM_AUSENTE,
                    "message, signature e alias da chave publica sao obrigatorios");
        }
        KeyStore ks = abrirKeyStore();
        try {
            Certificate cert = ks.getCertificate(publicKeyAlias);
            if (cert == null) {
                throw new AssinadorException(
                        AssinadorException.Codigo.PARAM_INVALIDO,
                        "alias '" + publicKeyAlias + "' nao encontrado no dispositivo");
            }
            PublicKey pub = cert.getPublicKey();
            Signature sig = Signature.getInstance("SHA256withRSA");
            sig.initVerify(pub);
            sig.update(message.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            byte[] assinatura;
            try {
                assinatura = Base64.getDecoder().decode(signature);
            } catch (IllegalArgumentException ex) {
                throw new AssinadorException(
                        AssinadorException.Codigo.PARAM_INVALIDO,
                        "assinatura nao e base64 valido", ex);
            }
            return sig.verify(assinatura);
        } catch (AssinadorException ex) {
            throw ex;
        } catch (GeneralSecurityException ex) {
            throw mapearExcecaoCriptografica(ex);
        }
    }

    /**
     * Constroi um KeyStore PKCS#11 ja autenticado.
     *
     * Falhas:
     * <ul>
     *   <li>driver nativo nao encontrado/carregavel → {@code DISPOSITIVO_INDISPONIVEL}</li>
     *   <li>PIN incorreto → {@code PIN_INVALIDO}</li>
     * </ul>
     */
    private KeyStore abrirKeyStore() {
        Provider provider;
        try {
            Provider base = Security.getProvider("SunPKCS11");
            if (base == null) {
                throw new AssinadorException(
                        AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL,
                        "Provider SunPKCS11 nao disponivel nesta JVM");
            }
            provider = base.configure(configPath.toString());
            if (Security.getProvider(provider.getName()) == null) {
                Security.addProvider(provider);
            }
        } catch (RuntimeException ex) {
            // SunPKCS11 lanca ProviderException (driver nao carrega) ou
            // InvalidParameterException (config invalido/inexistente). Em ambos os
            // casos, do ponto de vista do usuario, o dispositivo esta indisponivel.
            if (ex instanceof AssinadorException) throw ex;
            throw new AssinadorException(
                    AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL,
                    "Driver PKCS#11 nao pode ser carregado: " + configPath,
                    ex);
        }
        try {
            KeyStore ks = KeyStore.getInstance("PKCS11", provider);
            ks.load(null, pin);
            return ks;
        } catch (java.io.IOException ex) {
            // SunPKCS11 lanca IOException com causa FailedLoginException quando PIN errado.
            if (ex.getCause() != null
                    && ex.getCause().getClass().getName().contains("FailedLogin")) {
                throw new AssinadorException(
                        AssinadorException.Codigo.PIN_INVALIDO,
                        "PIN incorreto para o token PKCS#11", ex);
            }
            throw new AssinadorException(
                    AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL,
                    "Falha ao carregar KeyStore PKCS#11: " + ex.getMessage(), ex);
        } catch (GeneralSecurityException ex) {
            throw new AssinadorException(
                    AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL,
                    "Falha ao inicializar KeyStore PKCS#11: " + ex.getMessage(), ex);
        }
    }

    private AssinadorException mapearExcecaoCriptografica(GeneralSecurityException ex) {
        return new AssinadorException(
                AssinadorException.Codigo.ERRO_INTERNO,
                "erro criptografico: " + ex.getMessage(), ex);
    }

    Path getConfigPath() {
        return configPath;
    }

    char[] getPin() {
        return pin.clone();
    }
}
