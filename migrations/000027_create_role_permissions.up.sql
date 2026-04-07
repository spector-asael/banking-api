CREATE TABLE IF NOT EXISTS role_permissions (
    account_type_id SMALLINT NOT NULL REFERENCES account_types(id) ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (account_type_id, permission_id)
);