-- migrate:up

CREATE INDEX idx_users_name ON users USING btree (name) ;

-- migrate:down

DROP INDEX idx_users_name;