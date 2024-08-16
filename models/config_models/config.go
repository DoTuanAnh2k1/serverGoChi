package config_models

type Config struct {
	Db DatabaseConfig
}

type DatabaseConfig struct {
	DbType string
	Mysql  MySqlConfig
}

var DatabaseConfigInit Config
