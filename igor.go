package main

import (
	"fmt"

	"github.com/ISauve/igor/apt"
	"github.com/ISauve/igor/cos"
	"github.com/ISauve/igor/types"
	"github.com/ISauve/igor/wsl"
)

func GetBackend(target types.Target) (types.Backend, error) {
	var (
		backend types.Backend
		err     error
	)

	switch target.Distro.Display {
	// case "Fedora", "RHEL":
	// 	backend, err = rpm.NewRedHatBackend(&target)
	// case "CentOS":
	// 	backend, err = rpm.NewCentOSBackend(&target)
	// case "openSUSE":
	// 	backend, err = rpm.NewOpenSUSEBackend(&target)
	// case "SLE":
	// 	backend, err = rpm.NewSLESBackend(&target)
	case "Debian", "Ubuntu":
		backend, err = apt.NewBackend(&target)
	case "cos":
		backend, err = cos.NewBackend(&target)
	case "wsl":
		backend, err = wsl.NewBackend(&target)
	default:
		err = fmt.Errorf("Unsupported distribution '%s'", target.Distro.Display)
	}
	return backend, err
}