--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2025-07-27 23:21:45 UTC

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
-- TOC entry 3485 (class 1262 OID 16384)
-- Name: assjurdb; Type: DATABASE; Schema: -; Owner: assjurpg
--

CREATE DATABASE assjurdb WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';


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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 217 (class 1259 OID 16389)
-- Name: autos; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.autos (
    id_autos integer NOT NULL,
    id_ctxt integer NOT NULL,
    id_nat integer NOT NULL,
    id_pje character varying(20),
    dt_pje date,
    autos_json json,
    dt_inc date NOT NULL,
    status character(1) DEFAULT 'S'::bpchar NOT NULL
);


ALTER TABLE public.autos OWNER TO assjurpg;

--
-- TOC entry 218 (class 1259 OID 16395)
-- Name: autos_id_autos_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.autos_id_autos_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.autos_id_autos_seq OWNER TO assjurpg;

--
-- TOC entry 3486 (class 0 OID 0)
-- Dependencies: 218
-- Name: autos_id_autos_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.autos_id_autos_seq OWNED BY public.autos.id_autos;


--
-- TOC entry 219 (class 1259 OID 16396)
-- Name: contexto_id_ctxt_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.contexto_id_ctxt_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.contexto_id_ctxt_seq OWNER TO assjurpg;

--
-- TOC entry 220 (class 1259 OID 16397)
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
-- TOC entry 221 (class 1259 OID 16404)
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
-- TOC entry 222 (class 1259 OID 16405)
-- Name: docsocr; Type: TABLE; Schema: public; Owner: assjurpg
--

CREATE TABLE public.docsocr (
    id_doc integer NOT NULL,
    id_ctxt integer NOT NULL,
    nm_file_new character varying(255),
    nm_file_ori character varying(255),
    txt_doc text,
    dt_inc date NOT NULL,
    status character(1) DEFAULT 'S'::bpchar NOT NULL
);


ALTER TABLE public.docsocr OWNER TO assjurpg;

--
-- TOC entry 223 (class 1259 OID 16411)
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
-- TOC entry 224 (class 1259 OID 16412)
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
-- TOC entry 225 (class 1259 OID 16418)
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
-- TOC entry 226 (class 1259 OID 16422)
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
-- TOC entry 3487 (class 0 OID 0)
-- Dependencies: 226
-- Name: sessions_session_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.sessions_session_id_seq OWNED BY public.sessions.session_id;


--
-- TOC entry 227 (class 1259 OID 16423)
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
-- TOC entry 3488 (class 0 OID 0)
-- Dependencies: 227
-- Name: tab_prompts_id_prompt_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.tab_prompts_id_prompt_seq OWNED BY public.prompts.id_prompt;


--
-- TOC entry 228 (class 1259 OID 16424)
-- Name: temp_autos_id_doc_seq; Type: SEQUENCE; Schema: public; Owner: assjurpg
--

CREATE SEQUENCE public.temp_autos_id_doc_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.temp_autos_id_doc_seq OWNER TO assjurpg;

--
-- TOC entry 3489 (class 0 OID 0)
-- Dependencies: 228
-- Name: temp_autos_id_doc_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.temp_autos_id_doc_seq OWNED BY public.docsocr.id_doc;


--
-- TOC entry 229 (class 1259 OID 16425)
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
-- TOC entry 230 (class 1259 OID 16432)
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
-- TOC entry 3490 (class 0 OID 0)
-- Dependencies: 230
-- Name: temp_uploadfiles_id_file_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.temp_uploadfiles_id_file_seq OWNED BY public.uploads.id_file;


--
-- TOC entry 231 (class 1259 OID 16433)
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
-- TOC entry 232 (class 1259 OID 16439)
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
-- TOC entry 3491 (class 0 OID 0)
-- Dependencies: 232
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: assjurpg
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- TOC entry 3285 (class 2604 OID 16440)
-- Name: autos id_autos; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.autos ALTER COLUMN id_autos SET DEFAULT nextval('public.autos_id_autos_seq'::regclass);


--
-- TOC entry 3289 (class 2604 OID 16441)
-- Name: docsocr id_doc; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.docsocr ALTER COLUMN id_doc SET DEFAULT nextval('public.temp_autos_id_doc_seq'::regclass);


--
-- TOC entry 3291 (class 2604 OID 16442)
-- Name: prompts id_prompt; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.prompts ALTER COLUMN id_prompt SET DEFAULT nextval('public.tab_prompts_id_prompt_seq'::regclass);


--
-- TOC entry 3293 (class 2604 OID 16443)
-- Name: sessions session_id; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.sessions ALTER COLUMN session_id SET DEFAULT nextval('public.sessions_session_id_seq'::regclass);


--
-- TOC entry 3295 (class 2604 OID 16444)
-- Name: uploads id_file; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.uploads ALTER COLUMN id_file SET DEFAULT nextval('public.temp_uploadfiles_id_file_seq'::regclass);


--
-- TOC entry 3298 (class 2604 OID 16445)
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- TOC entry 3464 (class 0 OID 16389)
-- Dependencies: 217
-- Data for Name: autos; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.autos (id_autos, id_ctxt, id_nat, id_pje, dt_pje, autos_json, dt_inc, status) FROM stdin;
\.


--
-- TOC entry 3467 (class 0 OID 16397)
-- Dependencies: 220
-- Data for Name: contexto; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.contexto (id_ctxt, nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc, status) FROM stdin;
32	30011390420258060167	3ª VARA CIVEL DA COMARCA DE SOBRAL	Mandado de Segurança Cível	Liberação de mercadorias	0	0	2025-07-22	S
31	02029414120248060167	GADES - RAIMUNDO NONATO SILVA SANTOS	Apelação Cível	Práticas Abusivas	97553	8072	2025-06-19	S
\.


--
-- TOC entry 3469 (class 0 OID 16405)
-- Dependencies: 222
-- Data for Name: docsocr; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.docsocr (id_doc, id_ctxt, nm_file_new, nm_file_ori, txt_doc, dt_inc, status) FROM stdin;
\.


--
-- TOC entry 3471 (class 0 OID 16412)
-- Dependencies: 224
-- Data for Name: prompts; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.prompts (id_prompt, id_nat, id_doc, id_classe, id_assunto, nm_desc, txt_prompt, dt_inc, status) FROM stdin;
2	1	1000	1000	1000	Analisar e Identificar peças processuais	🧠 Prompt Completo para Extração de Dados Jurídicos em JSON\n⚖️ OBJETIVO GERAL\nVocê receberá um documento jurídico (ex.: petição inicial, contestação, decisão etc.) e deverá extrair as informações relevantes de forma literal e fiel ao conteúdo, preenchendo o JSON adequado de acordo com o tipo de peça identificada.\n\n🚨 REGRAS GERAIS\nJamais invente, deduza ou complete informações ausentes.\n\nUse linguagem formal e jurídica.\n\nPreencha todos os campos obrigatórios. Caso a informação não conste no documento, escreva: "informação não identificada no documento".\n\nMantenha consistência entre os campos (ex: pedidos, valores, fundamentos, jurisprudência).\n\nNão inclua comentários fora do JSON.\n\nNão use blocos de código, como ```json.\n\nResponda somente com o conteúdo do JSON gerado.\n\n🔍 SOBRE O CAMPO id_pje\nTrata-se de um número de exatamente 9 dígitos, que aparece no rodapé próximo a: Num. ######### - Pág.\n\nExtraia somente os 9 dígitos numéricos.\n\nExemplo: Num. 124984094 - Pág. 2 → "124984094"\n\nCaso não apareça nesse formato, use: "id_pje não identificado".\n\n✅ CHECKLIST FINAL\n Todos os campos obrigatórios preenchidos?\n\n Nenhuma informação presumida?\n\n Termos jurídicos mantidos com exatidão?\n\n Valores, datas e fundamentos incluídos conforme aparecem no texto?\n\n Nenhuma omissão de jurisprudência, doutrina ou normativos citados?\n\n\n\n## 🧩 TABELA DE TIPOS DE DOCUMENTOS\n[\n  { "key": 1, "description": "Petição inicial" },\n  { "key": 2, "description": "Contestação" },\n  { "key": 3, "description": "Réplica" },\n  { "key": 4, "description": "Despacho" }, \n  { "key": 5, "description": "Petição" },\n  { "key": 6, "description": "Decisão" },\n  { "key": 7, "description": "Sentença" },\n  { "key": 8, "description": "Embargos de declaração" },\n  { "key": 9, "description": "Recurso de Apelação" },\n  { "key": 10, "description": "Contra-razões" },\n  { "key": 11, "description": "Procuração" },\n  { "key": 12, "description": "Rol de Testemunhas" },\n  { "key": 13, "description": "Contrato" },\n  { "key": 14, "description": "Laudo Pericial" },\n  { "key": 15, "description": "Termo de Audiência" },\n  { "key": 16, "description": "Parecer do Ministério Público" },\n  { "key": 1000, "description": "Autos Processuais" }\n]\n\n\n## 📦 MODELOS JSON POR TIPO DE DOCUMENTO\n\n### a) Petição Inicial\n{\n  "tipo": { "key": 1, "description": "Petição inicial" },\n  "processo": "string",\n  "id_pje": "string",\n  "natureza": {\n    "nome_juridico": "string"\n  },\n  "partes": {\n    "autor": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "reu": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "fatos": "string",\n  "preliminares": [\n    "string"\n  ],\n  "atos_normativos": [\n    "string"\n  ],\n  "jurisprudencia": {\n    "sumulas": [ "string" ],\n    "acordaos": [\n      {\n        "tribunal": "string",\n        "processo": "string",\n        "ementa": "string",\n        "relator": "string",\n        "data": "string"\n      }\n    ]\n  },\n  "doutrina": [ "string" ],\n  "pedidos": [\n    "string"\n  ],\n  "tutela_provisoria": {\n    "detalhes": "string"\n  },\n  "provas": [\n    "string"\n  ],\n  "rol_testemunhas": [ "string" ],\n  "valor_da_causa": "string",\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n### b) Contestação\n\n{\n  "tipo": { "key": 2, "description": "Contestação" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": {\n    "autor": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "reu": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "fatos": "string",\n  "preliminares": [\n    "string"\n  ],\n  "atos_normativos": [ "string" ],\n  "jurisprudencia": {\n    "sumulas": [ ],\n    "acordaos": [ ]\n  },\n  "doutrina": [ ],\n  "pedidos": [ "string" ],\n  "tutela_provisoria": {\n    "detalhes": "string"\n  },\n  "questoes_controvertidas": [ "string" ],\n  "provas": [ ],\n  "rol_testemunhas": [ ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### c) Réplica\n\n{\n  "tipo": { "key": 3, "description": "Réplica" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes_peticionantes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "fatos": "string",\n  "questoes_controvertidas": [ "string" ],\n  "pedidos": [ "string" ],\n  "provas": [ "string" ],\n  "rol_testemunhas": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### d) Petição\n\n{\n  "tipo": { "key": 5, "description": "Petição" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes_peticionantes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "causaDePedir": "string",\n  "pedidos": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### e) Despacho\n\n{\n  "tipo": { "key": 4, "description": "Despacho" },\n  "processo": "string",\n  "id_pje": "string",\n  "conteudo": [ "string" ],\n  "deliberado": [\n    {\n      "finalidade": "string",\n      "destinatario": "string",\n      "prazo": "string"\n    }\n  ],\n  "juiz": {\n    "nome": "string"\n  }\n}\n\n### f) Decisão\n{\n  "tipo": { "key": 6, "description": "Decisão" },\n  "processo": "string",\n  "id_pje": "string",\n  "conteudo": [ "string" ],\n  "deliberado": [\n    {\n      "finalidade": "string",\n      "destinatario": "string",\n      "prazo": "string"\n    }\n  ],\n  "juiz": {\n    "nome": "string"\n  }\n}\n\n### h) Sentença\n\n{\n  "tipo": { "key": 7, "description": "Sentença" },\n  "processo": "string",\n  "id_pje": "string",\n  "preliminares": [\n    {\n      "assunto": "string",\n      "decisao": "string"\n    }\n  ],\n  "fundamentos": [\n    {\n      "texto": "string",\n      "provas": [ "string" ]\n    }\n  ],\n  "conclusao": [\n    {\n      "resultado": "string",\n      "destinatario": "string",\n      "prazo": "string",\n      "decisao": "string"\n    }\n  ],\n  "juiz": {\n    "nome": "string"\n  }\n}\n\n### i) embargos de declaração\n\n{\n  "tipo": { "key": 8, "description": "Embargos de declaração" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": {\n    "recorrentes": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "recorridos": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "juizoDestinatario": "string",\n  "causaDePedir": "string",\n  "pedidos": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### i) recurso de apelação\n\n{\n  "tipo": { "key": 9, "description": "Recurso de Apelação" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": {\n    "recorrentes": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ],\n    "recorridos": [\n      {\n        "nome": "string",\n        "cpf": "string",\n        "cnpj": "string",\n        "endereco": "string"\n      }\n    ]\n  },\n  "juizoDestinatario": "string",\n  "causaDePedir": "string",\n  "pedidos": [ "string" ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n### j) Procuração\n\n{\n  "tipo": { "key": 11, "description": "Procuração" },\n  "processo": "string",\n  "id_pje": "string",\n  "outorgantes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ],\n  "poderes": "string"\n}\n\n\n### j) Rol de testemunhas\n\n{\n  "tipo": { "key": 12, "description": "Rol de Testemunhas" },\n  "processo": "string",\n  "id_pje": "string",\n  "partes": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "testemunhas": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "advogados": [\n    {\n      "nome": "string",\n      "oab": "string"\n    }\n  ]\n}\n\n\n### j) laudo pericial\n\n{\n  "tipo": { "key": 14, "description": "Laudo Pericial" },\n  "processo": "string",\n  "id_pje": "string",\n  "peritos": [\n    {\n      "nome": "string",\n      "cpf": "string",\n      "cnpj": "string",\n      "endereco": "string"\n    }\n  ],\n  "conclusoes": "string"\n}\n\n### l) termo de audiência\n\n{\n  "tipo": { "key": 15, "description": "Termo de audiência" },\n  "processo": "string",\n  "id_pje": "string",\n  "local": "string",\n  "data": "string",\n  "hora": "string",\n  "presentes": [\n    {\n      "nome": "string",\n      "qualidade": "juiz, requerente, requerido, advogado, conciliador, acadêmico, estudante etc"\n    }\n  ],\n  "descricao": "Após o apregoamento das partes, o senhor Conciliador verificou a presença das partes acima citadas e considerou aberto o ato audiencial. Observou que há contestação às fls.183/200 dos presentes autos.",\n  "manifestacoes": [\n    {\n      "nome": "string",\n      "manifestacao": "string"\n    }\n  ]\n}\n\nSe algum campo não for encontrado no documento, use "informação não identificada no documento" como valor.\n	2025-07-25	S
\.


--
-- TOC entry 3472 (class 0 OID 16418)
-- Dependencies: 225
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.sessions (session_id, user_id, model, prompt_tokens, completion_tokens, total_tokens, session_start, session_end) FROM stdin;
1	1	OpenAI	100374	8072	108446	2025-01-10 18:35:38.198973	\N
\.


--
-- TOC entry 3476 (class 0 OID 16425)
-- Dependencies: 229
-- Data for Name: uploads; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.uploads (id_file, id_ctxt, nm_file_new, nm_file_ori, sn_autos, dt_inc, status) FROM stdin;
\.


--
-- TOC entry 3478 (class 0 OID 16433)
-- Dependencies: 231
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: assjurpg
--

COPY public.users (user_id, userrole, username, password, email, created_at) FROM stdin;
1	admin	aldenor	$2a$10$lXKdvjgcnkPKvcYzZCea7uh3CXjEim/IYOcaEauCPi3sXsZ7eor9m	aldenor.oliveira@uol.com.br	2025-01-10 15:33:49.838908
2	user	secretaria3c	$2a$10$JaQoxdo3HW51RHRs0qWZk.whE2f6UD6VNwBR.mMvQF3eRCnopMcAu	aldenor.oliveira2@uol.com.br	2025-03-09 05:59:17.191632
\.


--
-- TOC entry 3492 (class 0 OID 0)
-- Dependencies: 218
-- Name: autos_id_autos_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.autos_id_autos_seq', 80, true);


--
-- TOC entry 3493 (class 0 OID 0)
-- Dependencies: 219
-- Name: contexto_id_ctxt_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.contexto_id_ctxt_seq', 32, true);


--
-- TOC entry 3494 (class 0 OID 0)
-- Dependencies: 221
-- Name: contexto_id_proc_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.contexto_id_proc_seq', 1, false);


--
-- TOC entry 3495 (class 0 OID 0)
-- Dependencies: 223
-- Name: documentos_id_doc_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.documentos_id_doc_seq', 1, true);


--
-- TOC entry 3496 (class 0 OID 0)
-- Dependencies: 226
-- Name: sessions_session_id_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.sessions_session_id_seq', 1, true);


--
-- TOC entry 3497 (class 0 OID 0)
-- Dependencies: 227
-- Name: tab_prompts_id_prompt_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.tab_prompts_id_prompt_seq', 2, true);


--
-- TOC entry 3498 (class 0 OID 0)
-- Dependencies: 228
-- Name: temp_autos_id_doc_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.temp_autos_id_doc_seq', 829, true);


--
-- TOC entry 3499 (class 0 OID 0)
-- Dependencies: 230
-- Name: temp_uploadfiles_id_file_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.temp_uploadfiles_id_file_seq', 407, true);


--
-- TOC entry 3500 (class 0 OID 0)
-- Dependencies: 232
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: assjurpg
--

SELECT pg_catalog.setval('public.users_user_id_seq', 2, true);


--
-- TOC entry 3301 (class 2606 OID 16451)
-- Name: autos autos_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.autos
    ADD CONSTRAINT autos_pkey PRIMARY KEY (id_autos);


--
-- TOC entry 3303 (class 2606 OID 16453)
-- Name: contexto contexto_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.contexto
    ADD CONSTRAINT contexto_pkey PRIMARY KEY (id_ctxt);


--
-- TOC entry 3309 (class 2606 OID 16455)
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (session_id);


--
-- TOC entry 3307 (class 2606 OID 16457)
-- Name: prompts tab_prompts_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.prompts
    ADD CONSTRAINT tab_prompts_pkey PRIMARY KEY (id_prompt);


--
-- TOC entry 3305 (class 2606 OID 16459)
-- Name: docsocr temp_autos_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.docsocr
    ADD CONSTRAINT temp_autos_pkey PRIMARY KEY (id_doc);


--
-- TOC entry 3311 (class 2606 OID 16461)
-- Name: uploads temp_uploadfiles_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.uploads
    ADD CONSTRAINT temp_uploadfiles_pkey PRIMARY KEY (id_file);


--
-- TOC entry 3313 (class 2606 OID 16463)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 3315 (class 2606 OID 16465)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3317 (class 2606 OID 16467)
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- TOC entry 3318 (class 2606 OID 16468)
-- Name: uploads temp_uploadfiles_id_ctxt_fkey; Type: FK CONSTRAINT; Schema: public; Owner: assjurpg
--

ALTER TABLE ONLY public.uploads
    ADD CONSTRAINT temp_uploadfiles_id_ctxt_fkey FOREIGN KEY (id_ctxt) REFERENCES public.contexto(id_ctxt);


-- Completed on 2025-07-27 23:21:46 UTC

--
-- PostgreSQL database dump complete
--

