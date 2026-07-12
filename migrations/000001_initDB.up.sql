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
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMP DEFAULT NULL
);
