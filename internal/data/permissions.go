package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// Permissions holds the permission codes for a single user.
type Permissions []string

// Include checks whether the Permissions slice contains a specific permission code.
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}

type PermissionMode1 struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// GetAllForUser returns all permission codes for a specific user in a Permissions slice.
func (m PermissionMode1) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT permissions.code
		FROM permissions
			INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
			INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			m.ErrorLog.Println(err)
		}
	}()

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
