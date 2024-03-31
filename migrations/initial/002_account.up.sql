--
-- Init subproject structure
--
-- Original users admins/managers/etc in base database 
--

-- ----------------------------------------------------------------------------
-- Account / User / Profile
-- ----------------------------------------------------------------------------

CREATE TABLE account_user
( id                      BIGSERIAL                   PRIMARY KEY
, approve_status          INTEGER                     NOT NULL        DEFAULT 0

, email                   VARCHAR(128)                NOT NULL        CHECK (email ~* '^[^\s]+$')  UNIQUE
, password                VARCHAR(128)                NOT NULL        CHECK (LENGTH(password) = 0 OR LENGTH(password) > 5)

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON account_user FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON account_user
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();


-- Reset password token
CREATE TABLE IF NOT EXISTS account_user_password_reset
( user_id                 BIGINT                      NOT NULL        REFERENCES account_user (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT

, token                   VARCHAR(128)                NOT NULL        PRIMARY KEY CHECK (token ~* '^[^\s]+$')

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, expires_at              TIMESTAMP                   NOT NULL        DEFAULT NOW() + INTERVAL '30 minutes'
);

CREATE INDEX idx_account_user_password_reset_expires_at ON
    account_user_password_reset (expires_at);

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON account_user_password_reset
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();

CREATE TRIGGER keep_for_one_week
    BEFORE UPDATE OR INSERT ON account_user_password_reset
    FOR EACH ROW
        EXECUTE PROCEDURE keep_for(expires_at, '1 week');

-- Account this is the general entity which links all objects with one account
-- Account can be linked with planty of accounts with different access permissions
-- Account have to have atleast one user who is the admin of the account
CREATE TABLE account_base
( id                      BIGSERIAL                   PRIMARY KEY
, approve_status          INTEGER                     NOT NULL        DEFAULT 0

, title                   VARCHAR(128)                NOT NULL        CHECK (title ~* '^[^\s]+')
, description             TEXT

, logo_uri                VARCHAR(1024)               NOT NULL        DEFAULT ''
, policy_uri              VARCHAR(1024)               NOT NULL        DEFAULT ''
, tos_uri                 VARCHAR(1024)               NOT NULL        DEFAULT ''
, client_uri              VARCHAR(1024)               NOT NULL        DEFAULT ''
, contacts                TEXT[]

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON account_base FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON account_base
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();

-- ----------------------------------------------------------------------------
-- Roles and permissions
-- ----------------------------------------------------------------------------

CREATE TABLE rbac_role
( id                  BIGSERIAL                   PRIMARY KEY

-- Name of the permission to matching
, name                VARCHAR(256)                NOT NULL      CHECK (name ~* '^[\w\d@_:\.-]+$') UNIQUE

, title               VARCHAR(256)                NOT NULL
, description         TEXT                        NOT NULL      DEFAULT ''
, context             JSONB                                     DEFAULT NULL  -- {model:flags,@custom:value}

-- [`user.view.owner`, `account.*.owner`]
, permissions         TEXT[]

, access_level        INTEGER                     NOT NULL      DEFAULT 0

, created_at          TIMESTAMPTZ                 NOT NULL      DEFAULT NOW()
, updated_at          TIMESTAMPTZ                 NOT NULL      DEFAULT NOW()
, deleted_at          TIMESTAMPTZ
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON rbac_role FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

-- Role link child permissions
CREATE TABLE m2m_rbac_role
( parent_role_id      BIGINT                      NOT NULL      REFERENCES rbac_role (id) MATCH SIMPLE
                                                                      ON UPDATE NO ACTION
                                                                      ON DELETE RESTRICT
, child_role_id       BIGINT                      NOT NULL      REFERENCES rbac_role (id) MATCH SIMPLE
                                                                      ON UPDATE NO ACTION
                                                                      ON DELETE RESTRICT

, created_at          TIMESTAMPTZ                 NOT NULL      DEFAULT NOW()

, PRIMARY KEY(parent_role_id, child_role_id)
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON m2m_rbac_role FOR EACH ROW EXECUTE PROCEDURE updated_at_column();


CREATE TABLE account_member
( id                  BIGSERIAL                   PRIMARY KEY
, approve_status      INTEGER                     NOT NULL        DEFAULT 0

-- Link to account
, account_id          BIGINT                      NOT NULL      REFERENCES account_base (id) MATCH SIMPLE
                                                                      ON UPDATE NO ACTION
                                                                      ON DELETE RESTRICT

-- The user linked to the account
, user_id             BIGINT                      NOT NULL      REFERENCES account_user (id) MATCH SIMPLE
                                                                      ON UPDATE NO ACTION
                                                                      ON DELETE RESTRICT

-- Superuser permissions for the current account
-- Despite of that optinion that better to use roles as the only way of permission issue
--   the Owner flag in most of cases is very useful approach which prevent many problems related to
--   permission updates.
-- Admin permission restricted by some limits which available only to superusers and managers.
, is_admin            BOOL                        NOT NULL      DEFAULT FALSE

, created_at          TIMESTAMPTZ                 NOT NULL      DEFAULT NOW()
, updated_at          TIMESTAMPTZ                 NOT NULL      DEFAULT NOW()
, deleted_at          TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_account_member_unique_account_user
    ON account_member (account_id, user_id);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON account_member FOR EACH ROW EXECUTE PROCEDURE updated_at_column();


-- Link account member with user
CREATE TABLE m2m_account_member_role
( member_id           BIGINT                                    REFERENCES account_member (id) MATCH SIMPLE
                                                                      ON UPDATE NO ACTION
                                                                      ON DELETE RESTRICT
, role_id             BIGINT                      NOT NULL      REFERENCES rbac_role (id) MATCH SIMPLE
                                                                      ON UPDATE NO ACTION
                                                                      ON DELETE RESTRICT

, created_at          TIMESTAMPTZ                 NOT NULL      DEFAULT NOW()

, PRIMARY KEY (member_id, role_id)
);
