CREATE TABLE auth.activation_code (
    id BIGSERIAL PRIMARY KEY,
    user_id uuid,
    security_stamp uuid,
    expires_at timestamptz,
    sent_at timestamptz,
    token text,
    used boolean,

    CONSTRAINT fk_user_id FOREIGN KEY (user_Id) REFERENCES auth.user (id)
);
