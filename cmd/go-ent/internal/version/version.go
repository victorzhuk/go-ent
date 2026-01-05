package version

import (
	"fmt"
	"runtime/debug"
)

var (
	version = "dev"
	vcsRef  = "unknown"
)

type Info struct {
	Version   string
	VCSRef    string
	GoVersion string
}

func Get() Info {
	info := Info{
		Version:   version,
		VCSRef:    vcsRef,
		GoVersion: "unknown",
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		info.GoVersion = bi.GoVersion
	}
	return info
}

func String() string {
	v := Get()
	if v.VCSRef == "unknown" || v.VCSRef == "" {
		return v.Version
	}
	return fmt.Sprintf("%s (%s)", v.Version, v.VCSRef)
}
