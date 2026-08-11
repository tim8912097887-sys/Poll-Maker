-- +goose Up
CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id TEXT NOT NULL,
    poll_id UUID NOT NULL,
    option_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_poll
        FOREIGN KEY (poll_id)
        REFERENCES polls (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_option
        FOREIGN KEY (option_id)
        REFERENCES poll_options (id)
        ON DELETE CASCADE,
    CONSTRAINT unique_poll_session UNIQUE (poll_id, session_id)
);


-- +goose Down
DROP TABLE votes;
