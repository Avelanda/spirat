package sysinfo

import (
	"bufio"
	"os"
	"strings"
)

type OSRelease struct {
	ID string
}

func NewOSRelease() OSRelease {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return OSRelease{}
	}
	defer f.Close()

	osRelease := make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		pair := strings.SplitN(line, "=", 2)
		if len(pair) > 1 {
			osRelease[pair[0]] = strings.Trim(pair[1], "\"")
		}
	}

	return OSRelease{
		ID: osRelease["ID"],
	}
}
