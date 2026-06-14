-- Create roles
INSERT INTO rbac_role
  (name, title, description, context, access_level, permissions) VALUES
  -- System roles
  (
    'system:admin',
    'System admins',
    'System administrators have full access to all system resources',
    NULL,
    100,
    '{"*"}'
  ),
  (
    'system:manager',
    'System manager',
    'System managers have full access to all system resources except for some sensitive operations',
    NULL,
    90,
    '{"*.{view|list|count|create|update|delete|restore|approve|reject|reset}.*", "role.**", "user.password.reset", "account.member.**", "permission.**"}'
  ),
  (
    'system:analyst',
    'System analyst',
    'System analysts have read-only access to all system resources',
    NULL,
    80,
    '{"*.{view|list|count}.*", "*.*.{view|list|count}.*", "role.check", "user.password.reset", "permission.list"}'
  ),
  (
    'system:compliance', 
    'System compliance',
    'System compliance have access to all system resources with approve/reject permissions',
    NULL,
    70,
    '{"*.{view|list|count|approve|reject}.*", "*.*.{view|list|count|approve|reject}.*", "role.check", "user.password.reset", "permission.list"}'
  ),
  (
    'system:viewer',
    'System viewer',
    'System viewers have read-only access to all system resources',
    NULL,
    60,
    '{"*.{view|list|count}.*", "role.check", "user.password.reset", "permission.list"}'
  ),
  -- Account roles'
  (
    'account:admin',
    'Account admins',
    'Account administrators have full access to all account resources',
    NULL,
    2,
    '{"*.*.{account|owner}", "*.*.*.{account|owner}", "role.check", "user.password.reset.{account|owner}", "permission.list", "account.member.roles.set.account", "certificate.view.statistic.{account|owner}"}'
  ),
  (
    'account:writer',
    'Account writer',
    'Account writers have full access to all account resources except for some sensitive operations',
    NULL,
    2,
    '{"*.{view|list|restore}.{account|owner}", "*.*.{view|list|restore}.{account|owner}", "role.check", "user.password.reset", "permission.list", "certificate.view.statistic.{account|owner}"}'
  ),
  (
    'account:analyst',
    'Account analyst',
    'Account analysts have read-only access to all account resources and analytics',
    NULL,
    2,
    '{"*.{view|list}.{account|owner}", "*.*.{view|list}.{account|owner}", "role.check", "user.password.reset", "permission.list", "certificate.view.statistic.{account|owner}"}'
  ),
  (
    'account:viewer',
    'Account viewer',
    'Account viewers have read-only access to all account resources except analytics',
    NULL,
    2,
    '{"*.{view|list}.{account|owner}", "*.*.{view|list}.{account|owner}", "role.check", "user.password.reset", "permission.list", "certificate.view.statistic.{account|owner}"}'
  ),
  (
    'account:compliance',
    'Account compliance',
    'Account compliance have access to all account resources with approve/reject permissions',
    NULL,
    2,
    '{"*.{view|list|approve|reject}.{account|owner}", "*.*.{view|list|approve|reject}.{account|owner}", "role.check", "user.password.reset", "permission.list"}'
  )
ON CONFLICT (name) DO UPDATE
  SET title        = EXCLUDED.title,
      description  = EXCLUDED.description,
      context      = EXCLUDED.context,
      access_level = EXCLUDED.access_level,
      permissions  = EXCLUDED.permissions;

INSERT INTO m2m_account_member_role(member_id, role_id)
  SELECT m.id as member_id, (SELECT id FROM rbac_role WHERE name = 'system:admin') AS role_id
    FROM account_member AS m
    INNER JOIN account_user AS u ON u.email = 'super@project.com'
    WHERE m.user_id = u.id
    ON CONFLICT DO NOTHING;
