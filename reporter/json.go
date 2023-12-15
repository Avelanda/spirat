package reporter

import (
	"encoding/json"
	"github.com/Hitachi/spirat/spirat"
)

type Json struct {
	Spirat *spirat.Spirat
}

func (j *Json) Report() (string, error) {
	jsonStr, err := json.Marshal(j.Spirat)
	if err != nil {
		return "", err
	}

	return string(jsonStr), nil
}
