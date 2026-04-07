package br.gov.saude.assinador.servico;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class SignatureServiceTest {

    @Test
    void interfaceHasSignMethod() throws NoSuchMethodException {
        assertNotNull(SignatureService.class.getMethod("sign", String.class, String.class));
    }

    @Test
    void interfaceHasValidateMethod() throws NoSuchMethodException {
        assertNotNull(SignatureService.class.getMethod("validate", String.class, String.class, String.class));
    }
}
