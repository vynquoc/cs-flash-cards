CREATE TABLE IF NOT EXISTS cards (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    tags text[] NOT NULL,
    content text NOT NULL,
    next_review_date timestamp(0) with time zone NOT NULL,
    code_snippet JSONB
);