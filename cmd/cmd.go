package cmd

import (
	"fmt"
	"os"

	"github.com/cobaugh/osrelease"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ISauve/nikos/apt"
	"github.com/ISauve/nikos/cos"
	"github.com/ISauve/nikos/rpm"
	"github.com/ISauve/nikos/types"
	"github.com/ISauve/nikos/wsl"
)

var (
	osReleaseFile string
	target        types.Target
	outputDir     string
	verbose       bool
	aptConfigDir  string
	rpmReposDir   string
)

var RootCmd = &cobra.Command{
	Use:          "nikos [sub]",
	SilenceUsage: true,
	PersistentPreRun: func(c *cobra.Command, args []string) {
		if osReleaseFile != "" {
			var err error
			if target.OSRelease, err = osrelease.ReadFile(osReleaseFile); err != nil {
				log.Fatalf("failed to read %s", osReleaseFile)
			}
		}

		if verbose {
			log.SetLevel(log.DebugLevel)
			log.Debugf("Set log level to debug")
		}
	},
}

var DownloadCmd = &cobra.Command{
	Use: "download package",
	Run: func(c *cobra.Command, args []string) {
		log.Infof("Distribution: %s\n", target.Distro.Display)
		log.Infof("Release: %s\n", target.Distro.Release)
		log.Infof("Kernel: %s\n", target.Uname.Kernel)
		log.Debugf("OSRelease: %s\n", target.OSRelease)

		var (
			backend types.Backend
			err     error
		)

		switch target.Distro.Display {
		case "Fedora", "RHEL":
			backend, err = rpm.NewRedHatBackend(&target, rpmReposDir)
		case "CentOS":
			backend, err = rpm.NewCentOSBackend(&target, rpmReposDir)
		case "openSUSE":
			backend, err = rpm.NewOpenSUSEBackend(&target, rpmReposDir)
		case "SLE":
			backend, err = rpm.NewSLESBackend(&target, rpmReposDir)
		case "Debian", "Ubuntu":
			backend, err = apt.NewBackend(&target, aptConfigDir)
		case "cos":
			backend, err = cos.NewBackend(&target)
		case "wsl":
			backend, err = wsl.NewBackend(&target)
		default:
			err = fmt.Errorf("Unsupported distribution '%s'", target.Distro.Display)
		}
		if err != nil {
			log.Fatal(err)
		}

		if err = os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatal(err)
		}

		if err = backend.GetKernelHeaders(outputDir); err != nil {
			log.Fatalf("failed to download kernel headers: %s", err)
		}
	},
}

func SetupCommands() error {
	var err error
	target, err = types.NewTarget()
	if err != nil {
		return fmt.Errorf("failed to retrieve target information: %s", err)
	}

	if _, err := os.Stat("/run/WSL"); err == nil {
		target.Distro.Display = "wsl"
	} else if id := target.OSRelease["ID"]; target.Distro.Display == "" && id != "" {
		target.Distro.Display = id
	}

	RootCmd.PersistentFlags().StringVarP(&osReleaseFile, "os-release", "", "", "path to os-release file")
	RootCmd.PersistentFlags().StringVarP(&target.Distro.Display, "distribution", "d", target.Distro.Display, "distribution name")
	RootCmd.PersistentFlags().StringVarP(&target.Distro.Release, "release", "r", target.Distro.Release, "distribution release")
	RootCmd.PersistentFlags().StringVarP(&target.Uname.Kernel, "kernel", "k", target.Uname.Kernel, "kernel version")
	RootCmd.PersistentFlags().StringVarP(&target.Uname.Machine, "arch", "a", target.Uname.Machine, "architecture")
	RootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "/tmp", "output directory")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")

	switch target.Distro.Display {
	case "Debian", "Ubuntu":
		RootCmd.PersistentFlags().StringVarP(&aptConfigDir, "apt-config-dir", "", "/etc/apt", "APT configuration dir")
	case "Fedora", "RHEL", "CentOS":
		RootCmd.PersistentFlags().StringVarP(&rpmReposDir, "yum-repos-dir", "", "/etc/yum.repos.d", "YUM configuration dir")
	case "openSUSE", "SLE":
		RootCmd.PersistentFlags().StringVarP(&rpmReposDir, "yum-repos-dir", "", "/etc/zypp/repos.d", "YUM configuration dir")
	default:
	}

	RootCmd.AddCommand(DownloadCmd)
	return nil
}
