package br.gov.saude.assinador.servico;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;
import org.junit.jupiter.api.io.TempDir;

import java.nio.file.Files;
import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Testes de integracao PKCS#11 com SoftHSM2 (US-02.5 da spec).
 *
 * Opt-in: rodam apenas quando o usuario define:
 * <ul>
 *   <li>{@code SOFTHSM2_LIB}: caminho da biblioteca nativa (.so / .dylib / .dll)</li>
 *   <li>{@code SOFTHSM2_PIN}: PIN do token (default 1234 se ausente — ver setup)</li>
 *   <li>{@code SOFTHSM2_ALIAS}: alias da chave previamente importada (default testkey)</li>
 * </ul>
 *
 * Setup completo (Linux/macOS/Windows) em {@code docs/tecnico.md} §5.2.
 *
 * Executar com: {@code ./mvnw test -Dgroups=pkcs11}
 */
@Tag("pkcs11")
@EnabledIfEnvironmentVariable(named = "SOFTHSM2_LIB", matches = ".+")
class PKCS11IntegrationTest {

    private static Path pkcs11Cfg;
    private static char[] pin;
    private static String alias;

    @BeforeAll
    static void setup(@TempDir Path tmp) throws Exception {
        String lib = System.getenv("SOFTHSM2_LIB");
        pin = System.getenv().getOrDefault("SOFTHSM2_PIN", "1234").toCharArray();
        alias = System.getenv().getOrDefault("SOFTHSM2_ALIAS", "testkey");

        pkcs11Cfg = tmp.resolve("pkcs11.cfg");
        Files.writeString(pkcs11Cfg, String.join("\n",
                "name = SoftHSM2",
                "library = " + lib,
                "slot = 0"
        ));
    }

    @Test
    void assinaEValidaComToken() {
        PKCS11SignatureService svc = new PKCS11SignatureService(pkcs11Cfg, pin);
        String mensagem = "documento-teste";
        String assinatura = svc.sign(mensagem, alias);
        assertNotNull(assinatura);
        assertFalse(assinatura.isBlank());
        assertTrue(svc.validate(mensagem, assinatura, alias));
    }

    @Test
    void pinErradoSinalizaPinInvalido() {
        PKCS11SignatureService svc = new PKCS11SignatureService(
                pkcs11Cfg, "9999".toCharArray());
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> svc.sign("msg", alias));
        assertEquals(AssinadorException.Codigo.PIN_INVALIDO, ex.getCodigo());
    }

    @Test
    void aliasInexistenteSinalizaParamInvalido() {
        PKCS11SignatureService svc = new PKCS11SignatureService(pkcs11Cfg, pin);
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> svc.sign("msg", "alias-inexistente"));
        assertEquals(AssinadorException.Codigo.PARAM_INVALIDO, ex.getCodigo());
    }
}
