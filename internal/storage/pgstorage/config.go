package pgstorage

type PostgresConfig interface {
	GetDSN() string
}

type PostgresParams struct {
	DSN string
}

func (p *PostgresParams) GetDSN() string {
	return p.DSN
}
