CREATE TABLE IF NOT EXISTS users (
                                     id bigserial PRIMARY KEY,
                                     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
                                     name text NOT NULL,
--     citext is a case insestive text type, saves as inputed without changing case, but comparisons are case insensitive
                                     email citext UNIQUE NOT NULL,
--     The password_hash column has the type bytea (binary string). In this column we’ll store a one-way hash of the user’s password generated using bcrypt — not the plaintext password itself.
                                     password_hash bytea NOT NULL,
                                     activated bool NOT NULL,
--     We’ve also included a version number column, which we will increment each time a user record is updated. This will allow us to use optimistic locking to prevent race conditions when updating user records, in the same way that we did with movies earlier in the book.
                                     version integer NOT NULL DEFAULT 1
);