// Copyright 2017 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hexya-erp/hexya/src/tools/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/packages"
)

const (
	// PoolDirRel is the name of the generated pool directory (relative to the current project root)
	PoolDirRel = "pool"
	// ResDirRel is the name of the resources directory (relative to the current project root)
	ResDirRel = "res"
	// TempEmpty is the name of the temporary go file in the pool directory for startup
	TempEmpty     = "temp.go"
	startFileName = "main.go"
)

var generateCmd = &cobra.Command{
	Use:   "generate PROJECT_DIR",
	Short: "Generate the source code of the model pool",
	Long: `Generate the source code of the pool package which includes the definition of all the models.
This command also :
- creates the resource directory by symlinking all modules resources into the project directory.
- creates or updates the main.go of the project.
This command must be rerun after each source code modification, including module import.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("You must specify the project directory ")
			os.Exit(1)
		}
		runGenerate(args[0])
	},
}

var symlinkDirs = []string{"static", "data", "demo", "resources", "i18n"}

var (
	generateEmptyPool bool
	testEnabled       bool
)

func init() {
	HexyaCmd.AddCommand(generateCmd)
	generateCmd.Flags().BoolVarP(&testEnabled, "test", "t", false, "Generate pool for testing a module. When set projectDir must be the source directory of the module.")
	generateCmd.Flags().BoolVar(&generateEmptyPool, "empty", false, "Generate an empty pool package and returns. When set, resource dir and main.go are untouched.")
}

func runGenerate(projectDir string) {
	projectDir, poolDir := computeDirs(projectDir)
	cleanPoolDir(poolDir)
	if generateEmptyPool {
		return
	}
	var targetPaths []string
	if testEnabled {
		targetPaths = []string{projectDir}
	} else {
		targetPaths = viper.GetStringSlice("Modules")
	}
	replacePoolDirInGoMod(poolDir)

	fmt.Println(`Hexya Generate
	--------------`)
	fmt.Println("Modules paths:")
	fmt.Println(" -", strings.Join(targetPaths, "\n - "))

	fmt.Print(`1/5 - Loading program...`)
	packs, err := loadProgram(targetPaths)
	if err != nil {
		panic(err)
	}
	mods := generate.GetModulePackages(packs)
	fmt.Println("Ok")

	fmt.Print("2/5 - Generating symlinks...")
	createSymlinks(mods, projectDir)
	fmt.Println("Ok")

	fmt.Print("3/5 - Generating pool...")
	generate.CreatePool(mods, poolDir)
	fmt.Println("Ok")

	fmt.Print("4/5 - Checking the generated code...")
	_, err = loadProgram(targetPaths)
	if err != nil {
		fmt.Println("FAIL")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Ok")

	fmt.Print("5/5 - Creating main.go in project...")
	if testEnabled {
		fmt.Println("SKIPPED")
	} else {
		createStartFile(projectDir, targetPaths)
		fmt.Println("Ok")
	}

	fmt.Println("Pool generated successfully")
}

func createStartFile(projectDir string, targetPaths []string) {
	cmdName := filepath.Base(projectDir)
	tmplData := struct {
		Imports    []string
		Executable string
	}{
		Imports:    targetPaths,
		Executable: cmdName,
	}
	sfn := filepath.Join(projectDir, startFileName)
	generate.CreateFileFromTemplate(sfn, startFileTemplate, tmplData)
}

func createSymlinks(modules []*generate.ModuleInfo, projectDir string) {
	cleanModuleSymlinks(projectDir)
	for _, m := range modules {
		if m.ModType != generate.Base {
			continue
		}
		createModuleSymlinks(m, projectDir)
	}
}

func loadProgram(targetPaths []string) ([]*packages.Package, error) {
	conf := packages.Config{
		Mode: packages.LoadAllSyntax,
	}
	packs, err := packages.Load(&conf, targetPaths...)
	return packs, err
}

func replacePoolDirInGoMod(poolDir string) {
	runCommand("go", "mod", "edit", "-replace", fmt.Sprintf("github.com/hexya-erp/pool=%s", poolDir))
}

func computeDirs(projectDir string) (string, string) {
	poolDir, err := filepath.Abs(filepath.Join(projectDir, PoolDirRel))
	if err != nil {
		panic(err)
	}
	return projectDir, poolDir
}

// cleanPoolDir removes all files in the given directory and leaves only
// one empty file declaring package 'pool'.
func cleanPoolDir(dirName string) {
	os.RemoveAll(dirName)
	modelsDir := filepath.Join(dirName, generate.PoolModelPackage)
	queryDir := filepath.Join(dirName, generate.PoolQueryPackage)
	interfacesDir := filepath.Join(dirName, generate.PoolInterfacesPackage)
	os.MkdirAll(modelsDir, 0755)
	os.MkdirAll(queryDir, 0755)
	os.MkdirAll(interfacesDir, 0755)
	generate.CreateFileFromTemplate(filepath.Join(modelsDir, TempEmpty), emptyPoolTemplate, generate.PoolModelPackage)
	generate.CreateFileFromTemplate(filepath.Join(queryDir, TempEmpty), emptyPoolTemplate, generate.PoolQueryPackage)
	generate.CreateFileFromTemplate(filepath.Join(interfacesDir, TempEmpty), emptyPoolTemplate, generate.PoolInterfacesPackage)

	if err := writeFileFromTemplate(filepath.Join(dirName, "go.mod"), emptyPoolGoMod, nil); err != nil {
		log.Panic("Error while saving generated source file", "error", err, "fileName", "go.mod")
	}
	copyGoModReplaces(dirName)
}

func copyGoModReplaces(poolDir string) {
	type Module struct {
		Path    string
		Version string
	}
	type GoMod struct {
		Replace []struct {
			Old Module
			New Module
		}
	}
	modJSON, err := exec.Command("go", "mod", "edit", "-json").CombinedOutput()
	if err != nil {
		fmt.Println(string(modJSON))
		panic(err)
	}
	var replaces GoMod
	if err = json.Unmarshal(modJSON, &replaces); err != nil {
		panic(err)
	}
	for _, repl := range replaces.Replace {
		if repl.Old.Path == "github.com/hexya-erp/pool" {
			continue
		}
		oldPath := repl.Old.Path
		if repl.Old.Version != "" {
			oldPath += "@" + repl.Old.Version
		}
		newPath := repl.New.Path
		if repl.New.Version != "" {
			newPath += "@" + repl.New.Version
		}
		runCommand("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", oldPath, newPath), filepath.Join(poolDir, "go.mod"))
	}
}

func writeFileFromTemplate(fileName string, tmpl *template.Template, data interface{}) error {
	var buf bytes.Buffer
	tmpl.Execute(&buf, data)
	err := ioutil.WriteFile(fileName, buf.Bytes(), 0644)
	return err
}

// createModuleSymlinks create the symlinks of the given module in the
// project directory.
func createModuleSymlinks(mod *generate.ModuleInfo, projectDir string) {
	for _, dir := range symlinkDirs {
		mDir := filepath.Dir(mod.GoFiles[0])
		srcPath := filepath.Join(mDir, dir)
		dstPath := filepath.Join(projectDir, ResDirRel, dir)
		if _, err := os.Stat(srcPath); err != nil {
			// Subdir doesn't exist, so we don't symlink
			continue
		}
		if err := os.MkdirAll(dstPath, 0755); err != nil {
			panic(err)
		}
		if err := os.Symlink(srcPath, filepath.Join(dstPath, mod.Name)); err != nil {
			panic(err)
		}
	}
}

// cleanModuleSymlinks removes all symlinks in the server symlink directories.
// Note that this function actually removes and recreates the symlink directories.
func cleanModuleSymlinks(projectDir string) {
	for _, dir := range symlinkDirs {
		dirPath := filepath.Join(projectDir, ResDirRel, dir)
		os.RemoveAll(dirPath)
		os.Mkdir(dirPath, 0775)
	}
}

var emptyPoolTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package {{ . }}
`))

var emptyPoolGoMod = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

module github.com/hexya-erp/pool
`))

var startFileTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-server
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package main

import (
	"fmt"
	"os"

	"github.com/hexya-erp/hexya/cmd"
	"github.com/spf13/cobra"
{{ range .Imports }}	_ "{{ . }}"
{{ end }}
)

func main() {
	var hexyaCmd = &cobra.Command{
		Use:   "{{ .Executable }}",
		Short: "Hexya is an open source modular ERP",
		Long: "Hexya is an open source modular ERP written in Go. It is designed for high demand business data processing while being easily customizable",
	}
	cmd.SetHexyaFlags(hexyaCmd)

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the Hexya server",
		Long: "Start the Hexya server",
		Run: func(c *cobra.Command, args []string) {
			cmd.StartServer()
		},
	}
	hexyaCmd.AddCommand(serverCmd)
	cmd.SetServerFlags(serverCmd)

	var updateDBCmd = &cobra.Command{
		Use:   "updatedb",
		Short: "Update the database schema",
		Long: "Synchronize the database schema with the models definitions.",
		Run: func(c *cobra.Command, args []string) {
			cmd.UpdateDB()
		},
	}
	hexyaCmd.AddCommand(updateDBCmd)

	cobra.OnInitialize(cmd.InitConfig)

	if err := hexyaCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
`))
