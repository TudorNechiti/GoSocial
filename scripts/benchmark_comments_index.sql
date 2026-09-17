-- Benchmark: does a trigram GIN index speed up content search on comments?
-- Run interactively, section by section (this is NOT a migration).

-- 0. required extension for gin_trgm_ops
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 1. seed 100k comments with realistic variety:
--    post_id / user_id are picked from IDs that actually exist (so this
--    never trips the FK constraint, no matter how many posts/users you
--    really have seeded), content built from random phrase fragments,
--    plus a rare needle phrase planted in ~0.1% of rows so a search can
--    actually be selective.
WITH post_ids AS (
    SELECT array_agg(id) AS ids FROM posts
),
user_ids AS (
    SELECT array_agg(id) AS ids FROM users
)
INSERT INTO comments (post_id, user_id, content)
SELECT
    p.ids[floor(random() * array_length(p.ids, 1) + 1)],
    u.ids[floor(random() * array_length(u.ids, 1) + 1)],
    CASE
        WHEN random() < 0.001
            THEN 'zzframboyant this is a super rare needle phrase ' || gs
        ELSE
            (ARRAY[
                'Great post, really enjoyed reading this',
                'I disagree with this take entirely',
                'lol what is even happening here',
                'Extra cringe comment... Wow!',
                'This changed my perspective completely',
                'Not sure this is accurate but ok',
                'First! Also nice work on this one',
                'Could you elaborate on this point more'
            ])[floor(random() * 8 + 1)] || ' #' || gs
    END
FROM generate_series(1, 100000) AS gs, post_ids p, user_ids u;

-- 2. BEFORE creating the index: run this and note the plan/timing
--    (should show a Seq Scan)
EXPLAIN ANALYZE
SELECT * FROM comments WHERE content ILIKE '%zzframboyant%';

-- 3. now create the index
CREATE INDEX idx_comments_content ON comments USING gin (content gin_trgm_ops);

-- 4. AFTER: run the same query again and compare
--    (should show a Bitmap Heap Scan using idx_comments_content)
EXPLAIN ANALYZE
SELECT * FROM comments WHERE content ILIKE '%zzframboyant%';
