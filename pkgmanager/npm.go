package pkgmanager

import (
	"encoding/json"
	"fmt"
	"github.com/Hitachi/spirat/utils"
	"github.com/package-url/packageurl-go"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type npm struct{}

type npmList struct {
	Dependencies map[string]*dependency `json:"dependencies"`
}

type dependency struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Homepage     string                 `json:"homepage"`
	License      interface{}            `json:"license"`
	Licenses     interface{}            `json:"licenses"`
	Author       interface{}            `json:"author"`
	Repository   interface{}            `json:"repository"`
	Resolved     string                 `json:"resolved"`
	Path         string                 `json:"path"`
	Dependencies map[string]*dependency `json:"dependencies"`
}

type author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type repository struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type license struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

func (d *dependency) packageLicenses() []*License {
	if ls, ok := d.Licenses.([]license); ok {
		return utils.Map(ls, func(l license) *License {
			return &License{l.Type}
		})
	}

	if ls, ok := d.Licenses.([]string); ok {
		return utils.Map(ls, func(l string) *License {
			return &License{l}
		})
	}

	if l, ok := d.License.(license); ok {
		return []*License{{l.Type}}
	}

	if s, ok := d.License.(string); ok {
		return []*License{{s}}
	}

	return []*License{}
}

func (n *npm) Query() (*QueryResult, []error) {
	cmd := exec.Command("npm", "list", "--json", "--all", "--long", "--package-lock-only", "--unicode")
	output, err := cmd.Output()
	if err != nil {
		return nil, []error{fmt.Errorf("%s: %w", string(output), err)}
	}

	var list npmList
	err = json.Unmarshal(output, &list)
	if err != nil {
		return nil, []error{err}
	}

	queryResult := &QueryResult{Packages: make(map[PackageID]*Package)}
	n.fillInformationUsingPackageJson(list.Dependencies)
	for _, dep := range list.Dependencies {
		n.addPackage(queryResult, dep)
	}

	return queryResult, nil
}

func (n *npm) String() string {
	return "npm"
}

func (n *npm) Available() bool {
	if _, err := os.Stat("package-lock.json"); os.IsNotExist(err) {
		if _, err := os.Stat("npm-shrinkwrap.json"); os.IsNotExist(err) {
			return false
		}
	}

	return hasCommand("npm")
}

func (n *npm) addPackage(queryResult *QueryResult, dep *dependency) *Package {
	namespace, name := n.splitToNamespaceAndName(dep.Name)

	if queryResult.Packages[packageID(dep.Name, dep.Version)] == nil {
		queryResult.Packages[packageID(dep.Name, dep.Version)] = &Package{
			ID:           packageID(dep.Name, dep.Version),
			Name:         name,
			Namespace:    namespace,
			Version:      dep.Version,
			Licenses:     dep.packageLicenses(),
			LicenseFiles: []*LicenseFile{},
			HomepageUrl:  dep.Homepage,
			DownloadUrl:  dep.Resolved,
			Filename:     fmt.Sprintf("%s-%s.tgz", dep.Name, dep.Version),
			PackageURL: packageurl.NewPackageURL(
				packageurl.TypeNPM,
				namespace,
				name,
				dep.Version,
				packageurl.Qualifiers{},
				"",
			),
		}
	}

	pkg := queryResult.Packages[packageID(dep.Name, dep.Version)]

	for _, dd := range dep.Dependencies {
		queryResult.Dependencies = append(queryResult.Dependencies, &PackageDependency{
			RequiringPackageID: packageID(dep.Name, dep.Version),
			RequiredPackageID:  packageID(dd.Name, dd.Version),
			DependencyType:     "DEPENDS_ON",
		})

		n.addPackage(queryResult, dd)
	}

	return pkg
}

func (n *npm) splitToNamespaceAndName(namespaceAndName string) (string, string) {
	namespace, name, ok := strings.Cut(namespaceAndName, "/")
	if ok {
		return namespace, name
	} else {
		return "", namespaceAndName
	}
}

func (n *npm) fillInformationUsingPackageJson(deps map[string]*dependency) {
	for _, dep := range deps {
		p := path.Join(dep.Path, "package.json")
		pj, err := n.parsePackageJson(p)
		if err != nil {
			return
		}

		dep.License = pj.License
		dep.Licenses = pj.Licenses
		dep.Description = pj.Description
		dep.Repository = pj.Repository

		n.fillInformationUsingPackageJson(dep.Dependencies)
	}
}

type packageJson struct {
	License     interface{} `json:"license"`
	Licenses    interface{} `json:"licenses"`
	Description string      `json:"description"`
	Repository  interface{} `json:"repository"`
}

func (n *npm) parsePackageJson(path string) (*packageJson, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pj := &packageJson{}
	err = json.Unmarshal(bytes, pj)
	if err != nil {
		return nil, err
	}

	return pj, nil
}
