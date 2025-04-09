package config

import "time"

const EnvPath = "local.env"

type Config struct {
	LogLevel   string `envconfig:"logLevel" default:"info"`
	PostgreSQL PostgreSQL
	Grpc       GRPC
	Jwt        JWT
}

type PostgreSQL struct {
	Host                string        `envconfig:"DB_HOST" required:"true"`
	Port                int           `envconfig:"DB_PORT" required:"true"`
	Name                string        `envconfig:"DB_NAME" required:"true"`
	User                string        `envconfig:"DB_USER" required:"true"`
	Password            string        `envconfig:"DB_PASSWORD" required:"true"`
	SSLMode             string        `envconfig:"DB_SSL_MODE" default:"disable"`
	PoolSize            int           `envconfig:"DB_POOL_MAX_CONNS" default:"10"`
	PoolConnLifeTime    time.Duration `envconfig:"DB_POOL_MAX_CONN_LIFETIME" default:"180s"`
	PoolMaxConnIdleTime time.Duration `envconfig:"DB_POOL_MAX_CONN_IDLE_TIME" default:"100s"`
}

type GRPC struct {
	Port    string        `envconfig:"GRPC_PORT" required:"true"`
	Timeout time.Duration `envconfig:"GRPC_TIMEOUT" required:"true"`
}

type JWT struct {
	ExpireTime  time.Duration `envconfig:"JWT_EXPIRE_TIME" required:"true"`
	RefreshTime time.Duration `envconfig:"JWT_REFRESH_TIME" required:"true"`
	JwtSecret   string        `envconfig:"JWT_SECRET" required:"true"`
}
