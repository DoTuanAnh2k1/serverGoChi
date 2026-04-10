package config_models

type Config struct {
	Db     DatabaseConfig
	Svr    ServerConfig
	Log    LogConfig
	Token  TokenConfig
	Router RouterConfig
	Leader LeaderConfig
}

type ServerConfig struct {
	ServerName string
	Host       string
	Port       string
}

type RouterConfig struct {
	BasePath string
	Origins  string
	Methods  string
	Headers  string
}

type LogConfig struct {
	Level   string
	DbLevel string
}

type DatabaseConfig struct {
	DbType   string
	Mysql    MySqlConfig
	Mongo    MongoConfig
	Postgres PostgresConfig
}

type TokenConfig struct {
	SecretKey   string
	ExpiryHours int
}

type LeaderConfig struct {
	Enabled              bool   // enable/disable leader election
	LeaseName            string // Lease resource name in K8s
	Namespace            string // Lease namespace
	PodName              string // pod identity (usually metadata.name)
	LeaseDurationSeconds int    // lease hold duration (default 15s)
	RenewDeadlineSeconds int    // renew deadline before losing lease (default 10s)
	RetryPeriodSeconds   int    // lease acquire retry period (default 2s)
	CSVExportDir         string // CSV export directory
	CSVExportHour        int    // daily export hour (0-23)
}

var DatabaseConfigInit Config
