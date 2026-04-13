package db_models

// CliNeMonitor is a DTO derived from CliNe for monitor/command URL purposes.
// The ne_monitor table no longer exists; data is derived from CliNe.CommandURL.
type CliNeMonitor struct {
	NeID      int64  `json:"ne_id"`
	NeName    string `json:"ne_name"`
	NeIP      string `json:"ne_ip"`
	Namespace string `json:"namespace"`
}
