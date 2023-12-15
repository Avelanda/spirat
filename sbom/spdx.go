package sbom

import (
	"fmt"
	"github.com/Hitachi/spirat/pkgmanager"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/spdx/tools-golang/spdx"
	"strings"
	"time"
)

const (
	ElementPackage = "Package"
	NOASSERTION    = "NOASSERTION"
)

func ToSpdx(qrs []*pkgmanager.QueryResult) *spdx.Document {
	var doc spdx.Document
	doc.SPDXVersion = spdx.Version
	doc.DataLicense = spdx.DataLicense
	doc.DocumentName = "spirat-generated-document"
	doc.SPDXIdentifier = "SPDXRef-DOCUMENT"
	doc.CreationInfo = &spdx.CreationInfo{
		Created:  time.Now().Format(time.RFC3339),
		Creators: []spdx.Creator{{Creator: "spirat", CreatorType: "Tool"}},
	}

	for _, r := range qrs {
		for _, pkg := range r.Packages {
			spdxPkg, _ := toSpdxPackage(pkg)
			doc.Packages = append(doc.Packages, spdxPkg)
			doc.Relationships = append(doc.Relationships, &spdx.Relationship{
				RefA:         spdx.DocElementID{ElementRefID: doc.SPDXIdentifier},
				RefB:         spdx.DocElementID{ElementRefID: spdxPkg.PackageSPDXIdentifier},
				Relationship: spdx.RelationshipDescribes,
			})

			if len(pkg.Licenses) == 0 {
				for _, file := range pkg.LicenseFiles {
					h, _ := hashstructure.Hash(file.Path, hashstructure.FormatV2, nil)
					licenseRef := fmt.Sprintf("LicenseRef-%x", h)
					spdxPkg.PackageLicenseDeclared = licenseRef
					doc.OtherLicenses = append(doc.OtherLicenses, &spdx.OtherLicense{
						LicenseIdentifier: licenseRef,
						ExtractedText:     file.Content,
					})
				}
			}
		}

		for _, dep := range r.Dependencies {
			doc.Relationships = append(doc.Relationships, &spdx.Relationship{
				RefA:         spdx.DocElementID{ElementRefID: packageId(dep.RequiringPackageID)},
				RefB:         spdx.DocElementID{ElementRefID: packageId(dep.RequiredPackageID)},
				Relationship: spdx.RelationshipDependsOn,
			})
		}
	}

	return &doc
}

func toSpdxPackage(p *pkgmanager.Package) (*spdx.Package, error) {
	var spdxPkg spdx.Package
	spdxPkg.PackageSPDXIdentifier = packageId(p.ID)
	spdxPkg.PackageName = p.Name
	spdxPkg.PackageVersion = p.Version
	spdxPkg.PackageHomePage = p.HomepageUrl
	spdxPkg.PackageDownloadLocation = p.DownloadUrl
	spdxPkg.PackageSourceInfo = p.SourceInfo
	spdxPkg.PackageLicenseDeclared = spdxLicense(p.Licenses)
	spdxPkg.PackageExternalReferences = []*spdx.PackageExternalReference{
		{
			Category: spdx.CategoryPackageManager,
			RefType:  spdx.PackageManagerPURL,
			Locator:  p.PackageURL.String(),
		},
	}

	return &spdxPkg, nil
}

func packageId(id pkgmanager.PackageID) spdx.ElementID {
	return spdx.ElementID(ElementPackage + "-" + id)
}

func spdxLicense(licenses []*pkgmanager.License) string {
	if len(licenses) == 0 {
		return NOASSERTION
	}

	var licenseNames []string

	for _, license := range licenses {
		licenseNames = append(licenseNames, license.Name)
	}

	return strings.Join(licenseNames, " and ")
}
