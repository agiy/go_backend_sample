package config

import (
	"github.com/goravel/framework/contracts/database/driver"
	postgresfacades "github.com/goravel/postgres/facades"

	"github.com/agiy/go_backend_sample/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("database", map[string]any{
		// Default database connection name
		"default": config.Env("DB_CONNECTION"),
		// Database connections
		"connections": map[string]any{
			"postgres": map[string]any{
				"host":     config.Env("DB_HOST"),
				"port":     config.Env("DB_PORT"),
				"database": config.Env("DB_DATABASE"),
				"username": config.Env("DB_USERNAME"),
				"password": config.Env("DB_PASSWORD"),
				"sslmode":  config.Env("DB_SSLMODE", "verify-full"),
				"singular": false,
				"prefix":   "",
				"schema":   config.Env("DB_SCHEMA", "public"),
				"via": func() (driver.Driver, error) {
					return postgresfacades.Postgres("postgres")
				},
			},
		},
		// Pool configuration
		"pool": map[string]any{
			// Sets the maximum number of connections in the idle
			// connection pool.
			//
			// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
			// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
			//
			// If n <= 0, no idle connections are retained.
			"max_idle_conns": config.Env("DB_MAX_IDLE_CONNS", 5),
			// Sets the maximum number of open connections to the database.
			//
			// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
			// MaxIdleConns, then MaxIdleConns will be reduced to match the new
			// MaxOpenConns limit.
			//
			// If n <= 0, then there is no limit on the number of open connections.
			"max_open_conns": config.Env("DB_MAX_OPEN_CONNS", 20),
			// Sets the maximum amount of time a connection may be idle.
			//
			// Expired connections may be closed lazily before reuse.
			//
			// If d <= 0, connections are not closed due to a connection's idle time.
			// Unit: Second
			"conn_max_idletime": config.Env("DB_CONN_MAX_IDLETIME", 300),
			// Sets the maximum amount of time a connection may be reused.
			//
			// Expired connections may be closed lazily before reuse.
			//
			// If d <= 0, connections are not closed due to a connection's age.
			// Unit: Second
			"conn_max_lifetime": config.Env("DB_CONN_MAX_LIFETIME", 1800),
		},
		// Sets the threshold for slow queries in milliseconds, the slow query will be logged.
		// Unit: Millisecond
		"slow_threshold": 200,
		// Migration Repository Table
		//
		// This table keeps track of all the migrations that have already run for
		// your application. Using this information, we can determine which of
		// the migrations on disk haven't actually been run in the database.
		"migrations": map[string]any{
			"table": "migrations",
		},
	})
}
