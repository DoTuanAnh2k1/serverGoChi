package db_models

// CliNeSlave is a legacy DTO kept for interface compatibility.
// The cli_ne_slave table no longer exists in the new schema.
type CliNeSlave struct {
	NeID      int64  `json:"ne_id"`
	IPAddress string `json:"ip_address"`
	Port      int32  `json:"port"`
}
