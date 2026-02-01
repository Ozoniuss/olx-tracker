package config

//go:generate go tool genconfig -struct=Config -project=OlxTracker -env=.all
type Config struct {
	Port     string
	Postgres PostgresConfig
}

type PostgresConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Schema   string
}
