CREATE INDEX IF NOT EXISTS cards_title_idx ON cards USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS cards_tags_idx ON cards USING GIN (tags);