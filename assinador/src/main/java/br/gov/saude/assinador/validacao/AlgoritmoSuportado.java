package br.gov.saude.assinador.validacao;

import java.util.Set;

/**
 * Algoritmos de assinatura aceitos pelo assinador, conforme contrato FHIR.
 * Ver planejamento/contrato-fhir.md.
 */
public final class AlgoritmoSuportado {

    public static final Set<String> SUPORTADOS = Set.of("RS256", "ES256");

    private AlgoritmoSuportado() {}

    public static boolean aceita(String algoritmo) {
        return algoritmo != null && SUPORTADOS.contains(algoritmo);
    }
}
