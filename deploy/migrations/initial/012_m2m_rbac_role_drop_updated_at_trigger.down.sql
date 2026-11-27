-- Restores the (broken) trigger from 002_account for a full rollback.
CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON m2m_rbac_role FOR EACH ROW EXECUTE PROCEDURE updated_at_column();
