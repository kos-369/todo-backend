package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

// dbConfigFromEnv builds the MySQL config from environment variables,
// so connection details and secrets stay out of the code.
func dbConfigFromEnv() (*mysql.Config, error) {
	var errs []error
	get := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			errs = append(errs, fmt.Errorf("%s is not set", key))
		}
		return v
	}

	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = get("DB_ADDR")
	cfg.User = get("DB_USER")
	cfg.Passwd = get("DB_PASSWORD")
	cfg.DBName = get("DB_NAME")

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return cfg, nil
}
