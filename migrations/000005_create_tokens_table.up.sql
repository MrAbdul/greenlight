CREATE TABLE IF NOT EXISTS tokens (
                                      hash bytea PRIMARY KEY,
--     if we delete a user from users table, then their tokens shall be deleted
--                                       The user_id column will contain the ID of the user associated with the token.
--                                       We use the REFERENCES user syntax to create a foreign key constraint against
--                                       the primary key of our users table, which ensures that any value in the user_id column has a corresponding id entry in our users table.
                                      user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
                                      expiry timestamp(0) with time zone NOT NULL,
                                      scope text NOT NULL
);