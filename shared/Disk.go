package shared

type ServerStatus struct {
	Disks []Disk `json:"disks"`
}
type Disk struct {
	Filesystem string `json:"filesystem"`
	Used       uint64 `json:"used"`
	Available  uint64 `json:"available"`
	MountPoint string `json:"mountPoint"`
}
