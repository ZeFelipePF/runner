package br.gov.saude.assinador.erro;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class RespostaErroTest {

    @Test
    void deExceptionPreservaCodigoEMensagem() {
        AssinadorException ex = new AssinadorException(Codigo.PARAM_INVALIDO, "campo X invalido");
        RespostaErro r = RespostaErro.de(ex);
        assertEquals("PARAM_INVALIDO", r.error());
        assertEquals("campo X invalido", r.message());
    }

    @Test
    void deExceptionSemMensagemUsaStringVazia() {
        AssinadorException ex = new AssinadorException(Codigo.ERRO_INTERNO, null);
        RespostaErro r = RespostaErro.de(ex);
        assertEquals("ERRO_INTERNO", r.error());
        assertEquals("", r.message());
    }

    @Test
    void deCodigoEMensagem() {
        RespostaErro r = RespostaErro.de(Codigo.ALGORITMO_NAO_SUPORTADO, "alg X");
        assertEquals("ALGORITMO_NAO_SUPORTADO", r.error());
        assertEquals("alg X", r.message());
    }

    @Test
    void deCodigoComMensagemNullUsaStringVazia() {
        RespostaErro r = RespostaErro.de(Codigo.PAYLOAD_MUITO_GRANDE, null);
        assertEquals("PAYLOAD_MUITO_GRANDE", r.error());
        assertEquals("", r.message());
    }
}
