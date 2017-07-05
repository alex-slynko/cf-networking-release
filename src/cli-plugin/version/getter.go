package version

import (
	"code.cloudfoundry.org/cli/plugin"
)

var CurrentVersion = plugin.VersionType{
	Major: 1,
	Minor: 2,
	Build: 0,
}

type Getter struct{}

func (g *Getter) Get() plugin.VersionType {
	return CurrentVersion
}
