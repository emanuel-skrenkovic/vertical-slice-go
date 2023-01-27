CREATE TABLE game_session (
       id text PRIMARY KEY NOT NULL,
       owner_id uuid NOT NULL,
       player_1_id uuid,
       player_2_id uuid,
       game_id uuid,
       active boolean NOT NULL DEFAULT FALSE,
       name text NOT NULL
)
