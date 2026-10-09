-- backend/schema/001_schema.sql - the whole lamsza schema, structure only.
--
-- GENERATED FILE - do not edit by hand.
-- Regenerate with scripts/db-dump-schema.sh and commit the diff.
-- Apply it with scripts/db-bootstrap.sh, which is what CI uses.

--
-- PostgreSQL database dump
--

\restrict h6ZSDETeibZDkc1IwBnnkv859xf4EVXBnx66sTNYaXUQEXczoTW7eeY8KUMQdpC

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
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: unaccent; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;


--
-- Name: pg_slugify(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.pg_slugify(text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
DECLARE
    s text;
BEGIN
    s := lower($1);
    -- Simple Hungarian replacement
    s := translate(s, 'áéíóöőúüű', 'aeiooouuu');
    -- Replace non-alphanumeric with dash
    s := regexp_replace(s, '[^a-z0-9]+', '-', 'g');
    -- Trim dashes
    s := trim(both '-' from s);
    RETURN s;
END;
$_$;


--
-- Name: sync_entry_category(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_entry_category() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
		 BEGIN
		   UPDATE entries e
		   SET cat_name = (
		     SELECT string_agg(c.name, ' ' ORDER BY l.is_primary DESC, c.sort_order, c.name)
		     FROM entry_category_links l
		     JOIN entry_categories c ON c.id = l.category_id
		     WHERE l.entry_id = e.id
		   )
		   WHERE e.id IN (
		     SELECT entry_id FROM entry_category_links WHERE category_id = NEW.id
		   );
		   RETURN NEW;
		 END;
		 $$;


--
-- Name: sync_entry_location(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_entry_location() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE entries
    SET loc_name = NEW.name,
        loc_name_ro = NEW.name_ro,
        loc_name_de = NEW.name_de
    WHERE location_id = NEW.id;
    RETURN NEW;
END;
$$;


--
-- Name: sync_entry_location_names(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_entry_location_names() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.location_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT s.name, COALESCE(s.name_ro, ''), COALESCE(s.name_de, '')
      INTO NEW.loc_name, NEW.loc_name_ro, NEW.loc_name_de
      FROM settlements s
     WHERE s.id = NEW.location_id;

    RETURN NEW;
END;
$$;


--
-- Name: sync_entry_settlement(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_entry_settlement() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE entries SET loc_name = NEW.name, loc_name_ro = COALESCE(NEW.name_ro, ''), loc_name_de = COALESCE(NEW.name_de, '')
    WHERE location_id = NEW.id;
    RETURN NEW;
END;
$$;


--
-- Name: sync_entry_tags(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_entry_tags() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    e_id INT;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        e_id := OLD.entry_id;
    ELSE
        e_id := NEW.entry_id;
    END IF;

    UPDATE entries
    SET tag_names = (
        SELECT string_agg(t.name, ' ')
        FROM tags t
        JOIN entry_tags et ON t.id = et.tag_id
        WHERE et.entry_id = e_id
    )
    WHERE id = e_id;
    
    RETURN NULL;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_audit_log (
    id bigint NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    actor_user_id integer,
    actor_email text DEFAULT ''::text NOT NULL,
    resource text NOT NULL,
    resource_id text DEFAULT ''::text NOT NULL,
    action text NOT NULL,
    method text NOT NULL,
    route text NOT NULL,
    status_code integer DEFAULT 0 NOT NULL,
    payload jsonb,
    before_state jsonb,
    after_state jsonb,
    diff jsonb
);


--
-- Name: admin_audit_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.admin_audit_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: admin_audit_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.admin_audit_log_id_seq OWNED BY public.admin_audit_log.id;


--
-- Name: admin_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_sessions (
    token_hash character(64) NOT NULL,
    user_id integer NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: attraction_contributors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attraction_contributors (
    attraction_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: attraction_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attraction_images (
    id integer NOT NULL,
    attraction_id integer NOT NULL,
    url text NOT NULL,
    sort_order integer DEFAULT 0,
    copyright text DEFAULT ''::text NOT NULL
);


--
-- Name: attraction_images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attraction_images_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attraction_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attraction_images_id_seq OWNED BY public.attraction_images.id;


--
-- Name: attraction_suggestions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attraction_suggestions (
    id integer NOT NULL,
    attraction_id integer NOT NULL,
    user_id integer NOT NULL,
    changes jsonb NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT attraction_suggestions_status_check CHECK ((status = ANY (ARRAY['open'::text, 'accepted'::text, 'denied'::text])))
);


--
-- Name: attraction_suggestions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attraction_suggestions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attraction_suggestions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attraction_suggestions_id_seq OWNED BY public.attraction_suggestions.id;


--
-- Name: attractions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attractions (
    id integer NOT NULL,
    county_id integer NOT NULL,
    name character varying(255) NOT NULL,
    name_ro character varying(255),
    name_de character varying(255),
    slug character varying(120) NOT NULL,
    description text,
    location_id integer,
    featured_image text,
    content text,
    activities text DEFAULT ''::text NOT NULL,
    featured_image_copyright text DEFAULT ''::text NOT NULL,
    prohibitions text DEFAULT ''::text NOT NULL,
    elevation_m numeric(7,1),
    area_km2 numeric(10,3),
    depth_m numeric(7,1),
    CONSTRAINT attractions_area_km2_check CHECK ((area_km2 >= (0)::numeric)),
    CONSTRAINT attractions_depth_m_check CHECK ((depth_m >= (0)::numeric))
);


--
-- Name: attractions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attractions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attractions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attractions_id_seq OWNED BY public.attractions.id;


--
-- Name: catalog_event_subtypes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.catalog_event_subtypes (
    id integer NOT NULL,
    event_type_id integer NOT NULL,
    slug character varying(64) NOT NULL,
    label_hu text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: catalog_event_subtypes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.catalog_event_subtypes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: catalog_event_subtypes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.catalog_event_subtypes_id_seq OWNED BY public.catalog_event_subtypes.id;


--
-- Name: catalog_event_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.catalog_event_types (
    id integer NOT NULL,
    slug character varying(64) NOT NULL,
    label_hu text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: catalog_event_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.catalog_event_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: catalog_event_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.catalog_event_types_id_seq OWNED BY public.catalog_event_types.id;


--
-- Name: counties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.counties (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    name_ro character varying(255),
    name_de character varying(255),
    slug character varying(120) NOT NULL,
    location_id integer,
    content text
);


--
-- Name: counties_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.counties_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: counties_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.counties_id_seq OWNED BY public.counties.id;


--
-- Name: county_historical_seats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.county_historical_seats (
    county_id integer NOT NULL,
    historical_seat_id integer NOT NULL
);


--
-- Name: entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entries (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    phone character varying(50),
    address text,
    notes text,
    url character varying(255),
    category_id integer NOT NULL,
    slug character varying(255),
    languages character varying(10)[] DEFAULT '{HU}'::character varying[],
    loc_name character varying(255),
    loc_name_ro character varying(255),
    loc_name_de character varying(255),
    cat_name character varying(100),
    tag_names text,
    search_vector tsvector GENERATED ALWAYS AS ((((((((setweight(to_tsvector('simple'::regconfig, public.pg_slugify((name)::text)), 'A'::"char") || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify((loc_name)::text), ''::text)), 'B'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify((loc_name_ro)::text), ''::text)), 'B'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify((loc_name_de)::text), ''::text)), 'B'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify((cat_name)::text), ''::text)), 'C'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify(tag_names), ''::text)), 'C'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify(notes), ''::text)), 'D'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(public.pg_slugify(address), ''::text)), 'D'::"char"))) STORED,
    location_id integer,
    type_id integer NOT NULL,
    claimed boolean DEFAULT false NOT NULL,
    hours jsonb DEFAULT '{}'::jsonb NOT NULL,
    delivery_hours jsonb DEFAULT '{}'::jsonb NOT NULL,
    photos jsonb DEFAULT '[]'::jsonb NOT NULL,
    published boolean DEFAULT true NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    ratings_enabled boolean DEFAULT false NOT NULL,
    hours_enabled boolean DEFAULT false NOT NULL,
    delivery_enabled boolean DEFAULT false NOT NULL,
    social_links jsonb DEFAULT '[]'::jsonb NOT NULL
);


--
-- Name: entry_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_categories (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    slug character varying(120),
    parent_id integer,
    sort_order integer DEFAULT 0 NOT NULL,
    featured_order integer,
    CONSTRAINT entry_categories_featured_order_check CHECK (((featured_order >= 1) AND (featured_order <= 6)))
);


--
-- Name: entry_category_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_category_links (
    entry_id integer NOT NULL,
    category_id integer NOT NULL,
    is_primary boolean DEFAULT false NOT NULL
);


--
-- Name: entry_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_members (
    entry_id integer NOT NULL,
    user_id integer NOT NULL,
    role character varying(16) NOT NULL,
    status character varying(16) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: entry_reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_reviews (
    id integer NOT NULL,
    entry_id integer NOT NULL,
    user_id integer NOT NULL,
    score smallint NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT entry_reviews_score_check CHECK (((score >= 1) AND (score <= 5)))
);


--
-- Name: entry_reviews_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.entry_reviews_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: entry_reviews_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.entry_reviews_id_seq OWNED BY public.entry_reviews.id;


--
-- Name: entry_suggestions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_suggestions (
    id integer NOT NULL,
    entry_id integer NOT NULL,
    user_id integer NOT NULL,
    changes jsonb NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT entry_suggestions_status_check CHECK ((status = ANY (ARRAY['open'::text, 'accepted'::text, 'denied'::text])))
);


--
-- Name: entry_suggestions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.entry_suggestions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: entry_suggestions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.entry_suggestions_id_seq OWNED BY public.entry_suggestions.id;


--
-- Name: entry_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_tags (
    entry_id integer NOT NULL,
    tag_id integer NOT NULL
);


--
-- Name: entry_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entry_types (
    id integer NOT NULL,
    name character varying(50) NOT NULL
);


--
-- Name: entry_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.entry_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: entry_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.entry_types_id_seq OWNED BY public.entry_types.id;


--
-- Name: event_schedule_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_schedule_activities (
    id integer NOT NULL,
    event_day_id integer NOT NULL,
    starts_at time without time zone,
    ends_at time without time zone,
    title character varying(500) NOT NULL,
    description text,
    sort_order integer DEFAULT 0 NOT NULL,
    activity_type character varying(40) DEFAULT 'other'::character varying NOT NULL,
    venue_id integer
);


--
-- Name: event_schedule_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_schedule_activities_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_schedule_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_schedule_activities_id_seq OWNED BY public.event_schedule_activities.id;


--
-- Name: event_schedule_days; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_schedule_days (
    id integer NOT NULL,
    event_id integer NOT NULL,
    schedule_date date NOT NULL,
    notes text,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: event_schedule_days_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_schedule_days_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_schedule_days_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_schedule_days_id_seq OWNED BY public.event_schedule_days.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id integer NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    start_date date NOT NULL,
    start_time time without time zone,
    organizer character varying(255),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    end_date date NOT NULL,
    end_time time without time zone,
    location_id integer,
    default_venue_id integer,
    featured_image character varying(1024) DEFAULT ''::character varying,
    entry_price character varying(128) DEFAULT ''::character varying,
    access_type character varying(32) DEFAULT 'public'::character varying,
    event_type_id integer NOT NULL,
    event_subtype_id integer,
    attraction_id integer,
    featured_image_copyright text DEFAULT ''::text NOT NULL,
    CONSTRAINT events_access_type_check CHECK (((access_type)::text = ANY ((ARRAY['public'::character varying, 'members_only'::character varying, 'invitation_only'::character varying])::text[])))
);


--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: geo_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.geo_locations (
    id integer NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    address text,
    elevation integer
);


--
-- Name: geo_locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.geo_locations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: geo_locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.geo_locations_id_seq OWNED BY public.geo_locations.id;


--
-- Name: historical_seats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.historical_seats (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    name_ro character varying(255),
    name_de character varying(255),
    slug character varying(120) NOT NULL,
    content text
);


--
-- Name: historical_seats_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.historical_seats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: historical_seats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.historical_seats_id_seq OWNED BY public.historical_seats.id;


--
-- Name: settlements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.settlements (
    id integer NOT NULL,
    county_id integer NOT NULL,
    name character varying(255) NOT NULL,
    name_ro character varying(255),
    name_de character varying(255),
    slug character varying(120) NOT NULL,
    type character varying(50) NOT NULL,
    location_id integer,
    parent_id integer,
    post_code character varying(20),
    population character varying(50),
    area character varying(50),
    crest text,
    is_county_seat boolean DEFAULT false,
    content text
);


--
-- Name: locations; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.locations AS
 SELECT counties.id,
    counties.name,
    counties.name_ro,
    counties.name_de,
    counties.name AS county,
    counties.slug AS county_slug,
    'megye'::character varying AS type,
    counties.slug,
    NULL::text AS post_code,
    NULL::text AS coordinates,
    NULL::text AS population,
    NULL::text AS area,
    NULL::text AS crest,
    NULL::integer AS parent_id,
    false AS is_county_seat
   FROM public.counties
UNION ALL
 SELECT s.id,
    s.name,
    s.name_ro,
    s.name_de,
    c.name AS county,
    c.slug AS county_slug,
    s.type,
    s.slug,
    s.post_code,
    ( SELECT (((gl.latitude)::text || ', '::text) || (gl.longitude)::text)
           FROM public.geo_locations gl
          WHERE (gl.id = s.location_id)) AS coordinates,
    s.population,
    s.area,
    s.crest,
    s.parent_id,
    s.is_county_seat
   FROM (public.settlements s
     JOIN public.counties c ON ((s.county_id = c.id)));


--
-- Name: news_feeds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.news_feeds (
    id integer NOT NULL,
    title character varying(100) NOT NULL,
    feed_url character varying(255) NOT NULL,
    bg_color character varying(50) DEFAULT '#ffebd6'::character varying,
    county_slug character varying(120)
);


--
-- Name: news_feeds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.news_feeds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: news_feeds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.news_feeds_id_seq OWNED BY public.news_feeds.id;


--
-- Name: page_faq_sections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.page_faq_sections (
    id integer NOT NULL,
    section_key character varying(64) NOT NULL,
    label_hu character varying(255) DEFAULT ''::character varying NOT NULL,
    faq_title character varying(500) DEFAULT ''::character varying NOT NULL,
    disclaimer_markdown text DEFAULT ''::text NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    faq_items jsonb DEFAULT '[]'::jsonb NOT NULL
);


--
-- Name: page_faq_sections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.page_faq_sections_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: page_faq_sections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.page_faq_sections_id_seq OWNED BY public.page_faq_sections.id;


--
-- Name: pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pages (
    id integer NOT NULL,
    slug character varying(255) NOT NULL,
    title character varying(255) DEFAULT ''::character varying NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    greeting text DEFAULT ''::text NOT NULL
);


--
-- Name: pages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pages_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pages_id_seq OWNED BY public.pages.id;


--
-- Name: quick_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.quick_links (
    id integer NOT NULL,
    title character varying(100) NOT NULL,
    url character varying(255) NOT NULL,
    match_category character varying(100),
    bg_color character varying(20) DEFAULT '#ffffff'::character varying
);


--
-- Name: quick_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.quick_links_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: quick_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.quick_links_id_seq OWNED BY public.quick_links.id;


--
-- Name: service_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.service_categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: service_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.service_categories_id_seq OWNED BY public.entry_categories.id;


--
-- Name: services_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.services_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.services_id_seq OWNED BY public.entries.id;


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    token_hash character(64) NOT NULL,
    user_id integer NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: settlement_location_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.settlement_location_types (
    id integer NOT NULL,
    slug character varying(64) NOT NULL,
    label_hu text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: settlement_location_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.settlement_location_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: settlement_location_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.settlement_location_types_id_seq OWNED BY public.settlement_location_types.id;


--
-- Name: settlements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.settlements_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: settlements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.settlements_id_seq OWNED BY public.settlements.id;


--
-- Name: site_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_settings (
    key character varying(100) NOT NULL,
    value text NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tags (
    id integer NOT NULL,
    name character varying(100) NOT NULL
);


--
-- Name: tags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tags_id_seq OWNED BY public.tags.id;


--
-- Name: user_entry_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_entry_history (
    user_id integer NOT NULL,
    slug character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    category character varying(255) DEFAULT ''::character varying NOT NULL,
    location character varying(255) DEFAULT ''::character varying NOT NULL,
    photo text DEFAULT ''::text NOT NULL,
    viewed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: user_favorites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_favorites (
    user_id integer NOT NULL,
    entity_type character varying(20) NOT NULL,
    entity_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: user_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_links (
    id integer NOT NULL,
    user_id integer NOT NULL,
    title character varying(100) NOT NULL,
    url character varying(255) NOT NULL,
    bg_color character varying(50) DEFAULT '#e6f0ff'::character varying NOT NULL,
    "position" integer NOT NULL
);


--
-- Name: user_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_links_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_links_id_seq OWNED BY public.user_links.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    google_sub character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    name character varying(255) DEFAULT ''::character varying NOT NULL,
    last_login_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    given_name character varying(255) DEFAULT ''::character varying NOT NULL,
    family_name character varying(255) DEFAULT ''::character varying NOT NULL,
    picture text DEFAULT ''::text NOT NULL,
    locale character varying(35) DEFAULT ''::character varying NOT NULL,
    theme character varying(16),
    quicklink_slots integer,
    prefs_imported_at timestamp without time zone,
    preferred_settlement_id integer,
    website_banned boolean DEFAULT false NOT NULL,
    display_name character varying(24) DEFAULT ''::character varying NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: venue_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.venue_types (
    id integer NOT NULL,
    slug character varying(64) NOT NULL,
    label_hu character varying(255) NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: venue_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.venue_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: venue_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.venue_types_id_seq OWNED BY public.venue_types.id;


--
-- Name: venues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.venues (
    id integer NOT NULL,
    settlement_id integer NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(200) NOT NULL,
    kind character varying(40) DEFAULT 'other'::character varying NOT NULL,
    address text,
    notes text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    name_ro character varying(255) DEFAULT ''::character varying,
    name_de character varying(255) DEFAULT ''::character varying,
    latitude double precision,
    longitude double precision,
    seating_capacity integer,
    description text
);


--
-- Name: venues_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.venues_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: venues_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.venues_id_seq OWNED BY public.venues.id;


--
-- Name: weather_astro_daily; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.weather_astro_daily (
    place_kind character varying(20) NOT NULL,
    place_id integer NOT NULL,
    day date NOT NULL,
    sunrise timestamp with time zone,
    sunset timestamp with time zone,
    moonrise timestamp with time zone,
    moonset timestamp with time zone,
    moon_phase double precision
);


--
-- Name: weather_daily_archive; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.weather_daily_archive (
    settlement_id integer NOT NULL,
    day date NOT NULL,
    symbol character varying(60) NOT NULL,
    tmin double precision,
    tmax double precision,
    precip_mm double precision DEFAULT 0 NOT NULL,
    wind_max_kph double precision,
    hours integer NOT NULL
);


--
-- Name: weather_desc_translations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.weather_desc_translations (
    id integer NOT NULL,
    source_text character varying(255) NOT NULL,
    lang character varying(10) NOT NULL,
    translated_text character varying(255) NOT NULL
);


--
-- Name: weather_desc_translations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.weather_desc_translations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: weather_desc_translations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.weather_desc_translations_id_seq OWNED BY public.weather_desc_translations.id;


--
-- Name: weather_forecast_cache; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.weather_forecast_cache (
    place_kind character varying(20) NOT NULL,
    place_id integer NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    coord_source character varying(20) NOT NULL,
    source character varying(30),
    fetched_at timestamp with time zone,
    expires_at timestamp with time zone,
    last_modified text,
    payload jsonb
);


--
-- Name: weather_obs_hourly; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.weather_obs_hourly (
    settlement_id integer NOT NULL,
    observed_at timestamp with time zone NOT NULL,
    symbol character varying(60) NOT NULL,
    temp double precision,
    feels double precision,
    precip_mm double precision,
    wind_kph double precision,
    wind_dir double precision,
    humidity double precision,
    pressure_hpa double precision,
    cloud_pct double precision
);


--
-- Name: website_category_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.website_category_links (
    website_id integer NOT NULL,
    category_id integer NOT NULL,
    is_primary boolean DEFAULT false NOT NULL
);


--
-- Name: websites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.websites (
    id integer NOT NULL,
    domain_key character varying(253) NOT NULL,
    submitted_host character varying(253) NOT NULL,
    title character varying(120) DEFAULT ''::character varying NOT NULL,
    description character varying(300) DEFAULT ''::character varying NOT NULL,
    status character varying(16) NOT NULL,
    user_id integer,
    entry_id integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    approved_at timestamp without time zone,
    approved_by integer,
    category_id integer,
    CONSTRAINT websites_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying])::text[])))
);


--
-- Name: websites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.websites_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: websites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.websites_id_seq OWNED BY public.websites.id;


--
-- Name: admin_audit_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_audit_log ALTER COLUMN id SET DEFAULT nextval('public.admin_audit_log_id_seq'::regclass);


--
-- Name: attraction_images id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_images ALTER COLUMN id SET DEFAULT nextval('public.attraction_images_id_seq'::regclass);


--
-- Name: attraction_suggestions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_suggestions ALTER COLUMN id SET DEFAULT nextval('public.attraction_suggestions_id_seq'::regclass);


--
-- Name: attractions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attractions ALTER COLUMN id SET DEFAULT nextval('public.attractions_id_seq'::regclass);


--
-- Name: catalog_event_subtypes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_subtypes ALTER COLUMN id SET DEFAULT nextval('public.catalog_event_subtypes_id_seq'::regclass);


--
-- Name: catalog_event_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_types ALTER COLUMN id SET DEFAULT nextval('public.catalog_event_types_id_seq'::regclass);


--
-- Name: counties id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.counties ALTER COLUMN id SET DEFAULT nextval('public.counties_id_seq'::regclass);


--
-- Name: entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries ALTER COLUMN id SET DEFAULT nextval('public.services_id_seq'::regclass);


--
-- Name: entry_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_categories ALTER COLUMN id SET DEFAULT nextval('public.service_categories_id_seq'::regclass);


--
-- Name: entry_reviews id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_reviews ALTER COLUMN id SET DEFAULT nextval('public.entry_reviews_id_seq'::regclass);


--
-- Name: entry_suggestions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_suggestions ALTER COLUMN id SET DEFAULT nextval('public.entry_suggestions_id_seq'::regclass);


--
-- Name: entry_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_types ALTER COLUMN id SET DEFAULT nextval('public.entry_types_id_seq'::regclass);


--
-- Name: event_schedule_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_activities ALTER COLUMN id SET DEFAULT nextval('public.event_schedule_activities_id_seq'::regclass);


--
-- Name: event_schedule_days id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_days ALTER COLUMN id SET DEFAULT nextval('public.event_schedule_days_id_seq'::regclass);


--
-- Name: events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: geo_locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.geo_locations ALTER COLUMN id SET DEFAULT nextval('public.geo_locations_id_seq'::regclass);


--
-- Name: historical_seats id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historical_seats ALTER COLUMN id SET DEFAULT nextval('public.historical_seats_id_seq'::regclass);


--
-- Name: news_feeds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_feeds ALTER COLUMN id SET DEFAULT nextval('public.news_feeds_id_seq'::regclass);


--
-- Name: page_faq_sections id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_faq_sections ALTER COLUMN id SET DEFAULT nextval('public.page_faq_sections_id_seq'::regclass);


--
-- Name: pages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages ALTER COLUMN id SET DEFAULT nextval('public.pages_id_seq'::regclass);


--
-- Name: quick_links id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quick_links ALTER COLUMN id SET DEFAULT nextval('public.quick_links_id_seq'::regclass);


--
-- Name: settlement_location_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlement_location_types ALTER COLUMN id SET DEFAULT nextval('public.settlement_location_types_id_seq'::regclass);


--
-- Name: settlements id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlements ALTER COLUMN id SET DEFAULT nextval('public.settlements_id_seq'::regclass);


--
-- Name: tags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags ALTER COLUMN id SET DEFAULT nextval('public.tags_id_seq'::regclass);


--
-- Name: user_links id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_links ALTER COLUMN id SET DEFAULT nextval('public.user_links_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: venue_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venue_types ALTER COLUMN id SET DEFAULT nextval('public.venue_types_id_seq'::regclass);


--
-- Name: venues id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venues ALTER COLUMN id SET DEFAULT nextval('public.venues_id_seq'::regclass);


--
-- Name: weather_desc_translations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_desc_translations ALTER COLUMN id SET DEFAULT nextval('public.weather_desc_translations_id_seq'::regclass);


--
-- Name: websites id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites ALTER COLUMN id SET DEFAULT nextval('public.websites_id_seq'::regclass);


--
-- Name: admin_audit_log admin_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_audit_log
    ADD CONSTRAINT admin_audit_log_pkey PRIMARY KEY (id);


--
-- Name: admin_sessions admin_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_pkey PRIMARY KEY (token_hash);


--
-- Name: attraction_contributors attraction_contributors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_contributors
    ADD CONSTRAINT attraction_contributors_pkey PRIMARY KEY (attraction_id, user_id);


--
-- Name: attraction_images attraction_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_images
    ADD CONSTRAINT attraction_images_pkey PRIMARY KEY (id);


--
-- Name: attraction_suggestions attraction_suggestions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_suggestions
    ADD CONSTRAINT attraction_suggestions_pkey PRIMARY KEY (id);


--
-- Name: attractions attractions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attractions
    ADD CONSTRAINT attractions_pkey PRIMARY KEY (id);


--
-- Name: attractions attractions_slug_county_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attractions
    ADD CONSTRAINT attractions_slug_county_id_key UNIQUE (slug, county_id);


--
-- Name: catalog_event_subtypes catalog_event_subtypes_event_type_id_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_subtypes
    ADD CONSTRAINT catalog_event_subtypes_event_type_id_slug_key UNIQUE (event_type_id, slug);


--
-- Name: catalog_event_subtypes catalog_event_subtypes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_subtypes
    ADD CONSTRAINT catalog_event_subtypes_pkey PRIMARY KEY (id);


--
-- Name: catalog_event_types catalog_event_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_types
    ADD CONSTRAINT catalog_event_types_pkey PRIMARY KEY (id);


--
-- Name: catalog_event_types catalog_event_types_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_types
    ADD CONSTRAINT catalog_event_types_slug_key UNIQUE (slug);


--
-- Name: counties counties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.counties
    ADD CONSTRAINT counties_pkey PRIMARY KEY (id);


--
-- Name: counties counties_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.counties
    ADD CONSTRAINT counties_slug_key UNIQUE (slug);


--
-- Name: county_historical_seats county_historical_seats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.county_historical_seats
    ADD CONSTRAINT county_historical_seats_pkey PRIMARY KEY (county_id, historical_seat_id);


--
-- Name: entries entries_slug_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT entries_slug_unique UNIQUE (slug);


--
-- Name: entries entries_unique_entry; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT entries_unique_entry UNIQUE (name, location_id);


--
-- Name: entry_categories entry_categories_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_categories
    ADD CONSTRAINT entry_categories_slug_key UNIQUE (slug);


--
-- Name: entry_category_links entry_category_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_category_links
    ADD CONSTRAINT entry_category_links_pkey PRIMARY KEY (entry_id, category_id);


--
-- Name: entry_members entry_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_members
    ADD CONSTRAINT entry_members_pkey PRIMARY KEY (entry_id, user_id);


--
-- Name: entry_reviews entry_reviews_entry_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_reviews
    ADD CONSTRAINT entry_reviews_entry_id_user_id_key UNIQUE (entry_id, user_id);


--
-- Name: entry_reviews entry_reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_reviews
    ADD CONSTRAINT entry_reviews_pkey PRIMARY KEY (id);


--
-- Name: entry_suggestions entry_suggestions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_suggestions
    ADD CONSTRAINT entry_suggestions_pkey PRIMARY KEY (id);


--
-- Name: entry_tags entry_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_tags
    ADD CONSTRAINT entry_tags_pkey PRIMARY KEY (entry_id, tag_id);


--
-- Name: entry_types entry_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_types
    ADD CONSTRAINT entry_types_name_key UNIQUE (name);


--
-- Name: entry_types entry_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_types
    ADD CONSTRAINT entry_types_pkey PRIMARY KEY (id);


--
-- Name: event_schedule_activities event_schedule_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_activities
    ADD CONSTRAINT event_schedule_activities_pkey PRIMARY KEY (id);


--
-- Name: event_schedule_days event_schedule_days_event_id_schedule_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_days
    ADD CONSTRAINT event_schedule_days_event_id_schedule_date_key UNIQUE (event_id, schedule_date);


--
-- Name: event_schedule_days event_schedule_days_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_days
    ADD CONSTRAINT event_schedule_days_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: geo_locations geo_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.geo_locations
    ADD CONSTRAINT geo_locations_pkey PRIMARY KEY (id);


--
-- Name: historical_seats historical_seats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historical_seats
    ADD CONSTRAINT historical_seats_pkey PRIMARY KEY (id);


--
-- Name: historical_seats historical_seats_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historical_seats
    ADD CONSTRAINT historical_seats_slug_key UNIQUE (slug);


--
-- Name: news_feeds news_feeds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_feeds
    ADD CONSTRAINT news_feeds_pkey PRIMARY KEY (id);


--
-- Name: page_faq_sections page_faq_sections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_faq_sections
    ADD CONSTRAINT page_faq_sections_pkey PRIMARY KEY (id);


--
-- Name: page_faq_sections page_faq_sections_section_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_faq_sections
    ADD CONSTRAINT page_faq_sections_section_key_key UNIQUE (section_key);


--
-- Name: pages pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);


--
-- Name: pages pages_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_slug_key UNIQUE (slug);


--
-- Name: quick_links quick_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quick_links
    ADD CONSTRAINT quick_links_pkey PRIMARY KEY (id);


--
-- Name: entry_categories service_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_categories
    ADD CONSTRAINT service_categories_pkey PRIMARY KEY (id);


--
-- Name: entries services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);


--
-- Name: entries services_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT services_slug_key UNIQUE (slug);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (token_hash);


--
-- Name: settlement_location_types settlement_location_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlement_location_types
    ADD CONSTRAINT settlement_location_types_pkey PRIMARY KEY (id);


--
-- Name: settlement_location_types settlement_location_types_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlement_location_types
    ADD CONSTRAINT settlement_location_types_slug_key UNIQUE (slug);


--
-- Name: settlements settlements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlements
    ADD CONSTRAINT settlements_pkey PRIMARY KEY (id);


--
-- Name: settlements settlements_slug_county_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlements
    ADD CONSTRAINT settlements_slug_county_id_key UNIQUE (slug, county_id);


--
-- Name: site_settings site_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_settings
    ADD CONSTRAINT site_settings_pkey PRIMARY KEY (key);


--
-- Name: tags tags_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_name_key UNIQUE (name);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: entry_categories unique_category_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_categories
    ADD CONSTRAINT unique_category_name UNIQUE (name);


--
-- Name: news_feeds unique_news_feed; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.news_feeds
    ADD CONSTRAINT unique_news_feed UNIQUE (feed_url);


--
-- Name: quick_links unique_quick_link; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quick_links
    ADD CONSTRAINT unique_quick_link UNIQUE (url);


--
-- Name: user_entry_history user_entry_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_entry_history
    ADD CONSTRAINT user_entry_history_pkey PRIMARY KEY (user_id, slug);


--
-- Name: user_favorites user_favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_favorites
    ADD CONSTRAINT user_favorites_pkey PRIMARY KEY (user_id, entity_type, entity_id);


--
-- Name: user_links user_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_links
    ADD CONSTRAINT user_links_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_google_sub_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_google_sub_key UNIQUE (google_sub);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: venue_types venue_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venue_types
    ADD CONSTRAINT venue_types_pkey PRIMARY KEY (id);


--
-- Name: venue_types venue_types_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venue_types
    ADD CONSTRAINT venue_types_slug_key UNIQUE (slug);


--
-- Name: venues venues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venues
    ADD CONSTRAINT venues_pkey PRIMARY KEY (id);


--
-- Name: venues venues_settlement_slug_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venues
    ADD CONSTRAINT venues_settlement_slug_unique UNIQUE (settlement_id, slug);


--
-- Name: weather_astro_daily weather_astro_daily_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_astro_daily
    ADD CONSTRAINT weather_astro_daily_pkey PRIMARY KEY (place_kind, place_id, day);


--
-- Name: weather_daily_archive weather_daily_archive_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_daily_archive
    ADD CONSTRAINT weather_daily_archive_pkey PRIMARY KEY (settlement_id, day);


--
-- Name: weather_desc_translations weather_desc_translations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_desc_translations
    ADD CONSTRAINT weather_desc_translations_pkey PRIMARY KEY (id);


--
-- Name: weather_desc_translations weather_desc_translations_source_text_lang_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_desc_translations
    ADD CONSTRAINT weather_desc_translations_source_text_lang_key UNIQUE (source_text, lang);


--
-- Name: weather_forecast_cache weather_forecast_cache_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_forecast_cache
    ADD CONSTRAINT weather_forecast_cache_pkey PRIMARY KEY (place_kind, place_id);


--
-- Name: weather_obs_hourly weather_obs_hourly_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_obs_hourly
    ADD CONSTRAINT weather_obs_hourly_pkey PRIMARY KEY (settlement_id, observed_at);


--
-- Name: website_category_links website_category_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.website_category_links
    ADD CONSTRAINT website_category_links_pkey PRIMARY KEY (website_id, category_id);


--
-- Name: websites websites_domain_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites
    ADD CONSTRAINT websites_domain_key_key UNIQUE (domain_key);


--
-- Name: websites websites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites
    ADD CONSTRAINT websites_pkey PRIMARY KEY (id);


--
-- Name: attraction_suggestions_one_open; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX attraction_suggestions_one_open ON public.attraction_suggestions USING btree (attraction_id) WHERE (status = 'open'::text);


--
-- Name: entry_categories_featured_order; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX entry_categories_featured_order ON public.entry_categories USING btree (featured_order) WHERE (featured_order IS NOT NULL);


--
-- Name: entry_category_links_one_primary; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX entry_category_links_one_primary ON public.entry_category_links USING btree (entry_id) WHERE is_primary;


--
-- Name: entry_members_one_active_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX entry_members_one_active_owner ON public.entry_members USING btree (entry_id) WHERE (((role)::text = 'owner'::text) AND ((status)::text = 'active'::text));


--
-- Name: entry_suggestions_one_open; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX entry_suggestions_one_open ON public.entry_suggestions USING btree (entry_id) WHERE (status = 'open'::text);


--
-- Name: idx_admin_audit_log_actor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_audit_log_actor ON public.admin_audit_log USING btree (actor_user_id);


--
-- Name: idx_admin_audit_log_occurred; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_audit_log_occurred ON public.admin_audit_log USING btree (occurred_at DESC);


--
-- Name: idx_admin_audit_log_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_audit_log_resource ON public.admin_audit_log USING btree (resource, resource_id);


--
-- Name: idx_admin_sessions_expires; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_sessions_expires ON public.admin_sessions USING btree (expires_at);


--
-- Name: idx_admin_sessions_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_sessions_user ON public.admin_sessions USING btree (user_id);


--
-- Name: idx_entries_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entries_search_vector ON public.entries USING gin (search_vector);


--
-- Name: idx_entries_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entries_type_id ON public.entries USING btree (type_id);


--
-- Name: idx_event_schedule_activities_day; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_schedule_activities_day ON public.event_schedule_activities USING btree (event_day_id);


--
-- Name: idx_event_schedule_days_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_schedule_days_event ON public.event_schedule_days USING btree (event_id);


--
-- Name: idx_sessions_expires; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_expires ON public.sessions USING btree (expires_at);


--
-- Name: idx_sessions_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user ON public.sessions USING btree (user_id);


--
-- Name: idx_venue_types_sort; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_venue_types_sort ON public.venue_types USING btree (sort_order, id);


--
-- Name: idx_venues_settlement; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_venues_settlement ON public.venues USING btree (settlement_id);


--
-- Name: idx_websites_entry; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_websites_entry ON public.websites USING btree (entry_id);


--
-- Name: idx_websites_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_websites_status ON public.websites USING btree (status);


--
-- Name: website_category_links_one_primary; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX website_category_links_one_primary ON public.website_category_links USING btree (website_id) WHERE is_primary;


--
-- Name: entries trg_entries_sync_location_names; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_entries_sync_location_names BEFORE INSERT OR UPDATE OF location_id ON public.entries FOR EACH ROW EXECUTE FUNCTION public.sync_entry_location_names();


--
-- Name: entry_categories trg_sync_entry_category; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_sync_entry_category AFTER UPDATE OF name ON public.entry_categories FOR EACH ROW EXECUTE FUNCTION public.sync_entry_category();


--
-- Name: settlements trg_sync_entry_settlement; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_sync_entry_settlement AFTER UPDATE OF name, name_ro, name_de ON public.settlements FOR EACH ROW EXECUTE FUNCTION public.sync_entry_settlement();


--
-- Name: entry_tags trg_sync_entry_tags; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_sync_entry_tags AFTER INSERT OR DELETE OR UPDATE ON public.entry_tags FOR EACH ROW EXECUTE FUNCTION public.sync_entry_tags();


--
-- Name: admin_audit_log admin_audit_log_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_audit_log
    ADD CONSTRAINT admin_audit_log_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: admin_sessions admin_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: attraction_contributors attraction_contributors_attraction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_contributors
    ADD CONSTRAINT attraction_contributors_attraction_id_fkey FOREIGN KEY (attraction_id) REFERENCES public.attractions(id) ON DELETE CASCADE;


--
-- Name: attraction_contributors attraction_contributors_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_contributors
    ADD CONSTRAINT attraction_contributors_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: attraction_images attraction_images_attraction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_images
    ADD CONSTRAINT attraction_images_attraction_id_fkey FOREIGN KEY (attraction_id) REFERENCES public.attractions(id) ON DELETE CASCADE;


--
-- Name: attraction_suggestions attraction_suggestions_attraction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_suggestions
    ADD CONSTRAINT attraction_suggestions_attraction_id_fkey FOREIGN KEY (attraction_id) REFERENCES public.attractions(id) ON DELETE CASCADE;


--
-- Name: attraction_suggestions attraction_suggestions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attraction_suggestions
    ADD CONSTRAINT attraction_suggestions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: attractions attractions_county_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attractions
    ADD CONSTRAINT attractions_county_id_fkey FOREIGN KEY (county_id) REFERENCES public.counties(id);


--
-- Name: attractions attractions_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attractions
    ADD CONSTRAINT attractions_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.geo_locations(id);


--
-- Name: catalog_event_subtypes catalog_event_subtypes_event_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_event_subtypes
    ADD CONSTRAINT catalog_event_subtypes_event_type_id_fkey FOREIGN KEY (event_type_id) REFERENCES public.catalog_event_types(id) ON DELETE CASCADE;


--
-- Name: counties counties_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.counties
    ADD CONSTRAINT counties_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.geo_locations(id);


--
-- Name: county_historical_seats county_historical_seats_county_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.county_historical_seats
    ADD CONSTRAINT county_historical_seats_county_id_fkey FOREIGN KEY (county_id) REFERENCES public.counties(id) ON DELETE CASCADE;


--
-- Name: county_historical_seats county_historical_seats_historical_seat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.county_historical_seats
    ADD CONSTRAINT county_historical_seats_historical_seat_id_fkey FOREIGN KEY (historical_seat_id) REFERENCES public.historical_seats(id) ON DELETE CASCADE;


--
-- Name: entries entries_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT entries_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.entry_categories(id);


--
-- Name: entries entries_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT entries_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.settlements(id);


--
-- Name: entries entries_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entries
    ADD CONSTRAINT entries_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.entry_types(id) ON DELETE RESTRICT;


--
-- Name: entry_categories entry_categories_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_categories
    ADD CONSTRAINT entry_categories_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.entry_categories(id);


--
-- Name: entry_category_links entry_category_links_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_category_links
    ADD CONSTRAINT entry_category_links_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.entry_categories(id);


--
-- Name: entry_category_links entry_category_links_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_category_links
    ADD CONSTRAINT entry_category_links_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES public.entries(id) ON DELETE CASCADE;


--
-- Name: entry_members entry_members_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_members
    ADD CONSTRAINT entry_members_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES public.entries(id) ON DELETE CASCADE;


--
-- Name: entry_members entry_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_members
    ADD CONSTRAINT entry_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: entry_reviews entry_reviews_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_reviews
    ADD CONSTRAINT entry_reviews_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES public.entries(id) ON DELETE CASCADE;


--
-- Name: entry_reviews entry_reviews_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_reviews
    ADD CONSTRAINT entry_reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: entry_suggestions entry_suggestions_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_suggestions
    ADD CONSTRAINT entry_suggestions_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES public.entries(id) ON DELETE CASCADE;


--
-- Name: entry_suggestions entry_suggestions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_suggestions
    ADD CONSTRAINT entry_suggestions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: entry_tags entry_tags_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_tags
    ADD CONSTRAINT entry_tags_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES public.entries(id) ON DELETE CASCADE;


--
-- Name: entry_tags entry_tags_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entry_tags
    ADD CONSTRAINT entry_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;


--
-- Name: event_schedule_activities event_schedule_activities_event_day_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_activities
    ADD CONSTRAINT event_schedule_activities_event_day_id_fkey FOREIGN KEY (event_day_id) REFERENCES public.event_schedule_days(id) ON DELETE CASCADE;


--
-- Name: event_schedule_activities event_schedule_activities_venue_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_activities
    ADD CONSTRAINT event_schedule_activities_venue_id_fkey FOREIGN KEY (venue_id) REFERENCES public.venues(id) ON DELETE SET NULL;


--
-- Name: event_schedule_days event_schedule_days_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_schedule_days
    ADD CONSTRAINT event_schedule_days_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE CASCADE;


--
-- Name: events events_attraction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_attraction_id_fkey FOREIGN KEY (attraction_id) REFERENCES public.attractions(id) ON DELETE SET NULL;


--
-- Name: events events_default_venue_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_default_venue_id_fkey FOREIGN KEY (default_venue_id) REFERENCES public.venues(id) ON DELETE SET NULL;


--
-- Name: events events_event_subtype_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_event_subtype_id_fkey FOREIGN KEY (event_subtype_id) REFERENCES public.catalog_event_subtypes(id) ON DELETE SET NULL;


--
-- Name: events events_event_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_event_type_id_fkey FOREIGN KEY (event_type_id) REFERENCES public.catalog_event_types(id) ON DELETE RESTRICT;


--
-- Name: events events_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.settlements(id);


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: settlements settlements_county_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlements
    ADD CONSTRAINT settlements_county_id_fkey FOREIGN KEY (county_id) REFERENCES public.counties(id);


--
-- Name: settlements settlements_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlements
    ADD CONSTRAINT settlements_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.geo_locations(id);


--
-- Name: settlements settlements_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settlements
    ADD CONSTRAINT settlements_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.settlements(id);


--
-- Name: user_entry_history user_entry_history_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_entry_history
    ADD CONSTRAINT user_entry_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_favorites user_favorites_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_favorites
    ADD CONSTRAINT user_favorites_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_links user_links_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_links
    ADD CONSTRAINT user_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_preferred_settlement_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_preferred_settlement_fk FOREIGN KEY (preferred_settlement_id) REFERENCES public.settlements(id) ON DELETE SET NULL;


--
-- Name: venues venues_settlement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.venues
    ADD CONSTRAINT venues_settlement_id_fkey FOREIGN KEY (settlement_id) REFERENCES public.settlements(id) ON DELETE CASCADE;


--
-- Name: weather_daily_archive weather_daily_archive_settlement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_daily_archive
    ADD CONSTRAINT weather_daily_archive_settlement_id_fkey FOREIGN KEY (settlement_id) REFERENCES public.settlements(id) ON DELETE CASCADE;


--
-- Name: weather_obs_hourly weather_obs_hourly_settlement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.weather_obs_hourly
    ADD CONSTRAINT weather_obs_hourly_settlement_id_fkey FOREIGN KEY (settlement_id) REFERENCES public.settlements(id) ON DELETE CASCADE;


--
-- Name: website_category_links website_category_links_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.website_category_links
    ADD CONSTRAINT website_category_links_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.entry_categories(id);


--
-- Name: website_category_links website_category_links_website_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.website_category_links
    ADD CONSTRAINT website_category_links_website_id_fkey FOREIGN KEY (website_id) REFERENCES public.websites(id) ON DELETE CASCADE;


--
-- Name: websites websites_approved_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites
    ADD CONSTRAINT websites_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: websites websites_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites
    ADD CONSTRAINT websites_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.entry_categories(id);


--
-- Name: websites websites_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites
    ADD CONSTRAINT websites_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES public.entries(id) ON DELETE SET NULL;


--
-- Name: websites websites_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.websites
    ADD CONSTRAINT websites_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict h6ZSDETeibZDkc1IwBnnkv859xf4EVXBnx66sTNYaXUQEXczoTW7eeY8KUMQdpC

