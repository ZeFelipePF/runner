package br.gov.saude.assinador.cli;

import br.gov.saude.assinador.erro.MapeadorErro;
import br.gov.saude.assinador.erro.RespostaErro;
import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import br.gov.saude.assinador.servico.FakeSignatureService;
import br.gov.saude.assinador.servico.SignatureService;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.io.InputStream;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Map;

/**
 * Modo CLI do assinador.jar.
 * Subcomandos: sign, validate. Entrada via --input <arquivo> ou stdin (default).
 * Saida JSON em stdout (sucesso) ou stderr (erro), com codigo de saida nao-zero em falhas.
 */
public final class AssinadorCli {

    private static final String USO = """
            Uso: assinador <comando> [--input <arquivo>]

            Comandos:
              sign       Cria assinatura FHIR a partir de payload JSON
              validate   Valida assinatura FHIR a partir de payload JSON

            Opcoes:
              --input <arquivo>   Le payload do arquivo (default: stdin)
              -h, --help          Exibe esta mensagem
            """;

    private final ObjectMapper json = new ObjectMapper();
    private final SignatureService servico;
    private final InputStream stdin;
    private final PrintStream stdout;
    private final PrintStream stderr;

    public AssinadorCli(SignatureService servico, InputStream stdin, PrintStream stdout, PrintStream stderr) {
        this.servico = servico;
        this.stdin = stdin;
        this.stdout = stdout;
        this.stderr = stderr;
    }

    public int executar(String[] args) {
        if (args.length == 0 || ehAjuda(args[0])) {
            stdout.println(USO);
            return 0;
        }

        String comando = args[0];

        try {
            String inputArg = lerFlagInput(args);
            Map<String, Object> payload = lerPayload(inputArg);
            Map<String, Object> resposta = switch (comando) {
                case "sign" -> new AcaoAssinar(servico).executar(payload);
                case "validate" -> new AcaoValidar(servico).executar(payload);
                default -> throw new AssinadorException(Codigo.PARAM_INVALIDO,
                        "comando '" + comando + "' desconhecido. Use 'sign' ou 'validate'.");
            };
            stdout.println(json.writeValueAsString(resposta));
            return 0;
        } catch (AssinadorException ex) {
            return falhar(RespostaErro.de(ex), MapeadorErro.exitCode(ex.getCodigo()));
        } catch (IOException ex) {
            return falhar(RespostaErro.de(Codigo.PARAM_INVALIDO, "JSON invalido: " + ex.getMessage()),
                    MapeadorErro.exitCode(Codigo.PARAM_INVALIDO));
        } catch (RuntimeException ex) {
            return falhar(RespostaErro.de(Codigo.ERRO_INTERNO, ex.getMessage()),
                    MapeadorErro.exitCode(Codigo.ERRO_INTERNO));
        }
    }

    private int falhar(RespostaErro erro, int codigoSaida) {
        try {
            stderr.println(json.writeValueAsString(erro));
        } catch (IOException ignored) {
            stderr.println("{\"error\":\"" + erro.error() + "\",\"message\":\"\"}");
        }
        return codigoSaida;
    }

    private static boolean ehAjuda(String s) {
        return "-h".equals(s) || "--help".equals(s) || "help".equals(s);
    }

    private static String lerFlagInput(String[] args) {
        for (int i = 1; i < args.length; i++) {
            if ("--input".equals(args[i])) {
                if (i + 1 >= args.length) {
                    throw new AssinadorException(Codigo.PARAM_AUSENTE,
                            "flag '--input' requer um caminho (ou '-' para stdin)");
                }
                return args[i + 1];
            }
        }
        return null;
    }

    private Map<String, Object> lerPayload(String inputArg) throws IOException {
        byte[] bytes;
        if (inputArg == null || "-".equals(inputArg)) {
            bytes = stdin.readAllBytes();
        } else {
            Path p = Path.of(inputArg);
            if (!Files.isRegularFile(p)) {
                throw new AssinadorException(Codigo.PARAM_INVALIDO,
                        "arquivo de entrada nao encontrado: " + inputArg);
            }
            bytes = Files.readAllBytes(p);
        }
        if (bytes.length == 0) {
            throw new AssinadorException(Codigo.PARAM_AUSENTE, "payload vazio");
        }
        return json.readValue(bytes, new TypeReference<>() {});
    }

    public static void main(String[] args) {
        int code = new AssinadorCli(new FakeSignatureService(), System.in, System.out, System.err)
                .executar(args);
        if (code != 0) System.exit(code);
    }
}
