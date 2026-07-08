-- Username trait migration — adds username column to account_user.
-- Copy to your project's migrations/initial/ if your user embeds user.Username.
-- GORM AutoMigrate alternative: db.AutoMigrate(&MyUser{}) if MyUser embeds user.Username.

ALTER TABLE account_user
  ADD COLUMN IF NOT EXISTS username VARCHAR(64) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS account_user_username_uniq
  ON account_user (username)
  WHERE username != '' AND deleted_at IS NULL;
