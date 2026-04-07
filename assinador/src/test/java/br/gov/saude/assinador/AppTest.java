package br.gov.saude.assinador;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class AppTest {
    @Test
    void appStartsWithoutException() {
        assertDoesNotThrow(() -> App.main(new String[]{}));
    }
}
