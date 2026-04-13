package db_models

// CliNeConfig is a DTO representing the connection config for a NE in ne-config mode.
// In the new schema these fields are stored directly on CliNe (conf_* columns),
// so this struct has no backing table of its own.
type CliNeConfig struct {
	ID          int64  `json:"id"`
	NeID        int64  `json:"ne_id"`
	IPAddress   string `json:"ip_address"`
	Port        int32  `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Protocol    string `json:"protocol"`
	Description string `json:"description"`
}
