package config

import (
	"errors"
	"github.com/btcsuite/btcd/btcutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// PortKey is port on which server is running
	PortKey = "PORT_PORT"
	// LogLevelKey is log level used by dnsd
	LogLevelKey = "LOG_LEVEL"
	// DatadirKey is the local data directory to store the internal state of daemon
	DatadirKey = "DATADIR"
	// DbUserKey is the user name to connect to the database
	DbUserKey = "DB_USER"
	// DbPassKey is the password to connect to the database
	DbPassKey = "DB_PASS"
	// DbHostKey is the host address of the database
	DbHostKey = "DB_HOST"
	// DbPortKey is the port of the database
	DbPortKey = "DB_PORT"
	// DbNameKey is the name of the database
	DbNameKey = "DB_NAME"
	// DbMigrationPathKey is the path to the database migration files
	DbMigrationPathKey = "DB_MIGRATION_PATH"
	// RootApiKey is the root API key with unrestricted access
	RootApiKey = "ROOT_API_KEY"
	// AdminUserKey is the admin username
	AdminUserKey = "ADMIN_USER"
	// AdminPassKey is the admin password
	AdminPassKey = "ADMIN_PASS"
)

var (
	vip *viper.Viper
)

func LoadConfig() error {
	vip = viper.New()
	vip.SetEnvPrefix("PREM_GATEWAY_AUTH")
	vip.AutomaticEnv()
	defaultDataDir := btcutil.AppDataDir("authd", false)

	vip.SetDefault(PortKey, 8080)
	vip.SetDefault(LogLevelKey, int(log.DebugLevel))
	vip.SetDefault(DatadirKey, defaultDataDir)
	vip.SetDefault(DbUserKey, "root")
	vip.SetDefault(DbPassKey, "secret")
	vip.SetDefault(DbHostKey, "127.0.0.1")
	vip.SetDefault(DbPortKey, 5432)
	vip.SetDefault(DbNameKey, "authd-db")
	vip.SetDefault(DbMigrationPathKey, "file://auth/internal/infrastructure/storage/pg/migration")

	if vip.GetString(RootApiKey) == "" {
		return errors.New("root API key not set")
	}

	if vip.GetString(AdminUserKey) == "" {
		return errors.New("admin user not set")
	}

	if vip.GetString(AdminPassKey) == "" {
		return errors.New("admin password not set")
	}

	return nil
}

func GetString(key string) string {
	return vip.GetString(key)
}

func GetInt(key string) int {
	return vip.GetInt(key)
}

func GetServerAddress() string {
	return ":" + GetString(PortKey)
}
