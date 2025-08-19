package postgresql

import "github.com/root-ali/iris/pkg/scheduler/cache_receptors"
import _ "github.com/lib/pq"

func (s *Storage) GetGroupNumbers() ([]cache_receptors.GroupWithMobiles, error) {

	var gms []cache_receptors.GroupWithMobiles
	rows, err := s.db.Raw(`
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
    GROUP BY g.id, g.name
    ORDER BY g.name;
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
