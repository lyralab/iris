package postgresql

import (
	"database/sql"

	"github.com/root-ali/iris/pkg/scheduler/cache_receptors"
)
import _ "github.com/lib/pq"

func (s *Storage) GetGroupsNumbers(group ...string) ([]cache_receptors.GroupWithMobiles, error) {
	var gms []cache_receptors.GroupWithMobiles
	var err error
	// Prepare SQL query with a conditional WHERE clause.
	query := `
    SELECT
        g.id   AS group_id,
        g.name AS group_name,
        u.id AS user_id ,
        u.mobile AS mobiles,
    	u.email AS email,
        u.telegram_id AS telegram_id
    FROM groups g
             LEFT JOIN user_groups ug
                       ON ug.group_id = g.id
                           AND ug.deleted_at IS NULL
             LEFT JOIN users u
                       ON u.id = ug.user_id
                           AND u.deleted_at IS NULL
    WHERE g.deleted_at IS NULL
    `

	if len(group) > 0 {
		query += `AND g.name IN (?) `
	}

	var rows *sql.Rows
	if len(group) > 0 {
		rows, err = s.db.Raw(query, group).Rows()
	} else {
		rows, err = s.db.Raw(query).Rows()
	}

	if err != nil {
		s.logger.Errorw("Failed to execute query", "error", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
		}
	}(rows)

	for rows.Next() {
		var g cache_receptors.GroupWithMobiles
		err := rows.Scan(&g.GroupID, &g.GroupName, &g.UserId, &g.Mobile, &g.Email, &g.TelegramID)
		if err != nil {
			s.logger.Errorw("Failed to scan row", "error", err)
			return nil, err
		}
		gms = append(gms, g)
	}

	return gms, nil
}

func (s *Storage) GetGroupEmails() (string, []string, error) {
	// Implementation for fetching group emails from PostgreSQL database
	return "", nil, nil
}

func (s *Storage) GetUserEmail() (string, []string, error) {
	// Implementation for fetching user emails from PostgreSQL database
	return "", nil, nil
}

func (s *Storage) GetUserNumber() (string, []string, error) {
	// Implementation for fetching user numbers from PostgreSQL database
	return "", nil, nil
}
