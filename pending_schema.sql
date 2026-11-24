--
-- PostgreSQL database dump
--

\restrict ZeyVkCWM77BzrVFqlxTCa6tas7Pc6TVhvFbPHtSOvi1CLcmMF2QKKbaNWbKVVPp

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
-- Name: pending_payments id; Type: DEFAULT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments ALTER COLUMN id SET DEFAULT nextval('public.pending_payments_id_seq'::regclass);


--
-- Name: pending_payments pending_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: siddu
--

ALTER TABLE ONLY public.pending_payments
    ADD CONSTRAINT pending_payments_pkey PRIMARY KEY (id);


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
-- PostgreSQL database dump complete
--

\unrestrict ZeyVkCWM77BzrVFqlxTCa6tas7Pc6TVhvFbPHtSOvi1CLcmMF2QKKbaNWbKVVPp

