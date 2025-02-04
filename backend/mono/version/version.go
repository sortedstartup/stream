package version

import (
	_ "embed"
)

//go:embed version.number
var version_number string

//go:embed version.githash
var version_githash string

var Version VersionInfo

type VersionInfo struct {
	Version string `json:"version"`
	GitHash string `json:"gitHash"`
}

func init() {
	Version = VersionInfo{
		Version: version_number,
		GitHash: version_githash,
	}
}
