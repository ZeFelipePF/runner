package br.gov.saude.assinador.servico;

public class AssinadorException extends RuntimeException {

    public enum Codigo {
        PARAM_AUSENTE,
        PARAM_INVALIDO,
        ALGORITMO_NAO_SUPORTADO,
        PAYLOAD_MUITO_GRANDE,
        ERRO_INTERNO
    }

    private final Codigo codigo;

    public AssinadorException(Codigo codigo, String message) {
        super(message);
        this.codigo = codigo;
    }

    public AssinadorException(Codigo codigo, String message, Throwable cause) {
        super(message, cause);
        this.codigo = codigo;
    }

    public Codigo getCodigo() {
        return codigo;
    }
}
