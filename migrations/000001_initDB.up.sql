CREATE TABLE users
(
    id SERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL
);


CREATE TABLE chats
(
    id SERIAL PRIMARY KEY,
    name VARCHAR (64),
    owner_id INTEGER REFERENCES users(id)
);


CREATE TABLE chat_members
(
    user_id INTEGER REFERENCES users(id),
    chat_id INTEGER REFERENCES chats(id),
    PRIMARY KEY (user_id, chat_id)
);

CREATE TABLE messages
(
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id),
    user_id INTEGER REFERENCES users(id),
    text VARCHAR(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE sessions
(
    hashed_token TEXT PRIMARY KEY,
    user_id      INTEGER REFERENCES users (id),
    expires_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_chat_members_chat_id ON chat_members (chat_id, user_id);
CREATE INDEX idx_messages_chat_id_id ON messages (chat_id, id DESC);