package repository

import (
	"UsersService/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Repository interface {
	RegisterUser(ctx context.Context, user *User) (string, error)
	CheckUserExists(ctx context.Context, username string, email string) (bool, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateLoginTime(ctx context.Context, username string) error
	ShuttingDownPostgres() error
}

func NewRepository(ctx context.Context, cfg config.PostgreSQL) (Repository, error) {
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s 
        pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolSize,
		cfg.PoolConnLifeTime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)

	conf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse db config")
	}

	conf.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create poll")
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ping db")
	}

	return &repository{pool: pool}, nil
}

type repository struct {
	pool *pgxpool.Pool
}

func (r *repository) RegisterUser(ctx context.Context, user *User) (string, error) {
	var username string
	err := r.pool.QueryRow(ctx, createUser, user.Email, user.Username, user.HashPass, user.FirstName, user.LastName).Scan(&username)
	if err != nil {
		return "", errors.Wrap(err, "unable to create user")
	}
	return username, nil
}

func (r *repository) CheckUserExists(ctx context.Context, username string, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, checkExists, username, email).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "unable to check user")
	}
	return exists, nil
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.pool.QueryRow(ctx, getUserByUsername, username).Scan(
		&user.Email,
		&user.Username,
		&user.HashPass,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user by username")
	}
	return &user, nil
}

func (r *repository) UpdateLoginTime(ctx context.Context, username string) error {
	_, err := r.pool.Exec(ctx, updateLastLoginTime, username)
	if err != nil {
		return errors.Wrap(err, "unable to update last login time")
	}
	return nil
}

func (r *repository) ShuttingDownPostgres() error {
	if r.pool != nil {
		r.pool.Close()
		return nil
	}
	return errors.New("postgres pool is empty")
}
