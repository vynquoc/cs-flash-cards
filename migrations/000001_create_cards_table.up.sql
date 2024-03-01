CREATE TABLE IF NOT EXISTS cards (
    id bigserial PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_DATE,
    title text NOT NULL,
    tags text[] NOT NULL,
    content text NOT NULL,
    next_review_date DATE NOT NULL,
    code_snippet JSONB
);