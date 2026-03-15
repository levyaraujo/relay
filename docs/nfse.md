# Guia de Integração com a API NFS-e (Sistema Nacional)

> Documento baseado nos manuais oficiais do gov.br — versão 1.0 (fev/2026).

---

## Sumário

1. [Visão Geral](#1-visão-geral)
2. [Autenticação](#2-autenticação)
3. [Ambientes](#3-ambientes)
4. [API Parâmetros Municipais](#4-api-parâmetros-municipais)
5. [API NFS-e (Emissão e Consulta)](#5-api-nfs-e-emissão-e-consulta)
6. [API DPS (Declaração de Prestação de Serviço)](#6-api-dps)
7. [API Eventos](#7-api-eventos)
8. [API ADN — Distribuição para Contribuintes](#8-api-adn--distribuição-para-contribuintes)
9. [API DANFSE](#9-api-danfse)
10. [API CNC (Cadastro Nacional de Contribuintes)](#10-api-cnc)
11. [Fluxo Completo de Emissão](#11-fluxo-completo-de-emissão)
12. [Referências e Downloads](#12-referências-e-downloads)

---

## 1. Visão Geral

O **Sistema Nacional NFS-e** é a plataforma federal brasileira para emissão, consulta e gestão de Notas Fiscais de Serviço Eletrônicas. Ele é composto por múltiplas APIs REST que cobrem todo o ciclo de vida de uma NFS-e:

| Componente | Função |
|---|---|
| **Sefin Nacional** | Emissor Público Nacional — emissão (DPS → NFS-e), consulta e eventos |
| **ADN** | Ambiente de Dados Nacional — distribuição/compartilhamento de DF-e entre contribuintes |
| **CNC** | Cadastro Nacional de Contribuintes — consulta de inscrições municipais |
| **DANFSE** | Geração do Documento Auxiliar da NFS-e (PDF) |
| **Parametrização** | Parâmetros municipais (alíquotas, regimes, benefícios) |

### Conceitos-chave

- **DPS (Declaração de Prestação de Serviço):** documento XML assinado digitalmente pelo contribuinte que é enviado à Sefin Nacional para gerar a NFS-e.
- **NFS-e:** documento fiscal eletrônico gerado pelo sistema a partir de uma DPS válida.
- **Chave de Acesso:** identificador único de 50 dígitos que identifica uma NFS-e.
- **NSU (Número Sequencial Único):** identificador sequencial para distribuição de DF-e pelo ADN.
- **Evento:** registro de um fato relacionado a uma NFS-e (cancelamento, manifestação, etc.).

### Atores

- **Emitente (Prestador):** quem presta o serviço e emite a NFS-e.
- **Tomador:** quem contrata/recebe o serviço.
- **Intermediário:** terceiro envolvido na operação.

---

## 2. Autenticação

Todas as APIs utilizam **certificado digital ICP-Brasil (mTLS)** para autenticação.

### Requisitos

- Certificado digital **e-CNPJ** ou **e-CPF** tipo A1 ou A3 emitido por AC credenciada na ICP-Brasil.
- O CNPJ do certificado deve ter o **mesmo CNPJ Raiz** do contribuinte consultado.
- A conexão é estabelecida via **TLS mútuo** (mutual TLS / client certificate authentication).

### Como funciona

1. O cliente HTTP apresenta o certificado digital na conexão TLS.
2. O servidor valida a cadeia de certificação ICP-Brasil.
3. O CNPJ/CPF extraído do certificado é usado para autorização — o sistema verifica se o solicitante é um ator válido (prestador, tomador ou intermediário) na NFS-e consultada.

### Configuração técnica (Go)

```go
import (
    "crypto/tls"
    "net/http"
)

func newNFSeClient(certFile, keyFile string) (*http.Client, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("load certificate: %w", err)
    }

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
    }

    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: tlsConfig,
        },
    }, nil
}
```

---

## 3. Ambientes

### Produção Restrita (Homologação / Testes)

| API | Base URL |
|---|---|
| Sefin Nacional | `https://sefin.producaorestrita.nfse.gov.br/SefinNacional` |
| ADN (Contribuintes) | `https://adn.producaorestrita.nfse.gov.br/contribuintes` |
| ADN (Municípios) | `https://adn.producaorestrita.nfse.gov.br` |
| CNC | `https://adn.producaorestrita.nfse.gov.br/cnc` |
| DANFSE | `https://adn.producaorestrita.nfse.gov.br/danfse` |
| Parametrização | `https://adn.producaorestrita.nfse.gov.br/parametrizacao` |

**Swagger (Produção Restrita):**
- Contribuintes: `https://adn.producaorestrita.nfse.gov.br/contribuintes/docs/index.html`
- Sefin Nacional: `https://sefin.producaorestrita.nfse.gov.br/API/SefinNacional/docs/index`

### Produção

| API | Base URL |
|---|---|
| Sefin Nacional | `https://sefin.nfse.gov.br/SefinNacional` |
| ADN (Contribuintes) | `https://adn.nfse.gov.br/contribuintes` |
| ADN (Municípios) | `https://adn.nfse.gov.br` |
| CNC | `https://adn.nfse.gov.br/cnc` |
| DANFSE | `https://adn.nfse.gov.br/danfse` |
| Parametrização | `https://adn.nfse.gov.br/parametrizacao` |

**Swagger (Produção):**
- ADN: `https://adn.nfse.gov.br/docs/index.html`
- Sefin Nacional: `https://sefin.nfse.gov.br/SefinNacional/docs/index`

> **Nota:** Os Swaggers requerem certificado digital (mTLS) para acesso.

---

## 4. API Parâmetros Municipais

Consulta as configurações fiscais de cada município antes de emitir uma NFS-e. Essencial para preencher corretamente a DPS.

### 4.1 Consultar convênio do município

```
GET /parametros_municipais/{codigoMunicipio}/convenio
```

Retorna os parâmetros do convênio (se o município aderiu ao Sistema Nacional NFS-e e suas configurações).

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `codigoMunicipio` | string | path | Código IBGE do município (7 dígitos) |

**Exemplo:**
```
GET /parametros_municipais/3550308/convenio
```

---

### 4.2 Consultar alíquotas e regimes por serviço

```
GET /parametros_municipais/{codigoMunicipio}/{codigoServico}
```

Retorna alíquotas, regimes especiais de tributação e deduções/reduções por subitem da lista de serviço.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `codigoMunicipio` | string | path | Código IBGE do município (7 dígitos) |
| `codigoServico` | string | path | Código do subitem da lista de serviço |

**Exemplo:**
```
GET /parametros_municipais/3550308/0107
```

---

### 4.3 Consultar retenções de um contribuinte

```
GET /parametros_municipais/{codigoMunicipio}/{cpfCnpj}
```

Retorna as retenções que um contribuinte tem o dever de recolher para um município.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `codigoMunicipio` | string | path | Código IBGE do município |
| `cpfCnpj` | string | path | CPF ou CNPJ do contribuinte |

---

### 4.4 Consultar benefícios municipais

```
GET /parametros_municipais/{codigoMunicipio}/{cpfCnpj}
```

Retorna os benefícios municipais a que um contribuinte tem direito em um município.

> **Nota:** Os endpoints 4.3 e 4.4 compartilham a mesma rota — a distinção é feita internamente pelo sistema com base no contexto da requisição. Consulte o Swagger para detalhes exatos dos query parameters.

---

## 5. API NFS-e (Emissão e Consulta)

Esta é a API principal para **emitir** e **consultar** notas fiscais de serviço.

### 5.1 Emitir NFS-e (criar nota)

```
POST /nfse
Content-Type: application/xml
```

Recebe uma **DPS (Declaração de Prestação de Serviço)** em formato XML assinado digitalmente e gera a NFS-e de forma **síncrona**.

**Fluxo:**
1. O contribuinte monta e assina digitalmente o XML da DPS.
2. Envia via POST para `/nfse`.
3. A API valida as regras de negócio sobre a DPS.
4. Se válida → gera a NFS-e e retorna o XML da NFS-e gerada.
5. Se inválida → retorna mensagem de erro com o motivo da rejeição.

**Request body:** XML da DPS assinado digitalmente (conforme XSD `DPS_v1.01.xsd`).

**Responses:**

| Status | Descrição |
|---|---|
| `200` | NFS-e gerada com sucesso — retorna XML da NFS-e |
| `422` | DPS rejeitada — retorna lista de erros de validação |
| `400` | Requisição malformada |
| `401` | Certificado digital inválido ou ausente |
| `403` | CNPJ do certificado não autorizado |
| `500` | Erro interno do servidor |

**Estrutura simplificada da DPS (XML):**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<DPS xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.01">
  <infDPS Id="DPS...">
    <!-- Identificação da DPS -->
    <tpAmb>1</tpAmb>           <!-- 1=Produção, 2=Homologação -->
    <dhEmi>2026-03-12T10:00:00-03:00</dhEmi>
    <verAplic>RELAY_1.0</verAplic>
    <dCompet>2026-03-12</dCompet>

    <!-- Prestador -->
    <prest>
      <CNPJ>12345678000199</CNPJ>
      <IM>12345</IM>
    </prest>

    <!-- Tomador -->
    <toma>
      <CNPJ>98765432000188</CNPJ>
    </toma>

    <!-- Serviço -->
    <serv>
      <cServ>01.07</cServ>
      <xDescServ>Consultoria em tecnologia</xDescServ>
      <cMunPrestworked>3550308</cMunPrest>
    </serv>

    <!-- Valores -->
    <valores>
      <vServPrest>5000.00</vServPrest>
      <vBC>5000.00</vBC>
      <pAliqISS>2.00</pAliqISS>
      <vISS>100.00</vISS>
    </valores>
  </infDPS>

  <!-- Assinatura digital XML-DSig -->
  <Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
    ...
  </Signature>
</DPS>
```

> **Substituição de NFS-e:** Se a DPS contiver a chave de acesso de uma NFS-e existente, o sistema cancela a nota original (gera Evento de Cancelamento por Substituição) e emite a nota substituta.

---

### 5.2 Consultar NFS-e por chave de acesso

```
GET /nfse/{chaveAcesso}
```

Retorna o XML completo da NFS-e correspondente à chave de acesso.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `chaveAcesso` | string | path | Chave de acesso da NFS-e (50 dígitos) |

**Exemplo:**
```
GET /nfse/NFSe35503081234567800019955001000000001123456789
```

**Responses:**

| Status | Descrição |
|---|---|
| `200` | XML da NFS-e |
| `404` | NFS-e não encontrada |
| `401/403` | Certificado inválido ou não autorizado |

---

## 6. API DPS

Permite recuperar a chave de acesso de uma NFS-e a partir do identificador da DPS.

### Formato do ID da DPS

O identificador é composto por:

| Campo | Tamanho | Descrição |
|---|---|---|
| Código IBGE Município | 7 | Município emissor |
| Tipo Inscrição | 1 | Tipo de inscrição federal |
| Inscrição Federal | 14 | CNPJ (ou CPF com 000 à esquerda) |
| Série DPS | 5 | Série da DPS |
| Número DPS | 15 | Número sequencial da DPS |

### 6.1 Recuperar chave de acesso da NFS-e pela DPS

```
GET /dps/{id}
```

Retorna a chave de acesso da NFS-e correspondente à DPS informada.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `id` | string | path | Identificador da DPS (42 caracteres) |

**Restrição de acesso:** Somente retorna a chave se o CNPJ/CPF do certificado corresponder a um dos atores da NFS-e (Prestador, Tomador ou Intermediário). Caso contrário, a requisição é negada (**sigilo fiscal**).

**Responses:**

| Status | Descrição |
|---|---|
| `200` | Chave de acesso da NFS-e |
| `403` | Solicitante não é ator da NFS-e |
| `404` | DPS não encontrada |

---

### 6.2 Verificar existência de NFS-e pela DPS

```
HEAD /dps/{id}
```

Verifica apenas **se** uma NFS-e foi gerada a partir da DPS, sem retornar a chave de acesso.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `id` | string | path | Identificador da DPS |

**Responses:**

| Status | Descrição |
|---|---|
| `200` | NFS-e existe para esta DPS |
| `404` | Nenhuma NFS-e gerada para esta DPS |

> **Nota:** Este método é aberto para qualquer usuário com certificado digital válido (não exige ser ator da NFS-e).

---

## 7. API Eventos

Registra e consulta eventos vinculados a uma NFS-e. Eventos são fatos que podem alterar ou complementar o status de uma nota.

### Tipos de Eventos

| Código | Tipo de Evento | Origem |
|---|---|---|
| `e101101` | Cancelamento de NFS-e | Emitente |
| `e101103` | Cancelamento por Substituição | Sistema (automático) |
| `e105102` | Manifestação — Confirmação do Tomador | Tomador |
| `e105103` | Manifestação — Rejeição do Tomador | Tomador |
| `e105104` | Anulação de NFS-e | Emitente / Município |
| ... | (outros eventos definidos no Anexo II) | Diversos |

> Consulte o **Anexo II — LeiautesRN_Eventos-SNNFSe** para a lista completa de tipos de evento e suas regras de negócio.

### 7.1 Registrar evento (cancelar, manifestar, etc.)

```
POST /nfse/{chaveAcesso}/eventos
Content-Type: application/json
```

Registra um evento vinculado a uma NFS-e. A comunicação usa **JSON** como envelope, mas o conteúdo do evento (DF-e) é um **XML assinado digitalmente**.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `chaveAcesso` | string | path | Chave de acesso da NFS-e alvo |

**Estrutura da mensagem do evento:**

```json
{
  "pedRegEvento": {
    "infPedReg": {
      "tpAmb": 1,
      "verAplic": "RELAY_1.0",
      "dhEvento": "2026-03-12T14:30:00-03:00",
      "nSeqEvento": 1,
      "chNFSe": "NFSe35503081234567800019955001000000001123456789",
      "tpEvento": "e101101",
      "CNPJAutor": "12345678000199",
      "detEvento": {
        "xMotivo": "Erro no valor do serviço prestado"
      }
    },
    "Signature": "... (XML-DSig assinado) ..."
  }
}
```

**Campos obrigatórios do evento:**
- Identificação do autor da mensagem (`CNPJAutor`)
- Identificação do evento (`tpEvento`, `nSeqEvento`)
- Identificação da NFS-e vinculada (`chNFSe`)
- Informações específicas do evento (`detEvento`)
- Assinatura digital XML-DSig (`Signature`)

**Fluxo de processamento:**
1. O solicitante envia o pedido de registro de evento.
2. O sistema valida as regras de negócio e registra ou rejeita.
3. Se aceito, o evento é gerado e vinculado à NFS-e.
4. O sistema retorna aceite ou rejeição ao solicitante.

**Responses:**

| Status | Descrição |
|---|---|
| `200` | Evento registrado com sucesso |
| `422` | Pedido rejeitado — retorna motivo |
| `400` | Requisição malformada |
| `404` | NFS-e não encontrada |

---

### 7.2 Consultar todos os eventos de uma NFS-e

```
GET /nfse/{chaveAcesso}/eventos
```

Retorna **todos** os eventos vinculados à NFS-e.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `chaveAcesso` | string | path | Chave de acesso da NFS-e |

---

### 7.3 Consultar eventos por tipo

```
GET /nfse/{chaveAcesso}/eventos/{tipoEvento}
```

Retorna apenas os eventos do tipo especificado.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `chaveAcesso` | string | path | Chave de acesso da NFS-e |
| `tipoEvento` | string | path | Código identificador do evento (ex: `e101101`) |

---

### 7.4 Consultar evento específico

```
GET /nfse/{chaveAcesso}/eventos/{tipoEvento}/{numSeqEvento}
```

Retorna exatamente um evento pelo tipo e número sequencial.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `chaveAcesso` | string | path | Chave de acesso da NFS-e |
| `tipoEvento` | string | path | Código do tipo de evento |
| `numSeqEvento` | integer | path | Número sequencial do evento (inicia em 1) |

> Se o tipo de evento permite apenas um por NFS-e, `numSeqEvento` deve ser `1`.

---

## 8. API ADN — Distribuição para Contribuintes

Permite que contribuintes consultem DF-e (NFS-e e Eventos) em que figuram como **emitente**, **tomador** ou **intermediário**.

### 8.1 Consultar DF-e por NSU

```
GET /DFe/{NSU}
```

Retorna o documento fiscal correspondente ao NSU informado.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `NSU` | string | path | Número Sequencial Único |

**Observação sobre CNPJ de consulta:**
A consulta pode ser feita com um certificado cujo CNPJ tenha o mesmo CNPJ Raiz do contribuinte. Um parâmetro adicional permite informar um CNPJ de consulta diferente do CNPJ do certificado (validação de CNPJ Raiz é feita).

**Responses:**

| Status | Descrição |
|---|---|
| `200` | DF-e encontrado — retorna o documento |
| `404` | NSU não encontrado |
| `401/403` | Certificado inválido ou CNPJ não autorizado |

---

### 8.2 Consultar eventos por chave de acesso (ADN)

```
GET /NFSe/{ChaveAcesso}/Eventos
```

Retorna todos os eventos vinculados à chave de acesso, distribuídos pelo ADN.

| Parâmetro | Tipo | Local | Descrição |
|---|---|---|---|
| `ChaveAcesso` | string | path | Chave de acesso da NFS-e |

---

## 9. API DANFSE

Gera o **Documento Auxiliar da NFS-e** (representação visual/PDF da nota).

### Base URLs

| Ambiente | URL |
|---|---|
| Produção Restrita | `https://adn.producaorestrita.nfse.gov.br/danfse` |
| Produção | `https://adn.nfse.gov.br/danfse` |

> Consulte o Swagger para endpoints específicos. Tipicamente:
> ```
> GET /danfse/{chaveAcesso}
> ```
> Retorna o PDF do DANFSE para a NFS-e informada.

---

## 10. API CNC

O **Cadastro Nacional de Contribuintes** permite consultar informações cadastrais de prestadores de serviço.

### Base URLs

| Ambiente | URL |
|---|---|
| Produção Restrita | `https://adn.producaorestrita.nfse.gov.br/cnc` |
| Produção | `https://adn.nfse.gov.br/cnc` |

> Consulte o Swagger para endpoints e schemas específicos.

---

## 11. Fluxo Completo de Emissão

```
┌─────────────────────────────────────────────────────────────────┐
│                    FLUXO DE EMISSÃO NFS-e                       │
└─────────────────────────────────────────────────────────────────┘

1. Consultar parâmetros municipais
   GET /parametros_municipais/{codMunicipio}/convenio
   GET /parametros_municipais/{codMunicipio}/{codServico}
        │
        ▼
2. Montar XML da DPS com dados do serviço
   - Preencher prestador, tomador, serviço, valores
   - Usar alíquotas e regimes obtidos no passo 1
        │
        ▼
3. Assinar digitalmente o XML (XML-DSig com certificado ICP-Brasil)
        │
        ▼
4. Enviar DPS para emissão
   POST /nfse
   Body: XML da DPS assinada
        │
        ├── Sucesso (200) → Recebe XML da NFS-e gerada
        │                    Extrair chave de acesso
        │
        └── Rejeição (422) → Corrigir erros e reenviar
        │
        ▼
5. (Opcional) Consultar NFS-e gerada
   GET /nfse/{chaveAcesso}
        │
        ▼
6. (Opcional) Obter DANFSE (PDF)
   GET /danfse/{chaveAcesso}
        │
        ▼
7. (Se necessário) Cancelar NFS-e
   POST /nfse/{chaveAcesso}/eventos
   Body: Evento tipo e101101 (Cancelamento)
        │
        ▼
8. (Se necessário) Substituir NFS-e
   POST /nfse
   Body: Nova DPS com chave de acesso da NFS-e a substituir
   → Sistema cancela a original e emite a substituta
```

---

## 12. Referências e Downloads

### Manuais Oficiais (PDF)

| Documento | URL |
|---|---|
| Manual Contribuintes — APIs do ADN | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/manual-contribuintes-apis-adn-sistema-nacional-nfse.pdf) |
| Manual Contribuintes — Emissor Público API | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/manual-contribuintes-emissor-publico-api-sistema-nacional-nfs-e-v1-2-out2025.pdf) |
| Manual Municípios — ADN API | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/manual-municipios-apis-adn-sistema-nacional-nfs-e-v1-2-out2025.pdf) |
| Manual Municípios — CNC API | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/manual-municipios-cnc-api-sistema-nacional-nfs-e-v1-2-out21025.pdf) |

### Schemas XSD

| Arquivo | URL |
|---|---|
| NFSe-ESQUEMAS_XSD-v1.01 | [Download ZIP](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/nfse-esquemas_xsd-v1-01-20260209.zip) |

### Anexos de Leiaute e Regras de Negócio

| Anexo | Descrição | URL |
|---|---|---|
| Anexo I | Leiaute DPS/NFS-e + Regras de Negócio | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_i-sefin_adn-dps_nfse-snnfse-v1-01-20260209.xlsx) |
| Anexo II | Leiaute Eventos + Regras de Negócio | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_ii-sefin_adn-pedregevt_evt-snnfse-v1-01-20260122.xlsx) |
| Anexo III | CNC — Cadastro Nacional Contribuintes | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_iii-cnc-snnfse-v1-00-20251216.xlsx) |
| Anexo IV | ADN — Ambiente de Dados Nacional | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_iv-adn-snnfse-v1-00-20251216.xlsx) |

### Tabelas de Domínio

| Tabela | Descrição | URL |
|---|---|---|
| Anexo A | Municípios IBGE + Países ISO2 | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_a-municipio_ibge-paises_iso2-v1-00-snnfse-20251210.xlsx) |
| Anexo B | NBS2 — Lista de Serviço Nacional | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_b-nbs2-lista_servico_nacional-snnfse-v1-01-20260122.xlsx) |
| Anexo C | Indicadores de Operação CBS | [Download](https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/anexo_c-indop_ibscbs-snnfse-v1-01-20260122.xlsx) |

### Portal Oficial

- Documentação Técnica: https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica
- FAQ: https://www.gov.br/nfse/pt-br/biblioteca/copy_of_perguntas-frequentes/copy_of_faq-nfs-e

---

## Resumo de Endpoints

| Método | Endpoint | Descrição |
|---|---|---|
| `GET` | `/parametros_municipais/{codMun}/convenio` | Parâmetros de convênio municipal |
| `GET` | `/parametros_municipais/{codMun}/{codServ}` | Alíquotas e regimes por serviço |
| `GET` | `/parametros_municipais/{codMun}/{cpfCnpj}` | Retenções / Benefícios do contribuinte |
| `POST` | `/nfse` | **Emitir NFS-e** (enviar DPS) |
| `GET` | `/nfse/{chaveAcesso}` | Consultar NFS-e por chave de acesso |
| `GET` | `/dps/{id}` | Recuperar chave de acesso via DPS |
| `HEAD` | `/dps/{id}` | Verificar existência de NFS-e via DPS |
| `POST` | `/nfse/{chaveAcesso}/eventos` | **Registrar evento** (cancelar, manifestar) |
| `GET` | `/nfse/{chaveAcesso}/eventos` | Listar todos os eventos de uma NFS-e |
| `GET` | `/nfse/{chaveAcesso}/eventos/{tipo}` | Listar eventos por tipo |
| `GET` | `/nfse/{chaveAcesso}/eventos/{tipo}/{seq}` | Consultar evento específico |
| `GET` | `/DFe/{NSU}` | Distribuição — consultar DF-e por NSU |
| `GET` | `/NFSe/{ChaveAcesso}/Eventos` | Distribuição — eventos por chave (ADN) |
| `GET` | `/danfse/{chaveAcesso}` | Obter DANFSE (PDF) |
