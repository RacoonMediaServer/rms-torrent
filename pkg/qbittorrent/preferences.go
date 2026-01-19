package qbittorrent

type Preferences struct {
	AutorunEnabled     bool   `json:"autorun_enabled"`
	AutorunProgram     string `json:"autorun_program"`
	AutoTmmEnabled     bool   `json:"auto_tmm_enabled"`
	QueueMode          bool   `json:"queueing_enabled"`
	MaxActiveDownloads int32  `json:"max_active_downloads"`
}
