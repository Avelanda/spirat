package reporter

import (
	"encoding/json"
	"github.com/Hitachi/spirat/pkgmanager"
	"github.com/Hitachi/spirat/sbom"
	"github.com/Hitachi/spirat/spirat"
	"github.com/Hitachi/spirat/utils"
	"github.com/spdx/tools-golang/spdx"
)

type Diff struct {
	Spirat *spirat.Spirat
	Base   *spdx.Document
}

func (d *Diff) Report() (string, error) {
	qrs := utils.Map(d.Spirat.Results, func(r *spirat.Result) *pkgmanager.QueryResult {
		return r.QueryResult
	})
	doc := sbom.ToSpdx(qrs)
	d.diff(doc)
	jsonStr, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	return string(jsonStr), nil
}

func (d *Diff) diff(doc *spdx.Document) {
	basePackages := d.createBasePackages()

	var packages []*spdx.Package
	packageIds := make(map[spdx.ElementID]struct{})
	otherLicenseIds := make(map[string]struct{})
	for _, p := range doc.Packages {
		if basePackages[p.PackageSPDXIdentifier] == nil {
			packages = append(packages, p)
			packageIds[p.PackageSPDXIdentifier] = struct{}{}
			otherLicenseIds[p.PackageLicenseDeclared] = struct{}{}
			otherLicenseIds[p.PackageLicenseConcluded] = struct{}{}
		}
	}

	var relationships []*spdx.Relationship
	for _, r := range doc.Relationships {
		_, okA := packageIds[r.RefA.ElementRefID]
		_, okB := packageIds[r.RefB.ElementRefID]
		if okA && okB {
			relationships = append(relationships, r)
		}
	}

	var otherLicenses []*spdx.OtherLicense
	for _, ol := range doc.OtherLicenses {
		if _, ok := otherLicenseIds[ol.LicenseIdentifier]; ok {
			otherLicenses = append(otherLicenses, ol)
		}
	}

	doc.Packages = packages
	doc.Relationships = relationships
	doc.OtherLicenses = otherLicenses
}

func (d *Diff) createBasePackages() map[spdx.ElementID]*spdx.Package {
	pkgs := make(map[spdx.ElementID]*spdx.Package)

	for _, pkg := range d.Base.Packages {
		pkgs[pkg.PackageSPDXIdentifier] = pkg
	}

	return pkgs
}
