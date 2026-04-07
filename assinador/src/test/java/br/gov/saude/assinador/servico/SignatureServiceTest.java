package br.gov.saude.assinador.servico;

import org.junit.jupiter.api.Test;

import java.lang.reflect.Method;

import static org.junit.jupiter.api.Assertions.*;

class SignatureServiceTest {

    @Test
    void interfaceHasSignMethod() throws NoSuchMethodException {
        Method m = SignatureService.class.getMethod("sign", String.class, String.class);
        assertEquals(String.class, m.getReturnType());
    }

    @Test
    void interfaceHasValidateMethod() throws NoSuchMethodException {
        Method m = SignatureService.class.getMethod("validate", String.class, String.class, String.class);
        assertEquals(boolean.class, m.getReturnType());
    }

    @Test
    void signDeclaresAssinadorException() throws NoSuchMethodException {
        Method m = SignatureService.class.getMethod("sign", String.class, String.class);
        assertTrue(containsException(m, AssinadorException.class));
    }

    @Test
    void validateDeclaresAssinadorException() throws NoSuchMethodException {
        Method m = SignatureService.class.getMethod("validate", String.class, String.class, String.class);
        assertTrue(containsException(m, AssinadorException.class));
    }

    @Test
    void assinadorExceptionCodigosEstaoDefinidos() {
        assertNotNull(AssinadorException.Codigo.PARAM_AUSENTE);
        assertNotNull(AssinadorException.Codigo.PARAM_INVALIDO);
        assertNotNull(AssinadorException.Codigo.ALGORITMO_NAO_SUPORTADO);
        assertNotNull(AssinadorException.Codigo.PAYLOAD_MUITO_GRANDE);
        assertNotNull(AssinadorException.Codigo.ERRO_INTERNO);
    }

    @Test
    void assinadorExceptionPreservaCodigoEMensagem() {
        AssinadorException ex = new AssinadorException(
                AssinadorException.Codigo.PARAM_AUSENTE, "campo 'message' obrigatorio");
        assertEquals(AssinadorException.Codigo.PARAM_AUSENTE, ex.getCodigo());
        assertEquals("campo 'message' obrigatorio", ex.getMessage());
    }

    private boolean containsException(Method m, Class<?> exType) {
        for (Class<?> t : m.getExceptionTypes()) {
            if (t.isAssignableFrom(exType)) return true;
        }
        return false;
    }
}
