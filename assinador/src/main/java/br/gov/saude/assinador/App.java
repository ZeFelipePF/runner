package br.gov.saude.assinador;

import br.gov.saude.assinador.cli.AssinadorCli;
import br.gov.saude.assinador.servidor.AssinadorServidor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Arrays;

public class App {
    private static final Logger logger = LoggerFactory.getLogger(App.class);

    public static void main(String[] args) {
        logger.info("assinador iniciado");
        if (args.length > 0 && "server".equals(args[0])) {
            String[] resto = Arrays.copyOfRange(args, 1, args.length);
            AssinadorServidor.main(resto);
            return;
        }
        AssinadorCli.main(args);
    }
}
