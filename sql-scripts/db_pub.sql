-- ---------------------------------------------------------
--  DATABASE
-- ---------------------------------------------------------

CREATE DATABASE momentic
    OWNER momentic_admin
    ENCODING 'UTF8'
    LC_COLLATE = 'C'
    LC_CTYPE = 'C'
    TEMPLATE template0;

-- ---------------------------------------------------------
--  TYPES
-- ---------------------------------------------------------

CREATE TYPE friendship_status AS ENUM ('friends',
    'pending1', 'pending2',
    'blocked1',  'blocked2', 'blocked12'
);

CREATE TYPE reaction_kind AS ENUM ('heart', 'flame', 'funny', 'angry');

-- ---------------------------------------------------------
--  TABLES, INDEXES
-- ---------------------------------------------------------

CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    nickname VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    rating INTEGER NOT NULL DEFAULT 0 CHECK (rating >= 0),
    max_streak INTEGER NOT NULL DEFAULT 0 CHECK (max_streak >= 0),
    current_streak INTEGER NOT NULL DEFAULT 0 CHECK (current_streak >= 0),
    max_reactions INTEGER NOT NULL DEFAULT 0 CHECK (max_reactions >= 0),
    avatar_filepath TEXT,
    bio VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname);

CREATE TABLE videos (
    video_id BIGSERIAL PRIMARY KEY,
    filepath TEXT NOT NULL,
    author_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    description VARCHAR(70) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_videos_author ON videos(author_id);

CREATE TABLE friendships (
    user_id1 BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    user_id2 BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status friendship_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id1, user_id2),
    CHECK (user_id1 < user_id2)
);

CREATE INDEX IF NOT EXISTS idx_friendships_user1 ON friendships(user_id1);

CREATE TABLE reactions (
    user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos(video_id) ON DELETE CASCADE,
    reaction reaction_kind NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, video_id)
);

CREATE INDEX IF NOT EXISTS idx_reactions_video ON reactions(video_id);
CREATE INDEX IF NOT EXISTS idx_reactions_user ON reactions(user_id);

-- ---------------------------------------------------------
--  FUNCTIONS
-- ---------------------------------------------------------

--TODO: write functions for working with friendships table

-- ---------------------------------------------------------
--  TRIGGERS
-- ---------------------------------------------------------

