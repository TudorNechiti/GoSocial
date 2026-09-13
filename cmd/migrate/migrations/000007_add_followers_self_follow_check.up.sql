ALTER TABLE followers
    ADD CONSTRAINT followers_no_self_follow CHECK (user_id <> follower_id);
