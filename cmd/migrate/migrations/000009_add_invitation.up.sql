CREATE TABLE IF NOT EXISTS user_invitations (
    token bytea PRIMARY KEY,
    user_id bigint NOT NULL,
    expiry timestamp(0) with time zone NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT FALSE;