-- Permission list
--
-- view    - 
-- list    - 
-- create  - 
-- update  - 
-- delete  - 
-- restore - 
-- approve - 
-- reject  - 

-- Create roles
INSERT INTO rbac_role
  (name, title, type, context) VALUES
  ('system:admin',       'System admins',      'role', NULL),
  ('system:manager',     'System manager',     'role', NULL),
  ('system:analyst',     'System analyst',     'role', NULL),
  ('system:viewer',      'System viewer',      'role', NULL),
  ('system:compliance',  'System compliance',  'role', NULL),
  ('account:admin',      'Account admins',     'role', NULL),
  ('account:writer',     'Account writer',     'role', NULL),
  ('account:analyst',    'Account analyst',    'role', NULL),
  ('account:viewer',     'Account viewer',     'role', NULL),
  ('account:compliance', 'Account compliance', 'role', NULL);

-- Create permisssions
INSERT INTO rbac_role
  (name, title, type, context)
  SELECT name, '' AS title, 'permission' AS type, ('{"object":"' || object || '","cover":"' || cover || '"}')::jsonb AS context
    FROM  unnest(array[
            'view','list','create','update','delete','restore','approve','reject'
          ]) AS name,
          LATERAL unnest(array[
            'model:User',
            'model:Account',
            'model:AccountMember',
            'model:Role',
            'model:AuthClient'
          , 'model:Option'
          ]) AS object,
          LATERAL unnest(array[
            'none',
            'account', -- Modificator for access for the whole account
            'system'   -- Modificator for access to the all objects in the system
          ]) AS cover
    ON CONFLICT DO NOTHING;

INSERT INTO rbac_role
  (name, title, type, context)
  SELECT name, '' AS title, 'permission' AS type, ('{"object":"' || object || '","cover":"' || cover || '"}')::jsonb AS context
    FROM  unnest(array['view','list']) AS name,
          LATERAL unnest(array['model:HistoryAction']) AS object,
          LATERAL unnest(array['none', 'account', 'system']) AS cover
    ON CONFLICT DO NOTHING;

INSERT INTO rbac_role
  (name, title, type, context)
  SELECT name, '' AS title, 'permission' AS type, ('{"object":"' || object || '","cover":"' || cover || '"}')::jsonb AS context
    FROM  unnest(array['check']) AS name,
          LATERAL unnest(array['model:Role']) AS object,
          LATERAL unnest(array['none', 'account', 'system']) AS cover
    ON CONFLICT DO NOTHING;


-- Link all permissions to the system role
INSERT INTO m2m_rbac_role (parent_role_id, child_role_id)
  SELECT (SELECT id FROM rbac_role WHERE name = 'system:admin') AS parent_role_id, id AS child_role_id
    FROM rbac_role WHERE type = 'permission' AND COALESCE(context->>'cover', '') = 'system' AND deleted_at IS NULL
  ON CONFLICT DO NOTHING;


INSERT INTO m2m_account_member_role(member_id, role_id)
  SELECT m.id as member_id, (SELECT id FROM rbac_role WHERE name = 'system:admin') AS role_id
    FROM account_member AS m
    INNER JOIN account_user AS u ON u.email = 'super@project.com'
    WHERE m.user_id = u.id
    ON CONFLICT DO NOTHING;
