ALTER TABLE "user"
    DROP CONSTRAINT IF EXISTS email_format,
    DROP CONSTRAINT IF EXISTS nickname_length;