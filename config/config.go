package config

import "serverGoChi/models/config_models"

var config *config_models.Config

func Init(cfg *config_models.Config) {
	config = cfg
}

func Get() *config_models.Config {
    return config
}

func GetServerConfig() config_models.ServerConfig {
	return config.Svr
}

func GetDatabaseConfig() config_models.DatabaseConfig {
    return config.Db
}

func GetLogConfig() config_models.LogConfig{
	return config.Log
} 

func GetJwtConfig() config_models.TokenConfig{
    return config.Token
}

func GetRouterConfig() config_models.RouterConfig{
	return config.Router
}