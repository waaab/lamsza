-- backend/schema/002_reference.sql - reference rows the app needs to work.
--
-- GENERATED FILE - do not edit by hand.
-- Regenerate with scripts/db-dump-schema.sh and commit the diff.
-- Allowlisted tables only; see scripts/db-dump-schema.sh for why.

--
-- PostgreSQL database dump
--

\restrict eUdlqdWp4m3ZZgvzpaXiaVi1pVe2bsINoyGNbuzfzNEZDHsqEQoE5YVebpOVuRP

-- Dumped from database version 16.15
-- Dumped by pg_dump version 16.15

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: geo_locations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.geo_locations (id, latitude, longitude, address, elevation) VALUES
	(1, 46.1265, 25.8876, 'Szent Anna-tó, Harghita', NULL),
	(3, 46.604473, 25.08611, NULL, NULL),
	(4, 46.78894, 25.786796, NULL, NULL);


--
-- Data for Name: counties; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.counties (id, name, name_ro, name_de, slug, location_id, content) VALUES
	(1000, 'Hargita', 'Harghita', NULL, 'hargita', NULL, NULL),
	(1002, 'Maros', 'Mureș', NULL, 'maros', NULL, NULL),
	(1001, 'Kovászna', 'Covasna', '', 'kovaszna', NULL, '');


--
-- Data for Name: historical_seats; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.historical_seats (id, name, name_ro, name_de, slug, content) VALUES
	(1, 'Csíkszék', 'Ținutul Ciuc', '', 'csikszek', '## Csíkszék

A **Csíkszék** (románul *Ținutul Ciuc*) a székely székek egyike, történelmileg a **Csíki-medence** és környéke. Központja a hagyományos székely közigazgatásban **Csíkszereda** (Miercurea Ciuc) környéke volt.

Ma jórészt **Hargita megye** területére esik. A Csíkszékhez kötődő fiúszékek a néphagyományban a **Gyergyó-** és **Kászonszék**.

### Látnivalók, hagyomány

- Csíki-medence települései, templomok, népviselet, búcsúk
- Kapcsolódó megye: [Hargita megye](/hargita-megye)'),
	(2, 'Udvarhelyszék', 'Ținutul Odorhei', '', 'udvarhelyszek', '## Udvarhelyszék

Az **Udvarhelyszék** (románul *Ținutul Odorhei*) a székely székek egyike; központja **Székelyudvarhely** (Odorheiu Secuiesc) környéke.

A történelmi székhez tartozó terület jelentős része ma is **Hargita megyéhez** tartozik.

### Jellegzetességek

- Székelyudvarhely város és környék települései
- Kapcsolódó megye: [Hargita megye](/hargita-megye)'),
	(3, 'Háromszék', 'Trei Scaune', 'Drei Stühle', 'haromszek', '## Háromszék

A **Háromszék** (románul *Trei Scaune*, németül *Drei Stühle*) a legnagyobb kiterjedésű székely szék volt; központja **Sepsiszentgyörgy** (Sfântu Gheorghe) környéke.

Ma jórészt **Kovászna megye** területére esik (Sepsi-, Kézdi-, Orbai- és Miklósvárszék fiúszékekkel).

### Megye

- [Kovászna megye](/kovaszna-megye)'),
	(4, 'Marosszék', 'Ținutul Mureș', '', 'marosszek', '## Marosszék

A **Marosszék** (románul *Ținutul Mureș*) központja **Marosvásárhely** (Târgu Mureș) környéke; a Maros völgye és a Mezőség kapcsolódó részei tartoztak ide.

A történelmi terület nagy része ma **Maros megyéhez** tartozik.

### Megye

- [Maros megye](/maros-megye)'),
	(5, 'Aranyosszék', 'Ținutul Arieș', NULL, 'aranyosszek', '## Aranyosszék

Az **Aranyosszék** (románul *Ținutul Arieș*) a székely székek közül az **Aranyos völgyéhez** kötődő exklávé volt: a történelmi Magyarország nyugati részén, a **Fehér (Alba) megye** területén.

Ma Románia **Kolozs** és **Fehér** megyéinek határán felel meg; a székely hagyományok szempontjából a székelyföldi székekkel együtt szokás tárgyalni.

### Megjegyzés

Ez a szék nem esik egybe a mai három székelyföldi megyével (Hargita, Kovászna, Maros), de a Lámsza a teljes székely szék-hagyományt szeretné bemutatni.');


--
-- Data for Name: county_historical_seats; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.county_historical_seats (county_id, historical_seat_id) VALUES
	(1000, 1),
	(1000, 2),
	(1001, 3),
	(1002, 4);


--
-- Data for Name: settlement_location_types; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.settlement_location_types (id, slug, label_hu, sort_order) VALUES
	(1, 'municípium', 'Municípium', 0),
	(2, 'város', 'Város', 1),
	(3, 'község', 'Község', 2),
	(4, 'falu', 'Falu', 3),
	(5, 'megye', 'Megye', 4),
	(71, 'kozsegkozpont', 'Községközpont', 0);


--
-- Data for Name: settlements; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.settlements (id, county_id, name, name_ro, name_de, slug, type, location_id, parent_id, post_code, population, area, crest, is_county_seat, content) VALUES
	(1, 1000, 'Csíkszereda', 'Miercurea Ciuc', 'Szeklerburg', 'csikszereda', 'municípium', NULL, NULL, '53101', '1123', '12.5', 'https://upload.wikimedia.org/wikipedia/commons/f/f7/ROU_HR_Miercurea_Ciuc_CoA.PNG', true, NULL),
	(11, 1001, 'Kézdivásárhely', 'Târgu Secuiesc', 'Seklerneumarkt', 'kezdivasarhely', 'municípium', NULL, NULL, '525400', '2353', '37.5', 'https://upload.wikimedia.org/wikipedia/commons/2/26/ROU_CV_Targu_Secuiesc_CoA.jpg', false, NULL),
	(10, 1001, 'Kovászna', 'Covasna', 'Kovasna', 'kovaszna', 'város', NULL, NULL, '53110', '2230', '35.0', '', false, NULL),
	(453, 1001, 'Vargyas', 'Vârghiș', '', 'vargyas', 'község', NULL, NULL, '527180', '1472', '70.24', 'https://upload.wikimedia.org/wikipedia/commons/8/8f/ROU_CV_Varghis_CoA1.png', false, NULL),
	(448, 1001, 'Csernáton', 'Cernat', '', 'csernaton', 'község', NULL, NULL, '527065', '3936', '129', 'https://upload.wikimedia.org/wikipedia/commons/a/a3/ROU_CV_Cernat_CoA.PNG', false, NULL),
	(447, 1001, 'Ikafalva', 'Icafalău', '', 'ikafalva', 'falu', NULL, 448, '53547', '55981', '1127.5', '', false, NULL),
	(12, 1001, 'Torja', '', '', 'torja', 'község', NULL, NULL, '53112', '2476', '40', '', false, NULL),
	(439, 1001, 'Barót', 'Baraolt', '', 'barot', 'város', NULL, NULL, '555555', '8650', '102.3', '', false, NULL),
	(3, 1001, 'Sepsiszentgyörgy', 'Sfântu Gheorghe', 'Sankt Georgen', 'sepsiszentgyorgy', 'municípium', NULL, NULL, '520003 - 520150', '50080', '17.5', 'https://upload.wikimedia.org/wikipedia/commons/8/8a/ROU_CV_Sfantu_Gheorghe_CoA.svg', true, NULL),
	(442, 1002, 'Szováta', 'Sovata', '', 'szovata', 'város', NULL, NULL, NULL, NULL, NULL, NULL, false, NULL),
	(443, 1002, 'Nyárádszereda', 'Miercurea Nirajului', '', 'nyaradszereda', 'város', NULL, NULL, NULL, NULL, NULL, NULL, false, NULL),
	(2, 1000, 'Székelyudvarhely', 'Odorheiu Secuiesc', 'Oderhellen', 'szekelyudvarhely', 'municípium', NULL, NULL, '53102', '1246', '15.0', 'https://upload.wikimedia.org/wikipedia/commons/c/c7/Coa_Romania_Town_Sz%C3%A9kelyudvarhely.svg', false, NULL),
	(452, 1000, 'Csíksomlyó', 'Șumuleu', 'Schomlenberg', 'csiksomlyo', 'község', NULL, 1, '', '', '', '', false, NULL),
	(426, 1000, 'Csíkszentimre', 'Sântimbru', '', 'csikszentimre', 'falu', NULL, NULL, '53526', '53398', '1075.0', '', false, NULL),
	(433, 1000, 'Gyergyószentmiklós', 'Gheorgheni', 'Niklasmarkt', 'gyergyoszentmiklos', 'város', NULL, NULL, NULL, NULL, NULL, NULL, false, NULL),
	(449, 1002, 'Mezőmadaras', 'Mădăraș', '', 'mezomadaras', 'község', NULL, NULL, '547071', '1475', '23.27', '', false, NULL),
	(434, 1000, 'Tusnádfürdő', 'Băile Tușnad', 'Bad Tuschnad', 'tusnadfurdo', 'város', NULL, NULL, NULL, NULL, NULL, NULL, false, NULL),
	(441, 1002, 'Marosvásárhely', 'Târgu Mureș', 'Neumarkt am Mieresch', 'marosvasarhely', 'város', NULL, NULL, NULL, NULL, NULL, NULL, true, NULL),
	(457, 1000, 'Gyergyóalfalu', 'Joseni', '', 'gyergyoalfalu', 'község', NULL, NULL, '537130', '5268', '224,01', 'https://upload.wikimedia.org/wikipedia/commons/b/b3/ROU_HR_Joseni_CoA.png', false, NULL),
	(458, 1000, 'Borzont', '', '', 'borzont', 'falu', NULL, 457, '537131', '710', '', '', false, NULL),
	(454, 1001, 'Kézdialbis', 'Albiș', '', 'kezdialbis', 'falu', NULL, 448, '527071', '371', '', '', false, NULL),
	(14, 1001, 'Gelence', '', '', 'gelence', 'község', NULL, NULL, '53114', '2722', '45.0', '', false, NULL),
	(455, 1001, 'Nyújtód', 'Lunga', '', 'nyujtod', 'falu', NULL, 11, '525401', '1460', '', '', false, NULL),
	(456, 1001, 'Kézdioroszfalu', 'Ruseni', '', 'kezdioroszfalu', 'falu', NULL, 11, '525400', '428', '', '', false, NULL),
	(440, 1001, 'Sepsiszentiván', 'Sântionlunca', '', 'sepsiszentivan', 'falu', NULL, NULL, NULL, NULL, NULL, NULL, false, NULL),
	(451, 1001, 'Gidófalva', 'Ghidfalău', '', 'gidofalva', 'község', NULL, NULL, '527095', '1208', '42', 'https://www.szekeres.ro/wp-content/uploads/2010/10/Gidofalva1.jpg', false, NULL),
	(450, 1001, 'Zabola', 'Zăbala', '', 'zabola', 'község', NULL, NULL, '527190', '4332', '125.79', 'https://www.szekeres.ro/wp-content/uploads/2010/10/Zabola2.jpg', false, NULL);


--
-- Data for Name: venue_types; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.venue_types (id, slug, label_hu, sort_order) VALUES
	(5, 'park', 'Park', 50),
	(10, 'uszoda', 'Uszoda', 9),
	(13, 'egyeb', 'Egyéb', 0),
	(14, 'fedett-csarnok', 'Fedett csarnok', 0),
	(15, 'ideiglenes', 'Ideiglenes', 0),
	(16, 'piac-ter', 'Piac / tér', 0),
	(17, 'sportcsarnok', 'Sportcsarnok', 0),
	(19, 'szabadteri-terulet', 'Szabadtéri terület', 0),
	(20, 'tobb-helyszin', 'Több helyszín', 0),
	(22, 'sportpalya', 'Sportpálya', 0),
	(23, 'mujegpalya', 'Műjégpálya', 0);


--
-- Data for Name: weather_desc_translations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.weather_desc_translations (id, source_text, lang, translated_text) VALUES
	(2, 'részben felhős', 'hu', 'Vannak felhők es...'),
	(5, 'borult', 'hu', 'Béborult esment...'),
	(1, 'helyenként eső a közelben', 'hu', 'Itt-ott eseget...');


--
-- Name: counties_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.counties_id_seq', 1003, true);


--
-- Name: geo_locations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.geo_locations_id_seq', 4, true);


--
-- Name: historical_seats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.historical_seats_id_seq', 984, true);


--
-- Name: settlement_location_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.settlement_location_types_id_seq', 1540, true);


--
-- Name: settlements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.settlements_id_seq', 461, true);


--
-- Name: venue_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.venue_types_id_seq', 23, true);


--
-- Name: weather_desc_translations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.weather_desc_translations_id_seq', 5, true);


--
-- PostgreSQL database dump complete
--

\unrestrict eUdlqdWp4m3ZZgvzpaXiaVi1pVe2bsINoyGNbuzfzNEZDHsqEQoE5YVebpOVuRP

