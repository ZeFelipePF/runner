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
            case PAYLOAD_MUITO_GRANDE -> 413;
            case ERRO_INTERNO -> 500;
        };
    }

    public static int exitCode(Codigo codigo) {
        if (codigo == null) return 1;
        return switch (codigo) {
            case PARAM_AUSENTE -> 2;
            case PARAM_INVALIDO -> 3;
            case ALGORITMO_NAO_SUPORTADO -> 4;
            case PAYLOAD_MUITO_GRANDE -> 5;
            case ERRO_INTERNO -> 1;
        };
    }
}
