package reporter

import (
	"encoding/json"
	"github.com/Hitachi/spirat/pkgmanager"
	"github.com/Hitachi/spirat/sbom"
	"github.com/Hitachi/spirat/spirat"
	"github.com/Hitachi/spirat/utils"
)

type SpdxJson struct {
	Spirat *spirat.Spirat
}

func (j *SpdxJson) Report() (string, error) {
	qrs := utils.Map(j.Spirat.Results, func(r *spirat.Result) *pkgmanager.QueryResult {
		return r.QueryResult
	})
	doc := sbom.ToSpdx(qrs)
	jsonStr, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonStr), nil
}
