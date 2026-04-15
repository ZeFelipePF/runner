package br.gov.saude.assinador.servico;

import org.junit.jupiter.api.Test;

import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.*;

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
    void signIndicaIndisponibilidade() {
        PKCS11SignatureService svc = new PKCS11SignatureService(Path.of("pkcs11.cfg"), "1234".toCharArray());
        AssinadorException ex = assertThrows(AssinadorException.class, () -> svc.sign("m", "alias"));
        assertEquals(AssinadorException.Codigo.ERRO_INTERNO, ex.getCodigo());
    }

    @Test
    void validateIndicaIndisponibilidade() {
        PKCS11SignatureService svc = new PKCS11SignatureService(Path.of("pkcs11.cfg"), "1234".toCharArray());
        assertThrows(AssinadorException.class, () -> svc.validate("m", "s", "alias"));
    }
}
