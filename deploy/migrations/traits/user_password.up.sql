-- Password trait migration — adds password columns to account_user.
-- Copy to your project's migrations/initial/ if your user embeds user.Password.
-- GORM AutoMigrate alternative: db.AutoMigrate(&MyUser{}) if MyUser embeds user.Password.

ALTER TABLE account_user
  ADD COLUMN IF NOT EXISTS password VARCHAR(128) NOT NULL DEFAULT ''
    CHECK (LENGTH(password) = 0 OR LENGTH(password) > 5),
  ADD COLUMN IF NOT EXISTS required_password_reset BOOL NOT NULL DEFAULT FALSE;
