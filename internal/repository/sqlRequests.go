package repository

const (
	createUser          = `INSERT INTO users (email, username, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4, $5) returning username`
	checkExists         = `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1 or email = $2)`
	getUserByUsername   = `SELECT email, username, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE username = $1`
	updateLastLoginTime = `UPDATE users SET last_login_at = NOW() WHERE username = $1;`
)
