package br.gov.saude.assinador.cli;

import br.gov.saude.assinador.servico.AssinadorException;
import br.gov.saude.assinador.servico.AssinadorException.Codigo;
import br.gov.saude.assinador.servico.FakeSignatureService;
import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class AcaoAssinarTest {

    private final AcaoAssinar acao = new AcaoAssinar(new FakeSignatureService());

    private Map<String, Object> payloadValidoPem() {
        Map<String, Object> p = new HashMap<>();
        p.put("bundle", "{\"resourceType\":\"Bundle\"}");
        p.put("provenance", "{\"resourceType\":\"Provenance\"}");
        p.put("materialCriptografico", Map.of("tipo", "PEM", "chavePrivada", "---PEM---"));
        p.put("certificados", List.of("Y2VydA=="));
        p.put("timestampReferencia", 1_700_000_000L);
        p.put("estrategiaTimestamp", "iat");
        p.put("politicaAssinatura", "urn:policy:v1");
        return p;
    }

    @Test
    void executarRetornaSignatureFhir() {
        Map<String, Object> r = acao.executar(payloadValidoPem());
        assertEquals("Signature", r.get("resourceType"));
        assertEquals("application/jose", r.get("sigFormat"));
        assertEquals(FakeSignatureService.ASSINATURA_SIMULADA, r.get("data"));
        assertEquals("2023-11-14T22:13:20Z", r.get("when"));
    }

    @Test
    void executarPropagaErroDeValidacao() {
        Map<String, Object> p = payloadValidoPem();
        p.remove("bundle");
        AssinadorException ex = assertThrows(AssinadorException.class, () -> acao.executar(p));
        assertEquals(Codigo.PARAM_AUSENTE, ex.getCodigo());
    }

    @Test
    void executarComMaterialPkcs12() {
        Map<String, Object> p = payloadValidoPem();
        p.put("materialCriptografico", Map.of(
                "tipo", "PKCS12", "conteudo", "Y29udA==", "senha", "x", "alias", "minha-chave"));
        Map<String, Object> r = acao.executar(p);
        assertEquals("Signature", r.get("resourceType"));
    }

    @Test
    void executarComMaterialSmartcard() {
        Map<String, Object> p = payloadValidoPem();
        p.put("materialCriptografico", Map.of(
                "tipo", "SMARTCARD", "pin", "1234", "identificador", "alias-A1"));
        Map<String, Object> r = acao.executar(p);
        assertEquals("Signature", r.get("resourceType"));
    }

    @Test
    void executarComMaterialRemote() {
        Map<String, Object> p = payloadValidoPem();
        p.put("materialCriptografico", Map.of(
                "tipo", "REMOTE", "enderecoServico", "https://hsm.example.com"));
        Map<String, Object> r = acao.executar(p);
        assertEquals("Signature", r.get("resourceType"));
    }

    @Test
    void executarComMaterialDesconhecidoUsaTipoComoChave() {
        Map<String, Object> p = payloadValidoPem();
        p.put("materialCriptografico", Map.of("tipo", "INESPERADO", "x", "y"));
        Map<String, Object> r = acao.executar(p);
        assertEquals("Signature", r.get("resourceType"));
    }
}
