package br.gov.saude.assinador.erro;

import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class MapeadorErroTest {

    @Test
    void httpStatusParaCadaCodigo() {
        assertEquals(400, MapeadorErro.httpStatus(Codigo.PARAM_AUSENTE));
        assertEquals(400, MapeadorErro.httpStatus(Codigo.PARAM_INVALIDO));
        assertEquals(400, MapeadorErro.httpStatus(Codigo.ALGORITMO_NAO_SUPORTADO));
        assertEquals(413, MapeadorErro.httpStatus(Codigo.PAYLOAD_MUITO_GRANDE));
        assertEquals(500, MapeadorErro.httpStatus(Codigo.ERRO_INTERNO));
    }

    @Test
    void httpStatusNullCai500() {
        assertEquals(500, MapeadorErro.httpStatus(null));
    }

    @Test
    void exitCodeDistintoPorCodigo() {
        assertEquals(2, MapeadorErro.exitCode(Codigo.PARAM_AUSENTE));
        assertEquals(3, MapeadorErro.exitCode(Codigo.PARAM_INVALIDO));
        assertEquals(4, MapeadorErro.exitCode(Codigo.ALGORITMO_NAO_SUPORTADO));
        assertEquals(5, MapeadorErro.exitCode(Codigo.PAYLOAD_MUITO_GRANDE));
        assertEquals(1, MapeadorErro.exitCode(Codigo.ERRO_INTERNO));
    }

    @Test
    void exitCodeNullCaiUm() {
        assertEquals(1, MapeadorErro.exitCode(null));
    }
}
