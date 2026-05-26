package br.gov.saude.assinador.servidor;

import br.gov.saude.assinador.servico.FakeSignatureService;
import br.gov.saude.assinador.servico.SignatureService;
import io.javalin.Javalin;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Duration;
import java.time.Instant;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Bootstrap do modo servidor HTTP.
 *
 * Argumentos suportados:
 * <ul>
 *   <li>{@code --porta N} — porta (default 8088)</li>
 *   <li>{@code --idle-timeout-minutes N} — auto-shutdown apos N min sem requisicoes
 *       (default 0 = desativado). Atende US-01 ultimo criterio da especificacao.</li>
 * </ul>
 */
public final class AssinadorServidor {

    private static final Logger logger = LoggerFactory.getLogger(AssinadorServidor.class);
    private static final int PORTA_PADRAO = 8088;

    private AssinadorServidor() {}

    public static Javalin iniciar(int porta, SignatureService servico) {
        return iniciar(porta, servico, Duration.ZERO);
    }

    /**
     * Inicia o servidor com auto-shutdown opcional por inatividade.
     *
     * @param idleTimeout duracao maxima sem requisicoes antes de desligar.
     *                    {@code Duration.ZERO} ou negativa desativa o timeout.
     */
    public static Javalin iniciar(int porta, SignatureService servico, Duration idleTimeout) {
        AtomicReference<Instant> ultimoAcesso = new AtomicReference<>(Instant.now());
        Javalin app = Javalin.create(cfg -> cfg.showJavalinBanner = false);
        boolean timeoutAtivo = idleTimeout != null
                && !idleTimeout.isZero() && !idleTimeout.isNegative();
        if (timeoutAtivo) {
            app.before(ctx -> ultimoAcesso.set(Instant.now()));
        }
        new SignatureController(servico).registrar(app);
        app.start(porta);
        logger.info("assinador HTTP iniciado na porta {}", app.port());

        if (timeoutAtivo) {
            agendarAutoShutdown(app, ultimoAcesso, idleTimeout);
            logger.info("auto-shutdown ativo: desliga apos {} sem requisicoes", idleTimeout);
        }
        return app;
    }

    private static void agendarAutoShutdown(Javalin app,
                                            AtomicReference<Instant> ultimoAcesso,
                                            Duration idleTimeout) {
        ScheduledExecutorService scheduler = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = new Thread(r, "assinador-idle-watcher");
            t.setDaemon(true);
            return t;
        });
        long checkSegundos = Math.max(5, Math.min(30, idleTimeout.toSeconds() / 4));
        scheduler.scheduleAtFixedRate(() -> {
            Duration ocioso = Duration.between(ultimoAcesso.get(), Instant.now());
            if (ocioso.compareTo(idleTimeout) >= 0) {
                logger.info("auto-shutdown: ocioso por {} (>= timeout {}). Desligando.",
                        ocioso, idleTimeout);
                try {
                    app.stop();
                } finally {
                    scheduler.shutdown();
                    new Thread(() -> System.exit(0), "assinador-exit").start();
                }
            }
        }, checkSegundos, checkSegundos, TimeUnit.SECONDS);
    }

    public static void main(String[] args) {
        int porta = PORTA_PADRAO;
        Duration idle = Duration.ZERO;
        for (int i = 0; i < args.length; i++) {
            if ("--porta".equals(args[i]) && i + 1 < args.length) {
                try {
                    porta = Integer.parseInt(args[i + 1]);
                } catch (NumberFormatException ex) {
                    System.err.println("porta invalida: " + args[i + 1]);
                    System.exit(2);
                }
            } else if ("--idle-timeout-minutes".equals(args[i]) && i + 1 < args.length) {
                try {
                    int min = Integer.parseInt(args[i + 1]);
                    if (min > 0) {
                        idle = Duration.ofMinutes(min);
                    }
                } catch (NumberFormatException ex) {
                    System.err.println("idle-timeout-minutes invalido: " + args[i + 1]);
                    System.exit(2);
                }
            }
        }
        iniciar(porta, new FakeSignatureService(), idle);
    }
}
