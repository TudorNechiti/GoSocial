ALTER TABLE posts ADD COLUMN IF NOT EXISTS tags varchar(100)[];
ALTER TABLE posts ADD COLUMN IF NOT EXISTS updated_at timestamp(0) with time zone NOT NULL DEFAULT now();
