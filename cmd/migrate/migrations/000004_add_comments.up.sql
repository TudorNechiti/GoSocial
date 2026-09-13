CREATE TABLE IF NOT EXISTS comments(
    id bigserial PRIMARY KEY,
    post_id bigint NOT NULL REFERENCES posts(id),
    user_id bigint NOT NULL REFERENCES users(id),
    content text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT now()
);
