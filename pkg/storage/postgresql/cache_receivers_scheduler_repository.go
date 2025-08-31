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
        COALESCE(
                array_remove(array_agg(DISTINCT u.mobile), NULL),
                ARRAY[]::varchar[]
        ) AS mobiles
    FROM groups g
             LEFT JOIN user_groups ug
                       ON ug.group_id = g.id
                           AND ug.deleted_at IS NULL
             LEFT JOIN users u
                       ON u.id = ug.user_id
                           AND u.deleted_at IS NULL
    WHERE g.deleted_at IS NULL
    `

	// If group is not empty, add the condition for group names.
	if len(group) > 0 {
		query += `AND g.name IN (?) `
	}

	query += `GROUP BY g.id, g.name
			ORDER BY g.name;`

	// If group is provided, pass it as an argument to the query; otherwise, don't pass any argument.
	var rows *sql.Rows
	if len(group) > 0 {
		rows, err = s.db.Raw(query, group).Rows()
	} else {
		rows, err = s.db.Raw(query).Rows()
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Read rows and scan data.
	for rows.Next() {
		var g cache_receptors.GroupWithMobiles
		err := rows.Scan(&g.GroupID, &g.GroupName, &g.Mobiles)
		if err != nil {
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
