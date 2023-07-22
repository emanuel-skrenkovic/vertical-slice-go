CREATE TABLE session_invitation (
       id uuid PRIMARY KEY NOT NULL,
       session_id text NOT NULL,
       inviter_id uuid NOT NULL,
       invitee_id uuid NOT NULL,
       created_at timestamptz NOT NULL,

       CONSTRAINT fk_game_session FOREIGN KEY (session_id) REFERENCES game_session(id)
);
