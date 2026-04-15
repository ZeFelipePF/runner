package br.gov.saude.assinador.servico;

import org.junit.jupiter.api.Test;

import java.nio.charset.StandardCharsets;
import java.util.Base64;

import static org.junit.jupiter.api.Assertions.*;

class FakeSignatureServiceTest {

    private final FakeSignatureService service = new FakeSignatureService();
    private final String msg = Base64.getEncoder().encodeToString("ola".getBytes(StandardCharsets.UTF_8));

    @Test
    void signRetornaAssinaturaFixa() {
        String s = service.sign(msg, "chave");
        assertEquals(FakeSignatureService.ASSINATURA_SIMULADA, s);
    }

    @Test
    void signLancaQuandoMessageAusente() {
        AssinadorException ex = assertThrows(AssinadorException.class, () -> service.sign(null, "chave"));
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void signLancaQuandoPrivateKeyAusente() {
        AssinadorException ex = assertThrows(AssinadorException.class, () -> service.sign(msg, " "));
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void signLancaQuandoMessageNaoBase64() {
        AssinadorException ex = assertThrows(AssinadorException.class, () -> service.sign("!!!nao-base64!!!", "chave"));
        assertEquals(AssinadorException.Codigo.PARAM_INVALIDO, ex.getCodigo());
    }

    @Test
    void validateAceitaAssinaturaCorreta() {
        assertTrue(service.validate(msg, FakeSignatureService.ASSINATURA_SIMULADA, "pub"));
    }

    @Test
    void validateRejeitaAssinaturaDiferente() {
        String outra = Base64.getEncoder().encodeToString("outra".getBytes(StandardCharsets.UTF_8));
        assertFalse(service.validate(msg, outra, "pub"));
    }

    @Test
    void validateLancaQuandoParametroAusente() {
        AssinadorException ex = assertThrows(AssinadorException.class,
                () -> service.validate(msg, FakeSignatureService.ASSINATURA_SIMULADA, null));
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
    }
}
