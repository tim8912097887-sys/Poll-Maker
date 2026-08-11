-- +goose Up
CREATE TABLE IF NOT EXISTS "polls" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "title" varchar(255) NOT NULL,
    "is_private" boolean NOT NULL,
    "creator_session" varchar(255) NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "started_at" timestamp with time zone NOT NULL,
    "expired_at" timestamp with time zone NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS "polls";