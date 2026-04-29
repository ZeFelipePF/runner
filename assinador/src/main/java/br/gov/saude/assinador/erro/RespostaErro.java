package br.gov.saude.assinador.erro;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;

/**
 * Resposta de erro estruturada — formato JSON: { "error": "<CODIGO>", "message": "<descricao>" }.
 * Compativel com modo CLI (stderr) e modo HTTP (corpo da resposta).
 */
public record RespostaErro(String error, String message) {

    public static RespostaErro de(AssinadorException ex) {
        Codigo codigo = ex.getCodigo() == null ? Codigo.ERRO_INTERNO : ex.getCodigo();
        String msg = ex.getMessage() == null ? "" : ex.getMessage();
        return new RespostaErro(codigo.name(), msg);
    }

    public static RespostaErro de(Codigo codigo, String mensagem) {
        return new RespostaErro(codigo.name(), mensagem == null ? "" : mensagem);
    }
}
