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
	Enabled              bool   // bật/tắt leader election
	LeaseName            string // tên Lease resource trên K8s
	Namespace            string // namespace của Lease
	PodName              string // identity của pod này (thường = metadata.name)
	LeaseDurationSeconds int    // thời gian giữ lease (default 15s)
	RenewDeadlineSeconds int    // deadline để renew trước khi mất lease (default 10s)
	RetryPeriodSeconds   int    // chu kỳ thử acquire lease (default 2s)
	CSVExportDir         string // thư mục lưu file CSV
	CSVExportHour        int    // giờ chạy export hàng ngày (0-23)
}

var DatabaseConfigInit Config
