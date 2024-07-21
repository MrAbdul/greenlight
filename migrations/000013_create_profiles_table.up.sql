CREATE TABLE profiles (
                          id BIGSERIAL PRIMARY KEY,
                          user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
                          name TEXT NOT NULL,
                          avatar TEXT NOT NULL,
                          preferred_language INT DEFAULT 1 REFERENCES languages(id) ON DELETE SET DEFAULT,
                          created_at TIMESTAMP(0) WITH TIME ZONE DEFAULT NOW() NOT NULL
);

