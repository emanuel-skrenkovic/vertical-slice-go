CREATE TABLE auth.session (
    id uuid,
    user_id uuid,
    expires_at timestamptz,

    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES auth.user (id)
)