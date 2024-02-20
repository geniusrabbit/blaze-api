-- Create roles
INSERT INTO rbac_role
  (name, title, context, permissions) VALUES
  -- System roles
  ('system:admin',       'System admins',      NULL, '{"*"}'),
  ('system:manager',     'System manager',     NULL, '{"*.view.*", "*.list.*", "*.count.*", "*.create.*", "*.update.*", "*.delete.*", "*.restore.*", "*.approve.*", "*.reject.*", "role.*.*", "user.reset_password"}'),
  ('system:analyst',     'System analyst',     NULL, '{"*.view.*", "*.list.*", "*.count.*", "user.reset_password"}'),
  ('system:viewer',      'System viewer',      NULL, '{"*.view.*", "*.list.*", "*.count.*", "user.reset_password"}'),
  ('system:compliance',  'System compliance',  NULL, '{"*.view.*", "*.list.*", "*.count.*", "*.approve.*", "*.reject.*", "user.reset_password"}'),
  -- Account roles'
  ('account:admin',      'Account admins',     NULL, '{"*.*.{account|owner}", "user.reset_password"}'),
  ('account:writer',     'Account writer',     NULL, '{"*.{view|list|restore}.{account|owner}", "user.reset_password"}'),
  ('account:analyst',    'Account analyst',    NULL, '{"*.view.{account|owner}", "*.list.{account|owner}", "user.reset_password"}'),
  ('account:viewer',     'Account viewer',     NULL, '{"*.view.{account|owner}", "*.list.{account|owner}", "user.reset_password"}'),
  ('account:compliance', 'Account compliance', NULL, '{"*.view.{account|owner}", "*.list.{account|owner}", "*.approve.{account|owner}", "*.reject.{account|owner}", "user.reset_password"}');

INSERT INTO m2m_account_member_role(member_id, role_id)
  SELECT m.id as member_id, (SELECT id FROM rbac_role WHERE name = 'system:admin') AS role_id
    FROM account_member AS m
    INNER JOIN account_user AS u ON u.email = 'super@project.com'
    WHERE m.user_id = u.id
    ON CONFLICT DO NOTHING;
