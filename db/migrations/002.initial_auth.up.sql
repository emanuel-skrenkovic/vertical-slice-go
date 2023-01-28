CREATE SCHEMA auth;
CREATE TABLE auth.user (
    id uuid PRIMARY KEY,
    security_stamp uuid,
    username text UNIQUE,
    email text UNIQUE,
    email_confirmed boolean,
    password_hash text,
    locked boolean,
    unsuccessful_login_attempts integer
);
