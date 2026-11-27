-- m2m_rbac_role has no updated_at column, but 002_account created the
-- updated_at trigger on it, so every UPDATE of the table failed with
-- `record "new" has no field "updated_at"`. The link table only needs
-- created_at, so the trigger is dropped.
DROP TRIGGER IF EXISTS updated_at_triger ON m2m_rbac_role;
