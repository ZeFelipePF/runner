package br.gov.saude.assinador.servidor;

import br.gov.saude.assinador.servico.FakeSignatureService;
import br.gov.saude.assinador.servico.SignatureService;
import io.javalin.Javalin;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Bootstrap do modo servidor HTTP.
 * Recebe a porta via argumento; default 8088.
 */
public final class AssinadorServidor {

    private static final Logger logger = LoggerFactory.getLogger(AssinadorServidor.class);
    private static final int PORTA_PADRAO = 8088;

    private AssinadorServidor() {}

    public static Javalin iniciar(int porta, SignatureService servico) {
        Javalin app = Javalin.create(cfg -> cfg.showJavalinBanner = false);
        new SignatureController(servico).registrar(app);
        app.start(porta);
        logger.info("assinador HTTP iniciado na porta {}", app.port());
        return app;
    }

    public static void main(String[] args) {
        int porta = PORTA_PADRAO;
        for (int i = 0; i < args.length; i++) {
            if ("--porta".equals(args[i]) && i + 1 < args.length) {
                try {
                    porta = Integer.parseInt(args[i + 1]);
                } catch (NumberFormatException ex) {
                    System.err.println("porta invalida: " + args[i + 1]);
                    System.exit(2);
                }
            }
        }
        iniciar(porta, new FakeSignatureService());
    }
}
