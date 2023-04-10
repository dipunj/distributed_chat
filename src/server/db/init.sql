GRANT all privileges ON DATABASE postgres to postgres;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres GRANT ALL ON TABLES TO postgres;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres GRANT ALL ON SEQUENCES TO postgres;

GRANT ALL PRIVILEGES ON DATABASE postgres TO postgres;

-- add index on message_type
-- add index on group_name
-- add index on message_id
CREATE TABLE IF NOT EXISTS MESSAGES (
	id SERIAL primary key,
	message_type varchar(20),
	client_id varchar(55),
	sender_name varchar(128),
	group_name varchar(128),
	content varchar(90),
	client_sent_at timestamp with time zone NOT NULL,
	server_received_at timestamp with time zone NOT NULL,
	vector_ts BIGINT ARRAY [] NOT NULL,
	parent_msg_id int references MESSAGES (id) -- can be NULL
);

-- for faster access on each group
CREATE INDEX IF NOT EXISTS group_name_msg_type on MESSAGES(group_name, message_type);

-- a username can have only one reaction per message
-- hence the same sender_name and parent_msg_id must not occur twice in the table
CREATE UNIQUE INDEX IF NOT EXISTS unique_reactions ON MESSAGES (message_type, parent_msg_id, sender_name);

-- timestamps should also be unique
CREATE UNIQUE INDEX IF NOT EXISTS unique_ts ON MESSAGES (vector_ts);