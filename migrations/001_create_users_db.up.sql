CREATE TABLE IF NOT EXISTS "users" (
    "id" uuid NOT NULL,
    "first_name" varchar(255) NOT NULL,
    "last_name" varchar(255) NOT NULL,
    "nickname" varchar(255) NOT NULL,
    "password" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "country" varchar(255) NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    PRIMARY KEY ("id", "email")
);

