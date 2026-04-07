// Filename: internal/data/permissions.go
package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

// We will have the permissions in a slice which we will be able to search
type Permissions []string

// Is the permission code found for the Permissions slice
func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

// Setup our model
type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser looks up the user's account_type and returns the permissions for that role.
func (p PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	// UPDATED: Now referencing role_permissions consistently
	query := `
        SELECT permissions.code
        FROM permissions
        INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id
        INNER JOIN users ON users.account_type = role_permissions.account_type_id
        WHERE users.id = $1
    `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForAccountType replaces AddForUser. You now assign permissions to the role itself.
func (p PermissionModel) AddForAccountType(accountTypeID int, codes ...string) error {
	query := `
		INSERT INTO role_permissions (account_type_id, permission_id)
		SELECT $1, permissions.id FROM permissions 
		WHERE permissions.code = ANY($2)
		ON CONFLICT DO NOTHING -- Prevents errors if you accidentally assign a permission twice
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, query, accountTypeID, pq.Array(codes))

	return err
}
