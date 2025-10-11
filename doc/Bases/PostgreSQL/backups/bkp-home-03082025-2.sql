--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.4

-- Started on 2025-08-03 17:45:34 UTC

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 3420 (class 1262 OID 16384)
-- Name: assjurdb; Type: DATABASE; Schema: -; Owner: assjurpg
--

CREATE DATABASE assjurdb WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.UTF-8';


ALTER DATABASE assjurdb OWNER TO assjurpg;

\connect assjurdb

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 5 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: assjurpg
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO assjurpg;

--
-- TOC entry 217 (class 1259 OID 16393)
-- Name: contexto_id_ctxt_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.contexto_id_ctxt_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.contexto_id_ctxt_seq OWNER TO assjurpg;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 218 (class 1259 OID 16394)
-- Name: contexto; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.contexto (
    id_ctxt integer DEFAULT nextval('public.contexto_id_ctxt_seq'::regclass) NOT NULL,
    nr_proc character varying(24) NOT NULL,
    juizo character varying(255) NOT NULL,
    classe character varying(255) NOT NULL,
    assunto character varying(255) NOT NULL,
    prompt_tokens integer,
    completion_tokens integer,
    dt_inc date NOT NULL,
    status character(1) DEFAULT 'S'::bpchar NOT NULL
);


ALTER TABLE public.contexto OWNER TO assjurpg;

--
-- TOC entry 219 (class 1259 OID 16401)
-- Name: contexto_id_proc_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.contexto_id_proc_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.contexto_id_proc_seq OWNER TO assjurpg;

--
-- TOC entry 220 (class 1259 OID 16408)
-- Name: documentos_id_doc_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.documentos_id_doc_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.documentos_id_doc_seq OWNER TO assjurpg;

--
-- TOC entry 221 (class 1259 OID 16409)
-- Name: prompts; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.prompts (
    id_prompt integer NOT NULL,
    id_nat integer NOT NULL,
    id_doc integer NOT NULL,
    id_classe integer NOT NULL,
    id_assunto integer NOT NULL,
    nm_desc character varying(255),
    txt_prompt text,
    dt_inc date NOT NULL,
    status character(1) DEFAULT 'S'::bpchar NOT NULL
);


ALTER TABLE public.prompts OWNER TO assjurpg;

--
-- TOC entry 222 (class 1259 OID 16415)
-- Name: sessions; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.sessions (
    session_id integer NOT NULL,
    user_id integer,
    model character varying(30) NOT NULL,
    prompt_tokens integer,
    completion_tokens integer,
    total_tokens integer,
    session_start timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    session_end timestamp without time zone
);


ALTER TABLE public.sessions OWNER TO assjurpg;

--
-- TOC entry 223 (class 1259 OID 16419)
-- Name: sessions_session_id_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.sessions_session_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sessions_session_id_seq OWNER TO assjurpg;

--
-- TOC entry 3421 (class 0 OID 0)
-- Dependencies: 223
-- Name: sessions_session_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.sessions_session_id_seq OWNED BY public.sessions.session_id;


--
-- TOC entry 224 (class 1259 OID 16420)
-- Name: tab_prompts_id_prompt_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.tab_prompts_id_prompt_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tab_prompts_id_prompt_seq OWNER TO assjurpg;

--
-- TOC entry 3422 (class 0 OID 0)
-- Dependencies: 224
-- Name: tab_prompts_id_prompt_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.tab_prompts_id_prompt_seq OWNED BY public.prompts.id_prompt;


--
-- TOC entry 225 (class 1259 OID 16422)
-- Name: uploads; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.uploads (
    id_file integer NOT NULL,
    id_ctxt integer NOT NULL,
    nm_file_new character varying(255),
    nm_file_ori character varying(255),
    sn_autos character(1) DEFAULT 'N'::bpchar NOT NULL,
    dt_inc date NOT NULL,
    status character(1) DEFAULT 'S'::bpchar NOT NULL
);


ALTER TABLE public.uploads OWNER TO assjurpg;

--
-- TOC entry 226 (class 1259 OID 16429)
-- Name: temp_uploadfiles_id_file_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.temp_uploadfiles_id_file_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.temp_uploadfiles_id_file_seq OWNER TO assjurpg;

--
-- TOC entry 3423 (class 0 OID 0)
-- Dependencies: 226
-- Name: temp_uploadfiles_id_file_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.temp_uploadfiles_id_file_seq OWNED BY public.uploads.id_file;


--
-- TOC entry 227 (class 1259 OID 16430)
-- Name: users; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    userrole character varying(10) NOT NULL,
    username character varying(20) NOT NULL,
    password character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.users OWNER TO assjurpg;

--
-- TOC entry 228 (class 1259 OID 16436)
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_user_id_seq OWNER TO assjurpg;

--
-- TOC entry 3424 (class 0 OID 0)
-- Dependencies: 228
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- TOC entry 3234 (class 2604 OID 16439)
-- Name: prompts id_prompt; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.prompts ALTER COLUMN id_prompt SET DEFAULT nextval('public.tab_prompts_id_prompt_seq'::regclass);


--
-- TOC entry 3236 (class 2604 OID 16440)
-- Name: sessions session_id; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.sessions ALTER COLUMN session_id SET DEFAULT nextval('public.sessions_session_id_seq'::regclass);


--
-- TOC entry 3238 (class 2604 OID 16441)
-- Name: uploads id_file; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.uploads ALTER COLUMN id_file SET DEFAULT nextval('public.temp_uploadfiles_id_file_seq'::regclass);


--
-- TOC entry 3241 (class 2604 OID 16442)
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- TOC entry 3404 (class 0 OID 16394)
-- Dependencies: 218
-- Data for Name: contexto; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.contexto (id_ctxt, nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc, status) FROM stdin;
32	30011390420258060167	3ª VARA CIVEL DA COMARCA DE SOBRAL	Mandado de Segurança Cível	Liberação de mercadorias	3774	839	2025-07-22	S
31	02029414120248060167	GADES - RAIMUNDO NONATO SILVA SANTOS	Apelação Cível	Práticas Abusivas	900512	37933	2025-06-19	S
\.


--
-- TOC entry 3407 (class 0 OID 16409)
-- Dependencies: 221
-- Data for Name: prompts; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.prompts (id_prompt, id_nat, id_doc, id_classe, id_assunto, nm_desc, txt_prompt, dt_inc, status) FROM stdin;
2	1	1000	1000	1000	Analisar e Identificar peças processuais	🧠 Prompt Completo para Extração de Dados Jurídicos em JSON\n⚖️ OBJETIVO GERAL\nVocê receberá um documento jurídico (ex.: petição inicial, contestação, decisão etc.) e deverá extrair as informações relevantes de forma literal e fiel ao conteúdo, preenchendo o JSON adequado de acordo com o tipo de peça identificada.\n\n🚨 REGRAS GERAIS\nJamais invente, deduza ou complete informações ausentes.\n\nUse linguagem formal e jurídica.\n\nPreencha todos os campos obrigatórios. Caso a informação não conste no documento, escreva: "informação não identificada no documento".\n\nMantenha consistência entre os campos (ex: pedidos, valores, fundamentos, jurisprudência).\n\nNão inclua comentários fora do JSON.\n\nNão use blocos de código, como ```json.\n\nResponda somente com o conteúdo do JSON gerado.\n\n🔍 SOBRE O CAMPO id_pje\nTrata-se de um número de exatamente 9 dígitos, que aparece no rodapé próximo a: Num. ######### - Pág.\n\nExtraia somente os 9 dígitos numéricos.\n\nExemplo: Num. 124984094 - Pág. 2 → "124984094"\n\nCaso não apareça nesse formato, use: "id_pje não identificado".\n\n✅ CHECKLIST FINAL\n Todos os campos obrigatórios preenchidos?\n\n Nenhuma informação presumida?\n\n Termos jurídicos mantidos com exatidão?\n\n Valores, datas e fundamentos incluídos conforme aparecem no texto?\n\n Nenhuma omissão de jurisprudência, doutrina ou normativos citados?\n\n\n\n## 🧩 TABELA DE TIPOS DE DOCUMENTOS\n[\n  { "key": 1, "description": "Petição inicial" },\n  { "key": 2, "description": "Contestação" },\n  { "key": 3, "description": "Réplica" },\n  { "key": 4, "description": "Despacho" }, \n  { "key": 5, "description": "Petição" },\n  { "key": 6, "description": "Decisão" },\n  { "key": 7, "description": "Sentença" },\n  { "key": 8, "description": "Embargos de declaração" },\n  { "key": 9, "description": "Recurso de Apelação" },\n  { "key": 10, "description": "Contra-razões" },\n  { "key": 11, "description": "Procuração" },\n  { "key": 12, "description": "Rol de Testemunhas" },\n  { "key": 13, "description": "Contrato" },\n  { "key": 14, "description": "Laudo Pericial" },\n  { "key": 15, "description": "Termo de Audiência" },\n  { "key": 16, "description": "Parecer do Ministério Público" },\n  { "key": 1000, "description": "Autos Processuais" }\n]\n\n\n## 📦 MODELOS JSON POR TIPO DE DOCUMENTO\n\n### a) Petição Inicial\n{\n  "tipo": { "key": 1, "description": "Petição inicial" },\n  "processo": "string",\n  "id_pje": "string",\n  "natureza": {\n    "nome_juridico": "string"\n  },\n  "partes": {\n    "autor": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "reu": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "fatos": "string",\n  "preliminares": [\n    "string"\n  ],\n  "atos_normativos": [\n    "string"\n  ],\n  "jurisprudencia": {\n    "sumulas": [ "string" ],\n    "acordaos": [\n      {\n        "tribunal": "string",\n        "processo": "string",\n        "ementa": "string",\n        "relator": "string",\n        "data": "string"\n      }\n    ]\n  },\n  "doutrina": [ "string" ],\n  "pedidos": [\n    "string"\n  ],\n  "tutela_provisoria": {\n    "detalhes": "string"\n  },\n  "provas": [\n    "string"\n  ],\n  "rol_testemunhas": [ "string" ],\n  "valor_da_causa": "string",\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n### b) Contestação\n\n{\n  "tipo": { "key": 2, "description": "Contestação" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": {\n    "autor": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "reu": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "fatos": "string",\n  "preliminares": [\n    "string"\n  ],\n  "atos_normativos": [ "string" ],\n  "jurisprudencia": {\n    "sumulas": [ ],\n    "acordaos": [ ]\n  },\n  "doutrina": [ ],\n  "pedidos": [ "string" ],\n  "tutela_provisoria": {\n    "detalhes": "string"\n  },\n  "questoes_controvertidas": [ "string" ],\n  "provas": [ ],\n  "rol_testemunhas": [ ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### c) Réplica\n\n{\n  "tipo": { "key": 3, "description": "Réplica" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes_peticionantes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "fatos": "string",\n  "questoes_controvertidas": [ "string" ],\n  "pedidos": [ "string" ],\n  "provas": [ "string" ],\n  "rol_testemunhas": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### d) Petição\n\n{\n  "tipo": { "key": 5, "description": "Petição" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes_peticionantes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "causaDePedir": "string",\n  "pedidos": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### e) Despacho\n\n{\n  "tipo": { "key": 4, "description": "Despacho" },\n  "processo": "string",\n  "id_pje": "string",\n  "conteudo": [ "string" ],\n  "deliberado": [\n    {\n      "finalidade": "string",\n      "destinatario": "string",\n      "prazo": "string"\n    }\n  ],\n  "juiz": {\n    "nome": "string"\n  }\n}\n\n### f) Decisão\n{\n  "tipo": { "key": 6, "description": "Decisão" },\n  "processo": "string",\n  "id_pje": "string",\n  "conteudo": [ "string" ],\n  "deliberado": [\n    {\n      "finalidade": "string",\n      "destinatario": "string",\n      "prazo": "string"\n    }\n  ],\n  "juiz": {\n    "nome": "string"\n  }\n}\n\n### h) Sentença\n\n{\n  "tipo": { "key": 7, "description": "Sentença" },\n  "processo": "string",\n  "id_pje": "string",\n  "preliminares": [\n    {\n      "assunto": "string",\n      "decisao": "string"\n    }\n  ],\n  "fundamentos": [\n    {\n      "texto": "string",\n      "provas": [ "string" ]\n    }\n  ],\n  "conclusao": [\n    {\n      "resultado": "string",\n      "destinatario": "string",\n      "prazo": "string",\n      "decisao": "string"\n    }\n  ],\n  "juiz": {\n    "nome": "string"\n  }\n}\n\n### i) embargos de declaração\n\n{\n  "tipo": { "key": 8, "description": "Embargos de declaração" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": {\n    "recorrentes": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "recorridos": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "juizoDestinatario": "string",\n  "causaDePedir": "string",\n  "pedidos": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### i) recurso de apelação\n\n{\n  "tipo": { "key": 9, "description": "Recurso de Apelação" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": {\n    "recorrentes": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "recorridos": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "juizoDestinatario": "string",\n  "causaDePedir": "string",\n  "pedidos": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n### j) Procuração\n\n{\n  "tipo": { "key": 11, "description": "Procuração" },\n  "processo": "string",\n  "id_pje": "string",\n  "outorgantes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ],\n  "poderes": "string"\n}\n\n\n### j) Rol de testemunhas\n\n{\n  "tipo": { "key": 12, "description": "Rol de Testemunhas" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "testemunhas": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### j) laudo pericial\n\n{\n  "tipo": { "key": 14, "description": "Laudo Pericial" },\n  "processo": "string",\n  "id_pje": "string",\n  "peritos": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "conclusoes": "string"\n}\n\n### l) termo de audiência\n\n{\n  "tipo": { "key": 15, "description": "Termo de audiência" },\n  "processo": "string",\n  "id_pje": "string",\n  "local": "string",\n  "data": "string",\n  "hora": "string",\n  "presentes": [\n    {\n      "nome": "string",\n      "qualidade": "juiz, requerente, requerido, advogado, conciliador, acadêmico, estudante etc"\n    }\n  ],\n  "descricao": "Após o apregoamento das partes, o senhor Conciliador verificou a presença das partes acima citadas e considerou aberto o ato audiencial. Observou que há contestação às fls.183/200 dos presentes autos.",\n  "manifestacoes": [\n    {\n      "nome": "string",\n      "manifestacao": "string"\n    }\n  ]\n}\n\nSe algum campo não for encontrado no documento, use "informação não identificada no documento" como valor.\n	2025-07-25	S
4	3	1000	1	1	Prompt para análise de julgamento	Você é um assistente jurídico que analise processos e elabora sentenças judiciais.\n\nVocê deve buscar os documentos jurídicos (ex.: petição inicial, contestação, decisão etc.) por meio das funções  e deverá extrair as informações relevantes de forma literal e fiel ao conteúdo.\n\nJamais invente, deduza ou complete informações ausentes.\n\nUse linguagem formal e jurídica.\n\nPor favor, responda sempre no seguinte formato JSON:\n{\n  "tipo_resp": "<um dos valores inteiro da tabela Tipos de resposta válidos>",\n  "texto": "<a resposta textual correspondente>"\n}\n\nTipos de resposta válidos:\n1 - Chat\n2000 - Análise\n2001 - Sentenças\n\nsystemMessage := "Você é um assistente que deve responder sempre no formato JSON, com os campos tipo_resp e texto. Exemplo: {\\"tipo_resp\\":\\"1\\", \\"texto\\":\\"Sua resposta aqui\\"}. Não escreva nada fora do JSON."\n\nNão inclua texto fora desse JSON. Apenas o JSON completo.\n\nSe for pedida a elaboração de uma sentença, peça ao usuários as seguintes informações essenciais e aguarde:\n\n1. Qual é a conclusão da sentença? \n2. Os fatos foram provados? \n\nSe alguma dessas informações não estiver presente, peça que o usuário  as informe e NÃO chame nenhuma função para economizar tokens.\nSe as informações acima estiverem presentes, analise os documentos do processo e elabore a sentença.	2025-08-02	S
\.


--
-- TOC entry 3408 (class 0 OID 16415)
-- Dependencies: 222
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.sessions (session_id, user_id, model, prompt_tokens, completion_tokens, total_tokens, session_start, session_end) FROM stdin;
1	1	OpenAI	911341	40100	951441	2025-01-10 18:35:38.198973	\N
\.


--
-- TOC entry 3411 (class 0 OID 16422)
-- Dependencies: 225
-- Data for Name: uploads; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.uploads (id_file, id_ctxt, nm_file_new, nm_file_ori, sn_autos, dt_inc, status) FROM stdin;
\.


--
-- TOC entry 3413 (class 0 OID 16430)
-- Dependencies: 227
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.users (user_id, userrole, username, password, email, created_at) FROM stdin;
1	admin	aldenor	$2a$10$lXKdvjgcnkPKvcYzZCea7uh3CXjEim/IYOcaEauCPi3sXsZ7eor9m	aldenor.oliveira@uol.com.br	2025-01-10 15:33:49.838908
2	user	secretaria3c	$2a$10$JaQoxdo3HW51RHRs0qWZk.whE2f6UD6VNwBR.mMvQF3eRCnopMcAu	aldenor.oliveira2@uol.com.br	2025-03-09 05:59:17.191632
\.


--
-- TOC entry 3425 (class 0 OID 0)
-- Dependencies: 217
-- Name: contexto_id_ctxt_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.contexto_id_ctxt_seq', 32, true);


--
-- TOC entry 3426 (class 0 OID 0)
-- Dependencies: 219
-- Name: contexto_id_proc_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.contexto_id_proc_seq', 1, false);


--
-- TOC entry 3427 (class 0 OID 0)
-- Dependencies: 220
-- Name: documentos_id_doc_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.documentos_id_doc_seq', 1, true);


--
-- TOC entry 3428 (class 0 OID 0)
-- Dependencies: 223
-- Name: sessions_session_id_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.sessions_session_id_seq', 1, true);


--
-- TOC entry 3429 (class 0 OID 0)
-- Dependencies: 224
-- Name: tab_prompts_id_prompt_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.tab_prompts_id_prompt_seq', 4, true);


--
-- TOC entry 3430 (class 0 OID 0)
-- Dependencies: 226
-- Name: temp_uploadfiles_id_file_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.temp_uploadfiles_id_file_seq', 408, true);


--
-- TOC entry 3431 (class 0 OID 0)
-- Dependencies: 228
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.users_user_id_seq', 2, true);


--
-- TOC entry 3244 (class 2606 OID 16447)
-- Name: contexto contexto_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.contexto
    ADD CONSTRAINT contexto_pkey PRIMARY KEY (id_ctxt);


--
-- TOC entry 3248 (class 2606 OID 16449)
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (session_id);


--
-- TOC entry 3246 (class 2606 OID 16451)
-- Name: prompts tab_prompts_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.prompts
    ADD CONSTRAINT tab_prompts_pkey PRIMARY KEY (id_prompt);


--
-- TOC entry 3250 (class 2606 OID 16455)
-- Name: uploads temp_uploadfiles_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.uploads
    ADD CONSTRAINT temp_uploadfiles_pkey PRIMARY KEY (id_file);


--
-- TOC entry 3252 (class 2606 OID 16457)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 3254 (class 2606 OID 16459)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3256 (class 2606 OID 16461)
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- TOC entry 3257 (class 2606 OID 16462)
-- Name: uploads temp_uploadfiles_id_ctxt_fkey; Type: FK CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.uploads
    ADD CONSTRAINT temp_uploadfiles_id_ctxt_fkey FOREIGN KEY (id_ctxt) REFERENCES public.contexto(id_ctxt);


-- Completed on 2025-08-03 17:45:35 UTC

--
-- PostgreSQL database dump complete
--

