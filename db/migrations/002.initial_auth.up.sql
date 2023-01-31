CREATE SCHEMA auth;
CREATE TABLE auth.user (
    id uuid PRIMARY KEY,
    security_stamp uuid NOT NULL,
    username text UNIQUE NOT NULL,
    email text UNIQUE NOT NULL,
    email_confirmed boolean NOT NULL DEFAULT false,
    password_hash text NOT NULL,
    locked boolean NOT NULL DEFAULT false,
    unsuccessful_login_attempts integer NOT NULL DEFAULT 0
);
