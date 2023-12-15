package pkgmanager

import (
	"fmt"
	"github.com/Hitachi/spirat/sysinfo"
	"github.com/package-url/packageurl-go"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	formatRe  = regexp.MustCompile(`Format:(.+)`)
	licenseRe = regexp.MustCompile(`License:(.+)`)
)

type dpkg struct{}

func (d *dpkg) Query() (*QueryResult, []error) {
	names, err := d.queryNames()
	if err != nil {
		return nil, []error{err}
	}

	versions, err := d.queryVersions()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(versions) {
		return nil, []error{fmt.Errorf("names and versions should be the same length")}
	}

	architectures, err := d.queryArchitectures()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(architectures) {
		return nil, []error{fmt.Errorf("names and versions should be the same length")}
	}

	homepages, err := d.queryHomepages()
	if err != nil {
		return nil, []error{err}
	}

	if len(names) != len(homepages) {
		return nil, []error{fmt.Errorf("names and homepages should be the same length")}
	}

	osRelease := sysinfo.NewOSRelease()
	aptSourcesMap := d.queryAptSources(names)

	pkgs := make(map[PackageID]*Package)
	var errs []error
	for i := 0; i < len(names); i++ {
		name := names[i]
		version := versions[i]
		arch := architectures[i]

		licenses, err := d.queryLicense(name)
		if err != nil {
			err := fmt.Errorf("failed to find licenses in %s: %v", name, err)
			errs = append(errs, err)
		}

		licenseFiles, err := d.queryLicenseFiles(name)
		if err != nil {
			err := fmt.Errorf("failed to find licenses in %s: %v", name, err)
			errs = append(errs, err)
		}

		pkg := &Package{
			ID:           packageID(name, version),
			Name:         name,
			Version:      version,
			Licenses:     licenses,
			HomepageUrl:  homepages[i],
			SourceInfo:   aptSourcesMap[name],
			Filename:     d.constructFilename(name, version, arch),
			LicenseFiles: licenseFiles,
			PackageURL: packageurl.NewPackageURL(
				packageurl.TypeDebian,
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

func (d *dpkg) String() string {
	return "dpkg"
}

func (d *dpkg) Available() bool {
	return hasCommand("dpkg-query")
}

func (d *dpkg) extractInstalled(lines []string) []string {
	var ret []string

	for _, line := range lines {
		split := strings.SplitN(line, " ", 2)
		status := split[0]
		name := strings.TrimLeft(split[1], " ")
		if status == "ii" {
			ret = append(ret, name)
		}
	}

	return ret
}

func (d *dpkg) queryNames() ([]string, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${db:Status-Abbrev} ${Package}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return d.extractInstalled(strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")), nil
}

func (d *dpkg) queryVersions() ([]string, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${db:Status-Abbrev} ${Version}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	versions := d.extractInstalled(strings.Split(strings.TrimSuffix(string(output), "\n"), "\n"))
	for i, version := range versions {
		colon := strings.Index(version, ":")
		if colon != -1 {
			versions[i] = version[colon+1:]
		}
	}

	return versions, nil
}

func (d *dpkg) queryArchitectures() ([]string, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${db:Status-Abbrev} ${Architecture}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return d.extractInstalled(strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")), nil
}

func (d *dpkg) queryHomepages() ([]string, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${db:Status-Abbrev} ${Homepage}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return d.extractInstalled(strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")), nil
}

func (d *dpkg) queryAptSources(packageNames []string) map[string]string {
	if !hasCommand("apt") {
		return map[string]string{}
	}

	// To minimize the number of command invocations, we merge package names whenever possible.
	var argsList [][]string
	argN := 0
	argI := 0
	for _, n := range packageNames {
		if argI == len(argsList) {
			argsList = append(argsList, []string{"info"})
		}
		argsList[argI] = append(argsList[argI], n)
		argN += len(n)

		// The maximum length of arguments would be `ARG_MAX - len("apt info ") - MAXIMUM_LEN_OF_PACKAGE_NAME`, where
		// ARG_MAX is typically 4KB or larger on POSIX-compliant operating systems. However, the specification does not
		// seem to mention a specific value for MAXIMUM_LEN_OF_PACKAGE_NAME.
		// As a result, we assume a value of 255 for this parameter.
		if argN > 4096-len("apt info ")-255 {
			argN = 0
			argI++
		}
	}

	ret := make(map[string]string)
	for _, args := range argsList {
		cmd := exec.Command("apt", args...)
		output, err := cmd.Output()
		if err != nil {
			return map[string]string{}
		}

		name := ""
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Package:") {
				pair := strings.SplitN(line, ":", 2)
				name = strings.TrimSpace(pair[1])
			}

			if strings.HasPrefix(line, "APT-Sources:") && name != "" {
				pair := strings.SplitN(line, ":", 2)
				ret[name] = strings.TrimSpace(pair[1])
			}
		}
	}

	return ret
}

func (d *dpkg) queryLicense(name string) ([]*License, error) {
	path := d.findCopyrightPath(name)
	if path == "" {
		return nil, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	format, err := d.detectFormat(string(bytes))
	if err != nil {
		return nil, err
	}

	if format != "" {
		return d.findLicensesOfMachineReadableCopyright(string(bytes))
	}

	return nil, nil
}

// detectFormat returns the format of the specified text. If the detected format is not machine-readable, this method
// will return an empty string without an error. Errors will only be returned if something goes wrong. Additional
// information about the format can be found on the following website: https://dep-team.pages.debian.net/deps/dep5/
func (d *dpkg) detectFormat(content string) (string, error) {
	match := formatRe.FindStringSubmatch(content)
	if match == nil {
		// Machine-readable files must have a line beginning with `Format: `.
		return "", nil
	}

	if len(match) != 2 {
		return "", fmt.Errorf("unexpected match size %d", len(match))
	}

	return strings.Trim(match[1], " "), nil
}

func (d *dpkg) findCopyrightPath(name string) string {
	path := fmt.Sprintf("/usr/share/doc/%s/copyright", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}

	colonPos := strings.LastIndex(name, ":")
	if colonPos == -1 {
		return ""
	}

	shortName := name[0:colonPos]
	path = fmt.Sprintf("/usr/share/doc/%s/copyright", shortName)
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

func (d *dpkg) findLicensesOfMachineReadableCopyright(content string) ([]*License, error) {
	matches := licenseRe.FindAllStringSubmatch(content, -1)
	if matches == nil {
		return nil, nil
	}

	set := make(map[string]struct{})
	for _, match := range matches {
		licenseStr := strings.TrimSpace(match[1])

		if strings.Contains(licenseStr, " or ") {
			set["("+licenseStr+")"] = struct{}{}
			continue
		}

		ls := strings.Split(licenseStr, " and ")
		for _, l := range ls {
			name := strings.TrimSpace(l)
			set[name] = struct{}{}
		}
	}

	var licenses []*License
	for name := range set {
		license := License{
			Name: name,
		}
		licenses = append(licenses, &license)
	}

	return licenses, nil
}

func (d *dpkg) constructFilename(name, version, arch string) string {
	return fmt.Sprintf("%s_%s_%s.deb", name, version, arch)
}

func (d *dpkg) queryLicenseFiles(name string) ([]*LicenseFile, error) {
	path := d.findCopyrightPath(name)
	if path == "" {
		return []*LicenseFile{}, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return []*LicenseFile{{
		Path:    path,
		Content: string(bytes),
	}}, nil
}
