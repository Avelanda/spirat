package spirat

import "github.com/Hitachi/spirat/pkgmanager"

type Spirat struct {
	Command string    `json:"command"`
	Version string    `json:"version"`
	Results []*Result `json:"results"`
}

type Result struct {
	PackageManager string                  `json:"packageManager"`
	QueryResult    *pkgmanager.QueryResult `json:"queryResult"`
}
