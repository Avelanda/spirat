package reporter

import (
	"bytes"
	"fmt"
	"github.com/Hitachi/spirat/spirat"
	"github.com/fatih/color"
	"strings"
)

type Reporter interface {
	Report() (string, error)
}

type Plain struct {
	Spirat *spirat.Spirat
}

func (p *Plain) Report() (string, error) {
	var ret bytes.Buffer

	for _, r := range p.Spirat.Results {
		for _, pkg := range r.QueryResult.Packages {
			color.New(color.FgGreen).Fprint(&ret, pkg.Name)

			fmt.Fprint(&ret, " ")

			color.New(color.FgBlue).Fprint(&ret, pkg.Version)

			fmt.Fprint(&ret, "\n")

			if pkg.Filename != "" {
				fmt.Fprintf(&ret, "  Filename: %s\n", pkg.Filename)
			}

			if pkg.HomepageUrl != "" {
				fmt.Fprintf(&ret, "  URL: %s\n", pkg.HomepageUrl)
			}

			var licenseNames []string
			for _, license := range pkg.Licenses {
				licenseNames = append(licenseNames, license.Name)
			}
			licenses := strings.Join(licenseNames, "/")

			if licenses != "" {
				fmt.Fprintf(&ret, "  License: %s\n", licenses)
			} else {
				fmt.Fprintf(&ret, "  License: Not found\n")
			}

			fmt.Fprint(&ret, "\n")
		}
	}

	return ret.String(), nil
}
