package br.gov.saude.assinador.servico;

import org.junit.jupiter.api.Test;

import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Testes unitarios do PKCS11SignatureService que rodam sempre (sem SoftHSM2).
 * Cobrem validacao de construtor e mapeamento de erro quando o driver nao carrega.
 *
 * Testes de integracao funcionais com SoftHSM2 estao em
 * {@code PKCS11IntegrationTest} (tag {@code pkcs11}, opt-in).
 */
class PKCS11SignatureServiceTest {

    @Test
    void construtorExigeConfigPath() {
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> new PKCS11SignatureService(null, "1234".toCharArray()));
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void construtorExigePin() {
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> new PKCS11SignatureService(Path.of("pkcs11.cfg"), new char[0]));
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void signComConfigInexistenteSinalizaDispositivoIndisponivel() {
        PKCS11SignatureService svc = new PKCS11SignatureService(
                Path.of("/tmp/pkcs11-inexistente.cfg"), "1234".toCharArray());
        AssinadorException ex = assertThrows(AssinadorException.class, () -> svc.sign("m", "alias"));
        // Provider sem driver vira DISPOSITIVO_INDISPONIVEL (HTTP 503, exit 6).
        assertEquals(AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL, ex.getCodigo());
    }

    @Test
    void validateComConfigInexistenteSinalizaDispositivoIndisponivel() {
        PKCS11SignatureService svc = new PKCS11SignatureService(
                Path.of("/tmp/pkcs11-inexistente.cfg"), "1234".toCharArray());
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> svc.validate("m", "AAAA", "alias"));
        assertEquals(AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL, ex.getCodigo());
    }

    @Test
    void signExigeAlias() {
        PKCS11SignatureService svc = new PKCS11SignatureService(
                Path.of("/tmp/pkcs11.cfg"), "1234".toCharArray());
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> svc.sign("msg", "  "));
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
    }
}
