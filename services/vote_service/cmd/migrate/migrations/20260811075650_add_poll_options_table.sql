-- +goose Up
CREATE TABLE IF NOT EXISTS "poll_options" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "poll_id" uuid NOT NULL,
    "option_text" varchar(255) NOT NULL,
    "vote_counts" integer DEFAULT 0 NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT fk_poll
        FOREIGN KEY ("poll_id") REFERENCES "polls" ("id")
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS "poll_options";
