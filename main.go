package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Hitachi/spirat/pkgmanager"
	"github.com/Hitachi/spirat/reporter"
	"github.com/Hitachi/spirat/spirat"
	"github.com/spdx/tools-golang/spdx"
	"io"
	"os"
	"strings"
)

var (
	version string
	build   string

	diffFile string

	stdout  bool
	verbose bool

	toolNames string

	format   formatType
	filename string
	force    bool

	showVersion bool

	allFormats = [...]formatType{FormatPlain, FormatJson, FormatSpdxJson}
)

type formatType string

const (
	FormatPlain    formatType = "plain"
	FormatJson     formatType = "json"
	FormatSpdxJson formatType = "spdx-json"

	UsageMessage = `An SBOM generating tool

Usage: spirat [flags]

`
)

type formatFlag struct{}

func (f *formatFlag) Set(v string) error {
	format = formatType(v)
	for _, f := range allFormats {
		if format == f {
			return nil
		}
	}

	return fmt.Errorf("unknown format: %s", format)
}

func (f *formatFlag) String() string {
	var fs []string
	for _, f := range allFormats {
		fs = append(fs, string(f))
	}

	return strings.Join(fs, ", ")
}

func init() {
	flag.StringVar(&diffFile, "diff", "", "show additional packages from the specified file")

	flag.BoolVar(&stdout, "stdout", false, "output results to stdout")
	flag.BoolVar(&verbose, "verbose", false, "output verbose log")

	flag.StringVar(&toolNames, "tools", "", "output packages installed by the comma-separated specified tools")

	flag.Var(&formatFlag{}, "format", "specify file format: plain, json, spdx-json")
	flag.BoolVar(&force, "force", false, "overwrite existing file")
	flag.StringVar(&filename, "filename", "", "set filename")

	flag.BoolVar(&showVersion, "version", false, "show version")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprint(w, UsageMessage)

		flag.PrintDefaults()
	}
}

func main() {
	prepareFlags()
	diffJson := createBaseJsonForDiff()

	managers := getPackageManagers(toolNames)
	if len(managers) == 0 {
		exitWithMessage("no tool specified")
	}

	q, errs := runSpirat(managers)

	r := newReporter(q, format, diffJson)
	report, err := r.Report()
	exitIfError(err)

	if stdout {
		fmt.Print(report)
	} else {
		err = os.WriteFile(filename, []byte(report), os.ModePerm)
		exitIfError(err)
	}

	if errs != nil && verbose {
		fmt.Fprintln(os.Stderr, "some errors happen while querying packages.")
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func prepareFlags() {
	flag.Parse()

	if showVersion {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Version: %s\n", version)
		if build != "" {
			fmt.Fprintf(w, "Build: %s\n", build)
		}
		os.Exit(0)
	}

	if filename == "" && !stdout {
		if diffFile == "" {
			filename = "spirat" + extname()
		} else {
			filename = "spirat_diff" + extname()
		}
	}

	if diffFile != "" && format == "" {
		format = FormatSpdxJson
	}

	if !stdout {
		if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) && !force {
			exitWithMessage(fmt.Sprintf("%s already exists.\nUse -force option to overwrite the file.", filename))
		}
	}
}

func getPackageManagers(toolNames string) []pkgmanager.PackageManager {
	if toolNames == "" {
		return pkgmanager.GetAvailablePackageManagers()
	} else {
		return pkgmanager.GetPackageManagers(strings.Split(toolNames, ","))
	}
}

func createBaseJsonForDiff() *spdx.Document {
	if diffFile == "" {
		return nil
	}

	if _, err := os.Stat(diffFile); os.IsNotExist(err) {
		exitWithMessage(fmt.Sprintf("%s does not exist", diffFile))
	}

	if format != FormatSpdxJson {
		exitWithMessage("diff supports spdx-json only")
	}

	file, err := os.Open(diffFile)
	exitIfError(err)

	bytes, err := io.ReadAll(file)
	exitIfError(err)

	var diffJson spdx.Document
	err = json.Unmarshal(bytes, &diffJson)
	exitIfError(err)

	return &diffJson
}

func runSpirat(managers []pkgmanager.PackageManager) (*spirat.Spirat, []error) {
	q := spirat.Spirat{
		Command: strings.Join(os.Args, " "),
		Version: version,
		Results: []*spirat.Result{},
	}
	var allErrs []error

	for _, manager := range managers {
		var result *pkgmanager.QueryResult
		var errs []error

		result, errs = manager.Query()
		if result == nil {
			if len(errs) == 1 {
				exitIfError(errs[0])
			}

			exitIfError(fmt.Errorf("%v", errs))
		}

		q.Results = append(q.Results, &spirat.Result{PackageManager: manager.String(), QueryResult: result})
		allErrs = append(allErrs, errs...)
	}

	return &q, allErrs
}

func newReporter(spirat *spirat.Spirat, format formatType, diffJson *spdx.Document) reporter.Reporter {
	switch {
	case diffJson != nil:
		return &reporter.Diff{Spirat: spirat, Base: diffJson}
	case format == FormatSpdxJson:
		return &reporter.SpdxJson{Spirat: spirat}
	case format == FormatJson:
		return &reporter.Json{Spirat: spirat}
	case format == FormatPlain:
		return &reporter.Plain{Spirat: spirat}
	default:
		return &reporter.SpdxJson{Spirat: spirat}
	}
}

func exitIfError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func exitWithMessage(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func extname() string {
	switch {
	case format == "json":
		return ".json"
	case format == "spdx-json":
		return ".json"
	case format == "plain":
		return ".txt"
	default:
		return ".json"
	}
}
