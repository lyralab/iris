package postgresql

type Postgres struct {
	Host     string
	User     string
	Password string
	DBname   string
	Port     string
	SSLMode  bool
}
