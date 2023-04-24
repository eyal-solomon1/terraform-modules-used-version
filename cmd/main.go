package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kataras/tablewriter"
	"github.com/lensesio/tableprinter"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	moduleName := flag.String("module", "", "The name of the module to check for.")
	flag.Parse()

	if *moduleName == "" {
		panic("Must specify a module name.")
	}

	var terraformDirs = []string{".terraform", ".terragrunt-cache"}
	var existedFolder string

	cwd, err := os.Getwd()
	check(err)

	var errorCount int
	for _, dir := range terraformDirs {
		dirPath := filepath.Join(cwd, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			errorCount++
			if errorCount > 1 {
				fmt.Fprint(os.Stderr, "couldn't find '.terraform' or '.terragrunt-cache' directory\n")
				os.Exit(1)
			}
		} else {
			existedFolder = dir
		}
	}

	var modulesFile string

	err = filepath.Walk(existedFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "modules.json" {
			modulesFile = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error searching for modules.json file in '%s': %v\n", existedFolder, err)
		os.Exit(1)
	}

	if modulesFile == "" {
		fmt.Printf("Could not find 'modules.json' file in '%s'.\n", existedFolder)
		os.Exit(1)
	}

	data, err := os.ReadFile(modulesFile)
	check(err)

	var modules Modules

	if err := json.Unmarshal([]byte(data), &modules); err != nil {
		panic(err)
	}

	var foundModules = formatModules(modules, *moduleName)

	if len(foundModules) == 0 {
		panic(fmt.Sprintf("found no results for '%s' in modules.json", *moduleName))
	}

	printer := tableprinter.New(os.Stdout)
	printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
	printer.CenterSeparator = "│"
	printer.ColumnSeparator = "│"
	printer.RowSeparator = "─"
	printer.HeaderBgColor = tablewriter.BgBlackColor
	printer.HeaderFgColor = tablewriter.FgGreenColor

	printer.Print(foundModules)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Returns a formated formated module name where the provided substring is highligthed
func colorModuleName(moduleName string, searchPattern string) string {
	colorCode := "\033[31m"
	resetCode := "\033[0m"

	indices := []int{}
	for i := 0; i < len(moduleName); {
		index := strings.Index(moduleName[i:], searchPattern)
		if index == -1 {
			break
		}
		indices = append(indices, i+index)
		i += index + 1
	}

	if len(indices) > 0 {
		coloredString := ""
		lastIndex := 0
		for _, index := range indices {
			coloredString += moduleName[lastIndex:index] + colorCode + moduleName[index:index+len(searchPattern)] + resetCode
			lastIndex = index + len(searchPattern)
		}
		coloredString += moduleName[lastIndex:]
		return coloredString
	} else {
		return moduleName
	}
}

// Returns a formated modules table based on your substring input
func formatModules(modules Modules, substring string) []foundModule {
	var foundModules = []foundModule{}
	for _, module := range modules.Modules {
		if strings.Contains(module.Source, substring) && len(module.Version) > 0 {
			foundModules = append(foundModules, foundModule{ModuleName: colorModuleName(module.Source, substring), Version: module.Version, ModuleLocalName: fmt.Sprintf("'%s'", module.Key)})
		}
	}
	return foundModules
}

type Module struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version,omitempty"`
	Dir     string `json:"Dir"`
}

type Modules struct {
	Modules []Module `json:"Modules"`
}

type foundModule struct {
	ModuleName      string `header:"module name"`
	Version         string `header:"value"`
	ModuleLocalName string `header:"local name"`
}
