package br.gov.saude.assinador.erro;

import br.gov.saude.assinador.servico.AssinadorException.Codigo;

/**
 * Mapeia codigos de erro de dominio para HTTP status e codigo de saida do processo CLI.
 * Tabela de referencia em planejamento/entregavel-assinador.md.
 */
public final class MapeadorErro {

    private MapeadorErro() {}

    public static int httpStatus(Codigo codigo) {
        if (codigo == null) return 500;
        return switch (codigo) {
            case PARAM_AUSENTE, PARAM_INVALIDO, ALGORITMO_NAO_SUPORTADO -> 400;
            case ASSINATURA_INVALIDA -> 422;
            case PIN_INVALIDO -> 401;
            case DISPOSITIVO_INDISPONIVEL -> 503;
            case PAYLOAD_MUITO_GRANDE -> 413;
            case ERRO_INTERNO -> 500;
        };
    }

    /**
     * Exit code alinhado com {@code assinatura/cmd/exit.go}.
     */
    public static int exitCode(Codigo codigo) {
        if (codigo == null) return 1;
        return switch (codigo) {
            case PARAM_AUSENTE -> 2;
            case PARAM_INVALIDO -> 3;
            case ALGORITMO_NAO_SUPORTADO -> 4;
            case ASSINATURA_INVALIDA -> 5;
            case DISPOSITIVO_INDISPONIVEL, PIN_INVALIDO -> 6;
            case ERRO_INTERNO -> 7;
            case PAYLOAD_MUITO_GRANDE -> 1;
        };
    }
}
