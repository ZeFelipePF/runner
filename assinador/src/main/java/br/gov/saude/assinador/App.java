package br.gov.saude.assinador;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class App {
    private static final Logger logger = LoggerFactory.getLogger(App.class);

    public static void main(String[] args) {
        logger.info("assinador iniciado");

        if (args.length == 0) {
            System.out.println("{\"version\":\"0.1.0\",\"status\":\"skeleton\"}");
            return;
        }

        logger.info("comando recebido: {}", args[0]);
        System.err.println("{\"error\":\"NOT_IMPLEMENTED\",\"message\":\"Comandos serao implementados na Sprint 2\"}");
        System.exit(1);
    }
}
