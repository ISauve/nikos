// +build dnf

package rpm

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/ISauve/nikos/types"
)

type CentOSBackend struct {
	version    int
	release    string
	dnfBackend *DnfBackend
	target     *types.Target
}

func getRedhatRelease() (string, error) {
	redhatRelease, err := ioutil.ReadFile("/etc/redhat-release")
	if err != nil {
		return "", errors.Wrap(err, "failed to read /etc/redhat-release")
	}

	re := regexp.MustCompile(`.* release ([0-9\.]*)`)
	submatches := re.FindStringSubmatch(string(redhatRelease))
	if len(submatches) == 2 {
		return submatches[1], nil
	}

	return "", errors.New("failed to parse release from /etc/redhat-release")
}

func (b *CentOSBackend) GetKernelHeaders(directory string) error {
	pkgNevra := "kernel-headers-" + b.target.Uname.Kernel

	// First try with the 'base' and 'updates' repositories.
	// This should work if the user is using the latest minor version
	log.Infof("Trying with 'base' and 'updates' repositories")

	var disabledRepositories []*Repository
	for _, repo := range b.dnfBackend.GetEnabledRepositories() {
		if repo.id != "base" && repo.id != "updates" {
			b.dnfBackend.DisableRepository(repo)
		}
		disabledRepositories = append(disabledRepositories, repo)
	}

	if b.dnfBackend.GetKernelHeaders(pkgNevra, directory) == nil {
		return nil
	}

	// Otherwise, we try with Vault
	log.Infof("Trying with Vault repositories for %s", b.release)

	var baseURL string
	gpgKey := "file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-"
	if b.version >= 8 {
		gpgKey += "centosofficial"
		baseURL = fmt.Sprintf("http://vault.centos.org/%s/BaseOS/$basearch/os/", b.release)
	} else {
		gpgKey += strconv.Itoa(b.version)
		baseURL = fmt.Sprintf("http://vault.centos.org/%s/os/$basearch/", b.release)

		updatesURL := fmt.Sprintf("http://vault.centos.org/%s/updates/$basearch/", b.release)
		if _, err := b.dnfBackend.AddRepository("C"+b.release+"-updates", updatesURL, true, gpgKey); err != nil {
			return errors.Wrap(err, "failed to add Vault updates repository")
		}
	}

	if _, err := b.dnfBackend.AddRepository("C"+b.release+"-base", baseURL, true, gpgKey); err != nil {
		return errors.Wrap(err, "failed to add Vault base repository")
	}

	return b.dnfBackend.GetKernelHeaders(pkgNevra, directory)
}

func (b *CentOSBackend) Close() {
	b.dnfBackend.Close()
}

func NewCentOSBackend(target *types.Target, reposDir string) (*CentOSBackend, error) {
	release, err := getRedhatRelease()
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect CentOS release")
	}

	version, _ := strconv.Atoi(strings.SplitN(release, ".", 2)[0])
	dnfBackend, err := NewDnfBackend(fmt.Sprintf("%d", version), reposDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DNF backend")
	}

	return &CentOSBackend{
		version:    version,
		target:     target,
		release:    release,
		dnfBackend: dnfBackend,
	}, nil
}
