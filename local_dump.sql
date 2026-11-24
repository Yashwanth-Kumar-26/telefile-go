--
-- PostgreSQL database dump
--

\restrict IyUhQiPMS2VsHrHRzbuDndl8w3sUimrrmDDpqYHBDceHXOwfDgPrO8LaIUp7s32

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

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
-- Name: batches; Type: TABLE; Schema: public; Owner: siddu
--

CREATE TABLE public.batches (
    id integer NOT NULL,
    user_id bigint,
    unique_link_id character varying(12),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.batches OWNER TO siddu;

--
-- Name: batches_id_seq; Type: SEQUENCE; Schema: public; Owner: siddu
--

CREATE SEQUENCE public.batches_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.batches_id_seq OWNER TO siddu;

--
-- Name: batches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: siddu
--

ALTER SEQUENCE public.batches_id_seq OWNED BY public.batches.id;


--
-- Name: files; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.files (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id bigint NOT NULL,
    file_name text NOT NULL,
    file_size bigint NOT NULL,
    file_type text,
    telegram_file_id text NOT NULL,
    unique_link text NOT NULL,
    password text,
    download_limit integer DEFAULT '-1'::integer,
    download_count integer DEFAULT 0,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    batch_id integer
);


ALTER TABLE public.files OWNER TO postgres;

--
-- Name: pending_payments; Type: TABLE; Schema: public; Owner: siddu
--

CREATE TABLE public.pending_payments (
    id integer NOT NULL,
    user_id bigint,
    plan_name text NOT NULL,
    plan_id integer,
    amount numeric(10,2) NOT NULL,
    screenshot_file_id text,
    status text DEFAULT 'pending'::text,
    created_at timestamp without time zone DEFAULT now(),
    first_name text DEFAULT 'User'::text
);


ALTER TABLE public.pending_payments OWNER TO siddu;

--
-- Name: pending_payments_id_seq; Type: SEQUENCE; Schema: public; Owner: siddu
--

CREATE SEQUENCE public.pending_payments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pending_payments_id_seq OWNER TO siddu;

--
-- Name: pending_payments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: siddu
--

ALTER SEQUENCE public.pending_payments_id_seq OWNED BY public.pending_payments.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    username text,
    name text,
    storage_limit_gb bigint DEFAULT 20,
    storage_used_bytes bigint DEFAULT 0,
    is_banned boolean DEFAULT false,
    referral_code text,
    referred_by bigint,
    created_at timestamp with time zone DEFAULT now(),
    last_active timestamp with time zone DEFAULT now()
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: batches id; Type: DEFAULT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.batches ALTER COLUMN id SET DEFAULT nextval('public.batches_id_seq'::regclass);


--
-- Name: pending_payments id; Type: DEFAULT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments ALTER COLUMN id SET DEFAULT nextval('public.pending_payments_id_seq'::regclass);


--
-- Data for Name: batches; Type: TABLE DATA; Schema: public; Owner: siddu
--

COPY public.batches (id, user_id, unique_link_id, created_at) FROM stdin;
1	1096262005	a1b5be12-498	2026-05-01 16:59:38.325171
2	1096262005	17189dcf-8ef	2026-05-01 17:31:08.65818
3	1096262005	c15f2c24-c02	2026-05-01 18:21:35.859937
4	1750011039	18dc8e5b-9e1	2026-05-01 19:13:15.28555
5	1096262005	c21029d4-cae	2026-05-06 12:02:25.373248
6	1096262005	e3ec7750-47b	2026-05-06 12:02:25.786078
\.


--
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.files (id, user_id, file_name, file_size, file_type, telegram_file_id, unique_link, password, download_limit, download_count, expires_at, created_at, batch_id) FROM stdin;
338fef21-d886-4d29-ae46-21ed99f2cfa6	1096262005	1.sh	1265	\N	BQACAgUAAxkBAAMGafSKN3L93frxjct7V1mZYvar8AUAAkAiAAJ9maFXKwiFnO6j9Fo7BA	6086e5aa-9ca	asdf	-1	0	2026-05-08 16:40:55.791674+05:30	2026-05-01 16:40:55.791678+05:30	\N
b1e28b38-df73-4914-b2d1-ef0110322c61	1750011039	Inception_part_aa	1048576000	document	BQACAgUAAxkBAAOCafSt5-PZe1uqsD4FJ1khgqDn_6QAAnYcAAJZpqlXhYePR75jfxo7BA	b8af8c50-d6f	\N	-1	1	\N	2026-05-01 19:13:15.291156+05:30	4
9439d40b-fd99-4dc7-8943-6aec78345c18	1750011039	Inception_part_ab	1048576000	document	BQACAgUAAxkBAAODafSt52uIWZXWBMtXnMtdFloSqB8AAnkcAAJZpqlX1IvvPVOjEc07BA	43609055-702	\N	-1	1	\N	2026-05-01 19:13:15.296255+05:30	4
be348d4d-72d8-43a2-94b6-702b8b200f6b	1750011039	Inception_part_ac	1048576000	document	BQACAgUAAxkBAAOEafSt5zxdGscU1WL6JJ2GqtS0r0AAAnscAAJZpqlXZeN-D_oimBI7BA	3f7a6524-e51	\N	-1	1	\N	2026-05-01 19:13:15.301118+05:30	4
37f3fd51-2408-4faa-a241-ff63ff1203ca	1750011039	Inception_part_ad	523256691	document	BQACAgUAAxkBAAOFafSt51YJzvl-QbXjKXWTavYCPHIAAnwcAAJZpqlXmDHx8PRWLiY7BA	a1c1884b-c9e	\N	-1	1	\N	2026-05-01 19:13:15.305914+05:30	4
1c1d6c26-9c22-4617-ac83-8794127b1f58	1750011039	Archive.zip	45806947	document	BQACAgUAAxkBAAOpafpJpGqKfPqlgf8dg3rvzK2UV0oAAk4cAALMqdBXe-AcJNyciYY7BA	e99f0e62-9b5	\N	-1	0	\N	2026-05-06 01:19:00.13481+05:30	\N
c17fd8bc-f442-47bb-8a8d-c2d9e304109b	1750011039	photo_1778010687.jpg	24215	image	AgACAgUAAxkBAAOvafpKP1KJNf-hz7fvnqeIucR5ZwMAAvQPaxvMqdBX6KXd04vj4T8BAAMCAAN4AAM7BA	278d2bd7-551	\N	-1	0	\N	2026-05-06 01:21:54.894832+05:30	\N
\.


--
-- Data for Name: pending_payments; Type: TABLE DATA; Schema: public; Owner: siddu
--

COPY public.pending_payments (id, user_id, plan_name, plan_id, amount, screenshot_file_id, status, created_at, first_name) FROM stdin;
1	1096262005	Unlimited	4	999.00	AgACAgUAAxkBAAICz2n67yi0AAF-YKlWQy5u7T5bkSeyZQACGRBrGxwo2Vd-1Okaf5wlaAEAAwIAA3gAAzsE	pending	2026-05-06 13:05:04.272527	Mr
2	1096262005	Basic	2	99.00	AgACAgUAAxkBAAIC0Wn68bz_ssg3hyR8MMgfRBWsQ5g4AAIfEGsbHCjZV0xYWMc4xqXsAQADAgADeAADOwQ	pending	2026-05-06 13:16:04.399638	Mr
3	1096262005	Basic	2	99.00	AgACAgUAAxkBAAIC0Wn68bz_ssg3hyR8MMgfRBWsQ5g4AAIfEGsbHCjZV0xYWMc4xqXsAQADAgADeAADOwQ	pending	2026-05-06 13:16:06.891205	Mr
4	1096262005	Unlimited	4	999.00	AgACAgUAAxkBAAIC1Gn69SVd8vM2AR2fLJ9Yqmf_0JLkAAIqEGsbHCjZVx_yZCNDqVEwAQADAgADeAADOwQ	pending	2026-05-06 13:30:37.403902	Mr
6	1096262005	Plus	5	50.00	AgACAgUAAxkBAAICwWn64SnqR2YPiQkl_nLM6AGs0ScKAAL4D2sbHCjZV-aRMoau6o9oAQADAgADeAADOwQ	pending	2026-05-06 13:33:26.317356	Mr
7	1096262005	Plus	5	50.00	AgACAgUAAxkBAAICwWn64SnqR2YPiQkl_nLM6AGs0ScKAAL4D2sbHCjZV-aRMoau6o9oAQADAgADeAADOwQ	pending	2026-05-06 13:33:29.210457	Mr
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, name, storage_limit_gb, storage_used_bytes, is_banned, referral_code, referred_by, created_at, last_active) FROM stdin;
1750011039	luxesiblings	LuXe😍	69584231	3714815853	f	16ea1178	\N	2026-05-01 17:39:20.955608+05:30	2026-05-06 01:21:54.903761+05:30
1096262005	Unknown00219	Mr	15664	27889452	f	b383516e	\N	2026-05-01 16:37:29.967941+05:30	2026-05-06 12:02:26.35511+05:30
\.


--
-- Name: batches_id_seq; Type: SEQUENCE SET; Schema: public; Owner: siddu
--

SELECT pg_catalog.setval('public.batches_id_seq', 6, true);


--
-- Name: pending_payments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: siddu
--

SELECT pg_catalog.setval('public.pending_payments_id_seq', 7, true);


--
-- Name: batches batches_pkey; Type: CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.batches
    ADD CONSTRAINT batches_pkey PRIMARY KEY (id);


--
-- Name: batches batches_unique_link_id_key; Type: CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.batches
    ADD CONSTRAINT batches_unique_link_id_key UNIQUE (unique_link_id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: files files_unique_link_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_unique_link_key UNIQUE (unique_link);


--
-- Name: pending_payments pending_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments
    ADD CONSTRAINT pending_payments_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_referral_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);


--
-- Name: files files_batch_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_batch_id_fkey FOREIGN KEY (batch_id) REFERENCES public.batches(id);


--
-- Name: files files_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: pending_payments pending_payments_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments
    ADD CONSTRAINT pending_payments_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.subscriptions(id);


--
-- Name: pending_payments pending_payments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments
    ADD CONSTRAINT pending_payments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: TABLE files; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.files TO siddu;


--
-- Name: TABLE users; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.users TO siddu;


--
-- PostgreSQL database dump complete
--

\unrestrict IyUhQiPMS2VsHrHRzbuDndl8w3sUimrrmDDpqYHBDceHXOwfDgPrO8LaIUp7s32

