CREATE TABLE "user" (
    id bigserial primary key not null,
    external_id character varying(256),
    name character varying(64) not null,
    email character varying(255)
);

CREATE TABLE "link" (
    id bigserial primary key not null,
    external_id character varying(256),
    level integer,
    name character varying(64) not null,
    description character varying(255),
    origin_url character varying(255),
    subscribers_count integer default 0,
    watchers_count integer default 0,
    topics character varying(255),
    updated bigint
);

CREATE EXTENSION pg_trgm;
CREATE INDEX index_users_full_name ON link using gin ((name || ' ' || description) gin_trgm_ops);

-- ALTER TABLE public.link OWNER TO postgres;
-- CREATE SEQUENCE public.link_id_seq
--     AS integer
--     START WITH 1
--     INCREMENT BY 1
--     NO MINVALUE
--     NO MAXVALUE
--     CACHE 1;

-- ALTER SEQUENCE public.link_id_seq OWNER TO postgres;
-- ALTER SEQUENCE public.link_id_seq OWNED BY public.link.id;
-- ALTER TABLE ONLY public.link ALTER COLUMN id SET DEFAULT nextval('public.link_id_seq'::regclass);
-- ALTER TABLE ONLY public.link ADD CONSTRAINT link_pkey PRIMARY KEY (id);



-- ALTER TABLE "user" OWNER TO postgres;
-- CREATE SEQUENCE user_id_seq
--     AS integer
--     START WITH 1
--     INCREMENT BY 1
--     NO MINVALUE
--     NO MAXVALUE
--     CACHE 1;

-- ALTER SEQUENCE user_id_seq OWNER TO postgres;
-- ALTER SEQUENCE user_id_seq OWNED BY "user".id;
-- ALTER TABLE ONLY "user" ALTER COLUMN id SET DEFAULT nextval('user_id_seq'::regclass);
-- ALTER TABLE ONLY "user" ADD CONSTRAINT user_pkey PRIMARY KEY (id);

