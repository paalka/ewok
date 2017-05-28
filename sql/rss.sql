CREATE SCHEMA rss;
SET SEARCH_PATH TO rss;

CREATE TABLE rss_feed (
       id SERIAL PRIMARY KEY NOT NULL,
       url text NOT NULL,
       title text NOT NULL,

       last_updated timestamptz NOT NULL
);

CREATE TABLE rss_item (
       id SERIAL PRIMARY KEY NOT NULL,
       title text NOT NULL,
       link text NOT NULL,
       description text NOT NULL,

       publish_date timestamptz NOT NULL
);

GRANT USAGE ON SCHEMA rss TO rss;
GRANT SELECT,UPDATE,INSERT ON ALL TABLES IN SCHEMA rss TO rss;
GRANT SELECT,UPDATE ON rss.rss_item_id_seq TO rss;
