ALTER TABLE "user"
    ADD CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    ADD CONSTRAINT nickname_length CHECK (char_length(nickname) >= 3 AND char_length(nickname) <= 50);