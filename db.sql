CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

INSERT INTO channels (name) VALUES ('default');


CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    channel_name TEXT NOT NULL,
    user_name TEXT NOT NULL,
    message_text TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (channel_name) REFERENCES channels (name),
    FOREIGN KEY (user_name) REFERENCES users (username)
);