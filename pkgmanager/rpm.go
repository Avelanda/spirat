package pkgmanager

import (
	"fmt"
	"github.com/Hitachi/spirat/sysinfo"
	"github.com/Hitachi/spirat/utils"
	"github.com/package-url/packageurl-go"
	"os"
	"os/exec"
	"strings"
)

type rpm struct{}

func (r *rpm) Query() (*QueryResult, []error) {
	names, err := r.queryNames()
	if err != nil {
		return nil, []error{err}
	}

	versions, err := r.queryVersions()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(versions) {
		return nil, []error{fmt.Errorf("names and versions should be the same length")}
	}

	releases, err := r.queryReleases()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(releases) {
		return nil, []error{fmt.Errorf("names and releases should be the same length")}
	}

	architectures, err := r.queryArchitectures()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(architectures) {
		return nil, []error{fmt.Errorf("names and architectures should be the same length")}
	}

	licenses, err := r.queryLicenses()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(licenses) {
		return nil, []error{fmt.Errorf("names and licenses should be the same length")}
	}

	urls, err := r.queryUrls()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(urls) {
		return nil, []error{fmt.Errorf("names and urls should be the same length")}
	}

	osRelease := sysinfo.NewOSRelease()
	pkgs := make(map[PackageID]*Package)
	var errs []error
	for i := 0; i < len(names); i++ {
		name := names[i]
		version := versions[i]
		release := releases[i]
		arch := architectures[i]
		licenseFiles, licenseErrs := r.queryLicenseFiles(r.constructRpmName(name, version, release, arch))
		if licenseErrs != nil {
			errs = append(errs, licenseErrs...)
		}

		pkg := &Package{
			ID:           packageID(name, version),
			Name:         name,
			Version:      version,
			Licenses:     []*License{licenses[i]},
			LicenseFiles: licenseFiles,
			HomepageUrl:  urls[i],
			Filename:     r.constructFilename(name, version, release, arch),
			PackageURL: packageurl.NewPackageURL(
				packageurl.TypeRPM,
				osRelease.ID,
				name,
				version,
				packageurl.Qualifiers{},
				"",
			),
		}
		pkgs[packageID(name, version)] = pkg
	}

	queryResult := &QueryResult{Packages: pkgs}

	if len(errs) > 0 {
		return queryResult, errs
	}

	return queryResult, nil
}

func (r *rpm) String() string {
	return "rpm"
}

func (r *rpm) Available() bool {
	return hasCommand("rpm")
}

func (r *rpm) queryNames() ([]string, error) {
	cmd := exec.Command("rpm", "-q", "--all", "--qf", "%{NAME}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

func (r *rpm) queryVersions() ([]string, error) {
	cmd := exec.Command("rpm", "-q", "--all", "--qf", "%{VERSION}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

func (r *rpm) queryReleases() ([]string, error) {
	cmd := exec.Command("rpm", "-q", "--all", "--qf", "%{RELEASE}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

func (r *rpm) queryArchitectures() ([]string, error) {
	cmd := exec.Command("rpm", "-q", "--all", "--qf", "%{ARCH}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

func (r *rpm) queryUrls() ([]string, error) {
	cmd := exec.Command("rpm", "-q", "--all", "--qf", "%{URL}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

func (r *rpm) queryLicenses() ([]*License, error) {
	cmd := exec.Command("rpm", "-q", "--all", "--qf", "%{LICENSE}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	names := strings.Split(strings.TrimSpace(string(output)), "\n")

	return utils.Map(names, func(name string) *License {
		return &License{Name: name}
	}), nil
}

func (r *rpm) queryLicenseFiles(rpmName string) ([]*LicenseFile, []error) {
	var errs []error

	licensePaths, err := r.queryLicensePaths(rpmName)
	if err != nil {
		errs = append(errs, err)
	}

	var licenseFiles []*LicenseFile
	for _, licensePath := range licensePaths {
		licenseText, err := r.readLicenseText(licensePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		licenseFiles = append(licenseFiles, &LicenseFile{
			Path:    licensePath,
			Content: licenseText,
		})
	}

	if len(errs) > 0 {
		return licenseFiles, errs
	}

	return licenseFiles, nil
}

func (r *rpm) queryLicensePaths(name string) ([]string, error) {
	cmd := exec.Command("rpm", "-q", name, "-L")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

func (r *rpm) constructRpmName(name, version, release, arch string) string {
	return fmt.Sprintf("%s-%s-%s.%s", name, version, release, arch)
}

func (r *rpm) constructFilename(name, version, release, arch string) string {
	return fmt.Sprintf("%s.rpm", r.constructRpmName(name, version, release, arch))
}

func (r *rpm) readLicenseText(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return no error if the file does not exist.
		return "", nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
