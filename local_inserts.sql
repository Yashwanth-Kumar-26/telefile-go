--
-- PostgreSQL database dump
--

\restrict NtiAnkWkIxeiMHMIa6mkKWIy5Erw0k81MnmEgwZHgxYRHYkKpVVNDihHWTu8or4

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

--
-- Data for Name: batches; Type: TABLE DATA; Schema: public; Owner: siddu
--

INSERT INTO public.batches VALUES (1, 1096262005, 'a1b5be12-498', '2026-05-01 16:59:38.325171');
INSERT INTO public.batches VALUES (2, 1096262005, '17189dcf-8ef', '2026-05-01 17:31:08.65818');
INSERT INTO public.batches VALUES (3, 1096262005, 'c15f2c24-c02', '2026-05-01 18:21:35.859937');
INSERT INTO public.batches VALUES (4, 1750011039, '18dc8e5b-9e1', '2026-05-01 19:13:15.28555');
INSERT INTO public.batches VALUES (5, 1096262005, 'c21029d4-cae', '2026-05-06 12:02:25.373248');
INSERT INTO public.batches VALUES (6, 1096262005, 'e3ec7750-47b', '2026-05-06 12:02:25.786078');


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.users VALUES (1750011039, 'luxesiblings', 'LuXe😍', 69584231, 3714815853, false, '16ea1178', NULL, '2026-05-01 17:39:20.955608+05:30', '2026-05-06 01:21:54.903761+05:30');
INSERT INTO public.users VALUES (1096262005, 'Unknown00219', 'Mr', 15664, 27889452, false, 'b383516e', NULL, '2026-05-01 16:37:29.967941+05:30', '2026-05-06 12:02:26.35511+05:30');


--
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.files VALUES ('338fef21-d886-4d29-ae46-21ed99f2cfa6', 1096262005, '1.sh', 1265, NULL, 'BQACAgUAAxkBAAMGafSKN3L93frxjct7V1mZYvar8AUAAkAiAAJ9maFXKwiFnO6j9Fo7BA', '6086e5aa-9ca', 'asdf', -1, 0, '2026-05-08 16:40:55.791674+05:30', '2026-05-01 16:40:55.791678+05:30', NULL);
INSERT INTO public.files VALUES ('b1e28b38-df73-4914-b2d1-ef0110322c61', 1750011039, 'Inception_part_aa', 1048576000, 'document', 'BQACAgUAAxkBAAOCafSt5-PZe1uqsD4FJ1khgqDn_6QAAnYcAAJZpqlXhYePR75jfxo7BA', 'b8af8c50-d6f', NULL, -1, 1, NULL, '2026-05-01 19:13:15.291156+05:30', 4);
INSERT INTO public.files VALUES ('9439d40b-fd99-4dc7-8943-6aec78345c18', 1750011039, 'Inception_part_ab', 1048576000, 'document', 'BQACAgUAAxkBAAODafSt52uIWZXWBMtXnMtdFloSqB8AAnkcAAJZpqlX1IvvPVOjEc07BA', '43609055-702', NULL, -1, 1, NULL, '2026-05-01 19:13:15.296255+05:30', 4);
INSERT INTO public.files VALUES ('be348d4d-72d8-43a2-94b6-702b8b200f6b', 1750011039, 'Inception_part_ac', 1048576000, 'document', 'BQACAgUAAxkBAAOEafSt5zxdGscU1WL6JJ2GqtS0r0AAAnscAAJZpqlXZeN-D_oimBI7BA', '3f7a6524-e51', NULL, -1, 1, NULL, '2026-05-01 19:13:15.301118+05:30', 4);
INSERT INTO public.files VALUES ('37f3fd51-2408-4faa-a241-ff63ff1203ca', 1750011039, 'Inception_part_ad', 523256691, 'document', 'BQACAgUAAxkBAAOFafSt51YJzvl-QbXjKXWTavYCPHIAAnwcAAJZpqlXmDHx8PRWLiY7BA', 'a1c1884b-c9e', NULL, -1, 1, NULL, '2026-05-01 19:13:15.305914+05:30', 4);
INSERT INTO public.files VALUES ('1c1d6c26-9c22-4617-ac83-8794127b1f58', 1750011039, 'Archive.zip', 45806947, 'document', 'BQACAgUAAxkBAAOpafpJpGqKfPqlgf8dg3rvzK2UV0oAAk4cAALMqdBXe-AcJNyciYY7BA', 'e99f0e62-9b5', NULL, -1, 0, NULL, '2026-05-06 01:19:00.13481+05:30', NULL);
INSERT INTO public.files VALUES ('c17fd8bc-f442-47bb-8a8d-c2d9e304109b', 1750011039, 'photo_1778010687.jpg', 24215, 'image', 'AgACAgUAAxkBAAOvafpKP1KJNf-hz7fvnqeIucR5ZwMAAvQPaxvMqdBX6KXd04vj4T8BAAMCAAN4AAM7BA', '278d2bd7-551', NULL, -1, 0, NULL, '2026-05-06 01:21:54.894832+05:30', NULL);


--
-- Data for Name: pending_payments; Type: TABLE DATA; Schema: public; Owner: siddu
--

INSERT INTO public.pending_payments VALUES (1, 1096262005, 'Unlimited', 4, 999.00, 'AgACAgUAAxkBAAICz2n67yi0AAF-YKlWQy5u7T5bkSeyZQACGRBrGxwo2Vd-1Okaf5wlaAEAAwIAA3gAAzsE', 'pending', '2026-05-06 13:05:04.272527', 'Mr');
INSERT INTO public.pending_payments VALUES (2, 1096262005, 'Basic', 2, 99.00, 'AgACAgUAAxkBAAIC0Wn68bz_ssg3hyR8MMgfRBWsQ5g4AAIfEGsbHCjZV0xYWMc4xqXsAQADAgADeAADOwQ', 'pending', '2026-05-06 13:16:04.399638', 'Mr');
INSERT INTO public.pending_payments VALUES (3, 1096262005, 'Basic', 2, 99.00, 'AgACAgUAAxkBAAIC0Wn68bz_ssg3hyR8MMgfRBWsQ5g4AAIfEGsbHCjZV0xYWMc4xqXsAQADAgADeAADOwQ', 'pending', '2026-05-06 13:16:06.891205', 'Mr');
INSERT INTO public.pending_payments VALUES (4, 1096262005, 'Unlimited', 4, 999.00, 'AgACAgUAAxkBAAIC1Gn69SVd8vM2AR2fLJ9Yqmf_0JLkAAIqEGsbHCjZVx_yZCNDqVEwAQADAgADeAADOwQ', 'pending', '2026-05-06 13:30:37.403902', 'Mr');
INSERT INTO public.pending_payments VALUES (6, 1096262005, 'Plus', 5, 50.00, 'AgACAgUAAxkBAAICwWn64SnqR2YPiQkl_nLM6AGs0ScKAAL4D2sbHCjZV-aRMoau6o9oAQADAgADeAADOwQ', 'pending', '2026-05-06 13:33:26.317356', 'Mr');
INSERT INTO public.pending_payments VALUES (7, 1096262005, 'Plus', 5, 50.00, 'AgACAgUAAxkBAAICwWn64SnqR2YPiQkl_nLM6AGs0ScKAAL4D2sbHCjZV-aRMoau6o9oAQADAgADeAADOwQ', 'pending', '2026-05-06 13:33:29.210457', 'Mr');


--
-- Name: batches_id_seq; Type: SEQUENCE SET; Schema: public; Owner: siddu
--

SELECT pg_catalog.setval('public.batches_id_seq', 6, true);


--
-- Name: pending_payments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: siddu
--

SELECT pg_catalog.setval('public.pending_payments_id_seq', 7, true);


--
-- PostgreSQL database dump complete
--

\unrestrict NtiAnkWkIxeiMHMIa6mkKWIy5Erw0k81MnmEgwZHgxYRHYkKpVVNDihHWTu8or4

