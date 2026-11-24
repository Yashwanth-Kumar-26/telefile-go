--
-- PostgreSQL database dump
--

\restrict 8nqyLpcDppmRk64txMJgKpG5fhxgf4bfYWOQvdoyTfRCDwPaqMpk2PHaaIdnoAE

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
-- Name: gift_claims; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.gift_claims (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    code text NOT NULL,
    claimed_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.gift_claims OWNER TO postgres;

--
-- Name: gift_claims_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.gift_claims_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.gift_claims_id_seq OWNER TO postgres;

--
-- Name: gift_claims_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.gift_claims_id_seq OWNED BY public.gift_claims.id;


--
-- Name: gift_codes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.gift_codes (
    code text NOT NULL,
    storage_gb integer NOT NULL,
    expires_at timestamp with time zone,
    max_claims integer DEFAULT 1,
    claim_count integer DEFAULT 0,
    created_by bigint,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.gift_codes OWNER TO postgres;

--
-- Name: gifts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.gifts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.gifts_id_seq OWNER TO postgres;

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
-- Name: referrals; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.referrals (
    id integer NOT NULL,
    referrer_id bigint NOT NULL,
    referred_id bigint NOT NULL,
    bonus_gb integer DEFAULT 5,
    is_redeemed boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.referrals OWNER TO postgres;

--
-- Name: referrals_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.referrals_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.referrals_id_seq OWNER TO postgres;

--
-- Name: referrals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.referrals_id_seq OWNED BY public.referrals.id;


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscriptions (
    id integer NOT NULL,
    name text NOT NULL,
    storage_gb integer NOT NULL,
    price_inr numeric(10,2) DEFAULT 0,
    features jsonb DEFAULT '{}'::jsonb,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.subscriptions OWNER TO postgres;

--
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.subscriptions_id_seq OWNER TO postgres;

--
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.subscriptions_id_seq OWNED BY public.subscriptions.id;


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
-- Name: gift_claims id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.gift_claims ALTER COLUMN id SET DEFAULT nextval('public.gift_claims_id_seq'::regclass);


--
-- Name: pending_payments id; Type: DEFAULT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments ALTER COLUMN id SET DEFAULT nextval('public.pending_payments_id_seq'::regclass);


--
-- Name: referrals id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals ALTER COLUMN id SET DEFAULT nextval('public.referrals_id_seq'::regclass);


--
-- Name: subscriptions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions ALTER COLUMN id SET DEFAULT nextval('public.subscriptions_id_seq'::regclass);


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
-- Name: gift_claims gift_claims_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.gift_claims
    ADD CONSTRAINT gift_claims_pkey PRIMARY KEY (id);


--
-- Name: gift_claims gift_claims_user_id_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.gift_claims
    ADD CONSTRAINT gift_claims_user_id_code_key UNIQUE (user_id, code);


--
-- Name: gift_codes gift_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.gift_codes
    ADD CONSTRAINT gift_codes_pkey PRIMARY KEY (code);


--
-- Name: pending_payments pending_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments
    ADD CONSTRAINT pending_payments_pkey PRIMARY KEY (id);


--
-- Name: referrals referrals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT referrals_pkey PRIMARY KEY (id);


--
-- Name: referrals referrals_referred_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT referrals_referred_id_key UNIQUE (referred_id);


--
-- Name: subscriptions subscriptions_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_name_key UNIQUE (name);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


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
-- Name: idx_referrals_referred; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_referrals_referred ON public.referrals USING btree (referred_id);


--
-- Name: idx_referrals_referrer; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_referrals_referrer ON public.referrals USING btree (referrer_id);


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
-- Name: gift_claims gift_claims_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.gift_claims
    ADD CONSTRAINT gift_claims_code_fkey FOREIGN KEY (code) REFERENCES public.gift_codes(code);


--
-- Name: gift_claims gift_claims_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.gift_claims
    ADD CONSTRAINT gift_claims_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


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
-- Name: referrals referrals_referred_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT referrals_referred_id_fkey FOREIGN KEY (referred_id) REFERENCES public.users(id);


--
-- Name: referrals referrals_referrer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT referrals_referrer_id_fkey FOREIGN KEY (referrer_id) REFERENCES public.users(id);


--
-- Name: TABLE files; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.files TO siddu;


--
-- Name: TABLE gift_claims; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.gift_claims TO siddu;


--
-- Name: SEQUENCE gift_claims_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.gift_claims_id_seq TO siddu;


--
-- Name: TABLE gift_codes; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.gift_codes TO siddu;


--
-- Name: TABLE referrals; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.referrals TO siddu;


--
-- Name: SEQUENCE referrals_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.referrals_id_seq TO siddu;


--
-- Name: TABLE subscriptions; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.subscriptions TO siddu;


--
-- Name: SEQUENCE subscriptions_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.subscriptions_id_seq TO siddu;


--
-- Name: TABLE users; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.users TO siddu;


--
-- PostgreSQL database dump complete
--

\unrestrict 8nqyLpcDppmRk64txMJgKpG5fhxgf4bfYWOQvdoyTfRCDwPaqMpk2PHaaIdnoAE

