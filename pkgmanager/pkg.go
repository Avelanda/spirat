package pkgmanager

import (
	"encoding/json"
	"github.com/package-url/packageurl-go"
	"strings"
)

type PackageID string

type QueryResult struct {
	Packages     map[PackageID]*Package `json:"packages"`
	Dependencies []*PackageDependency   `json:"dependencies"`
}

type Package struct {
	ID           PackageID              `json:"id"`
	Name         string                 `json:"name"`
	Namespace    string                 `json:"namespace"`
	Version      string                 `json:"version"`
	Licenses     []*License             `json:"licenses"`
	LicenseFiles []*LicenseFile         `json:"licenseFiles"`
	HomepageUrl  string                 `json:"homepageUrl"`
	DownloadUrl  string                 `json:"downloadUrl"`
	SourceInfo   string                 `json:"sourceInfo"`
	Filename     string                 `json:"filename"`
	PackageURL   *packageurl.PackageURL `json:"purl"`
}

type PackageDependency struct {
	RequiringPackageID PackageID      `json:"requiringPackageID"`
	RequiredPackageID  PackageID      `json:"requiredPackageID"`
	DependencyType     DependencyType `json:"dependencyType"`
}

type DependencyType string

type packageForEncoding struct {
	Name         string         `json:"name"`
	Namespace    string         `json:"namespace"`
	Version      string         `json:"version"`
	Licenses     []*License     `json:"licenses"`
	LicenseFiles []*LicenseFile `json:"licenseFiles"`
	HomepageUrl  string         `json:"homepageUrl"`
	DownloadUrl  string         `json:"downloadUrl"`
	Filename     string         `json:"filename"`
	PackageURL   string         `json:"purl"`
}

func (p *Package) MarshalJSON() ([]byte, error) {
	pfe := &packageForEncoding{
		Name:         p.Name,
		Namespace:    p.Namespace,
		Version:      p.Version,
		Licenses:     p.Licenses,
		LicenseFiles: p.LicenseFiles,
		HomepageUrl:  p.HomepageUrl,
		DownloadUrl:  p.DownloadUrl,
		Filename:     p.Filename,
		PackageURL:   p.PackageURL.String(),
	}
	return json.Marshal(pfe)
}

type License struct {
	Name string `json:"name"`
}

type LicenseFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type PackageManager interface {
	Query() (*QueryResult, []error)
	String() string
	Available() bool
}

var toolMap map[string]PackageManager = map[string]PackageManager{
	"dpkg": &dpkg{},
	"rpm":  &rpm{},
	"npm":  &npm{},
}

func GetAvailablePackageManagers() []PackageManager {
	var tools []PackageManager

	for _, man := range toolMap {
		if man.Available() {
			tools = append(tools, man)
		}
	}

	return tools
}

func GetPackageManagers(toolNames []string) []PackageManager {
	var tools []PackageManager

	for _, n := range toolNames {
		if man, ok := toolMap[n]; ok && man.Available() {
			tools = append(tools, man)
		}
	}

	return tools
}

func packageID(name, version string) PackageID {
	id := name + "-" + version
	id = strings.ReplaceAll(id, "@", "")
	id = strings.ReplaceAll(id, "/", "-")

	return PackageID(id)
}
