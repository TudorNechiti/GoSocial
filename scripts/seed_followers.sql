-- Seed random follower relationships so GetUserFeed can be tested.
-- Run interactively (this is NOT a migration).

WITH user_ids AS (
    SELECT array_agg(id) AS ids FROM users
),
pairs AS (
    SELECT
        u.ids[floor(random() * array_length(u.ids, 1) + 1)] AS user_id,
        u.ids[floor(random() * array_length(u.ids, 1) + 1)] AS follower_id
    FROM generate_series(1, 5000) AS gs, user_ids u
)
INSERT INTO followers (user_id, follower_id)
SELECT user_id, follower_id
FROM pairs
WHERE user_id <> follower_id          -- avoid self-follow check constraint
ON CONFLICT (user_id, follower_id) DO NOTHING;

-- sanity check: pick a real user, see who they follow and who follows them
SELECT * FROM followers WHERE follower_id = (SELECT id FROM users LIMIT 1);
