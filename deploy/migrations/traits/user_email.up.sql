-- Email trait migration — adds email column to account_user.
-- Copy to your project's migrations/initial/ if your user embeds user.Email.
-- GORM AutoMigrate alternative: db.AutoMigrate(&MyUser{}) if MyUser embeds user.Email.

ALTER TABLE account_user
  ADD COLUMN IF NOT EXISTS email VARCHAR(128) NOT NULL DEFAULT ''
    CHECK (email ~* '^[^\s]+$' OR email = '');

CREATE UNIQUE INDEX IF NOT EXISTS account_user_email_uniq
  ON account_user (email)
  WHERE email != '' AND deleted_at IS NULL;
