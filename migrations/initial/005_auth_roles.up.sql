-- Create roles
INSERT INTO rbac_role
  (name, title, context, permissions) VALUES
  -- System roles
  ('system:admin',       'System admins',      NULL, '{"*"}'),
  ('system:manager',     'System manager',     NULL, '{"*.{view|list|count|create|update|delete|restore|approve|reject|reset}.*", "role.**", "user.password.reset", "account.member.**", "permission.**"}'),
  ('system:analyst',     'System analyst',     NULL, '{"*.{view|list|count}.*", "*.*.{view|list|count}.*", "role.check", "user.password.reset", "permission.list"}'),
  ('system:viewer',      'System viewer',      NULL, '{"*.{view|list|count}.*", "role.check", "user.password.reset", "permission.list"}'),
  ('system:compliance',  'System compliance',  NULL, '{"*.{view|list|count|approve|reject}.*", "*.*.{view|list|count|approve|reject}.*", "role.check", "user.password.reset", "permission.list"}'),
  -- Account roles'
  ('account:admin',      'Account admins',     NULL, '{"*.*.{account|owner}", "*.*.*.{account|owner}", "role.check", "user.password.reset", "permission.list"}'),
  ('account:writer',     'Account writer',     NULL, '{"*.{view|list|restore}.{account|owner}", "*.*.{view|list|restore}.{account|owner}", "role.check", "user.password.reset", "permission.list"}'),
  ('account:analyst',    'Account analyst',    NULL, '{"*.{view|list}.{account|owner}", "*.*.{view|list}.{account|owner}", "role.check", "user.password.reset", "permission.list"}'),
  ('account:viewer',     'Account viewer',     NULL, '{"*.{view|list}.{account|owner}", "*.*.{view|list}.{account|owner}", "role.check", "user.password.reset", "permission.list"}'),
  ('account:compliance', 'Account compliance', NULL, '{"*.{view|list|approve|reject}.{account|owner}", "*.*.{view|list|approve|reject}.{account|owner}", "role.check", "user.password.reset", "permission.list"}');

INSERT INTO m2m_account_member_role(member_id, role_id)
  SELECT m.id as member_id, (SELECT id FROM rbac_role WHERE name = 'system:admin') AS role_id
    FROM account_member AS m
    INNER JOIN account_user AS u ON u.email = 'super@project.com'
    WHERE m.user_id = u.id
    ON CONFLICT DO NOTHING;
