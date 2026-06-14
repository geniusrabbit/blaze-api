INSERT INTO account_base (approve_status, title, description) VALUES(1, 'system', 'System account');
INSERT INTO account_user (approve_status, email, password)    VALUES(1, 'super@project.com', '$2a$10$1Eh45WHxlcO4Tb90mVc62Og.T8Q81QJTpGyF9pT1EwSkQUPA8XyAS');

INSERT INTO account_member (user_id, account_id, approve_status, is_admin)
  SELECT MAX(user_id) AS user_id, MAX(account_id) AS account_id, 1, 't' AS is_admin FROM (
    SELECT id AS user_id, 0 AS account_id FROM account_user WHERE email = 'super@project.com'
    UNION ALL
    SELECT 0 AS user_id, id AS account_id FROM account_base WHERE title = 'system'
  ) AS t;

-- Create new oauth client
INSERT INTO auth_client (
  id,
  account_id,
  user_id,
  title,
  secret,
  redirect_uris,
  grant_types,
  response_types,
  scope,
  audience,
  subject_type,
  allowed_cors_origins,
  public,
  expires_at
)
SELECT
  'system',
  account_id,
  user_id,
  'system gate',
  '$2a$10$IxMdI6d.LIRZPpSfEwNoeu4rY3FhDREsxFJXikcgdRRAStxUlsuEO',
  '{http://localhost:3846/callback}',
  '{password,client_credentials,authorization_code,implicit,refresh_token}',
  '{id_token,code,token}',
  'superuser menu openid offline',
  '{}',
  'public',
  '{http://localhost:3846/users}',
  FALSE,
  NOW() + interval '10 years'
FROM (
  SELECT MAX(user_id) AS user_id, MAX(account_id) AS account_id FROM (
    SELECT id AS user_id, 0 AS account_id FROM account_user WHERE email = 'super@project.com'
    UNION ALL
    SELECT 0 AS user_id, id AS account_id FROM account_base WHERE title = 'system'
  ) AS t
) AS t;
