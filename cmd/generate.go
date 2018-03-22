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
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	// We need to import models because of generated code
	_ "github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/loader"
)

const (
	// PoolDirRel is the name of the generated pool directory (relative to the hexya root)
	PoolDirRel = "pool"
	// TempEmpty is the name of the temporary go file in the pool directory for startup
	TempEmpty = "temp.go"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the source code of the model pool",
	Long: `Generate the source code of the pool package which includes the definition of all the models.
Additionally, this command creates the startup file of the project.
This command must be rerun after each source code modification, including module import.`,
	Run: func(cmd *cobra.Command, args []string) {
		runGenerate()
	},
}

var symlinkDirs = []string{"static", "templates", "data", "demo", "resources", "i18n"}

var (
	generateEmptyPool bool
	testedModule      string
)

func init() {
	HexyaCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&testedModule, "test", "t", "", "Generate pool for testing the module in the given source directory. When set projectDir is ignored.")
	generateCmd.Flags().BoolVar(&generateEmptyPool, "empty", false, "Generate an empty pool package. When set projectDir is ignored.")
}

func runGenerate() {
	poolDir := filepath.Join(generate.HexyaDir, PoolDirRel)
	cleanPoolDir(poolDir)
	if generateEmptyPool {
		return
	}

	conf := loader.Config{
		AllowErrors: true,
	}

	fmt.Println(`Hexya Generate
------------`)
	fmt.Printf("Detected Hexya root directory at %s.\n", generate.HexyaDir)

	targetPaths := viper.GetStringSlice("Modules")
	if testedModule != "" {
		testDir, _ := filepath.Abs(testedModule)
		importPack, err := build.ImportDir(testDir, 0)
		if err != nil {
			panic(err)
		}
		targetPaths = []string{importPack.ImportPath}
	}
	fmt.Println("Modules paths:")
	fmt.Println(" -", strings.Join(targetPaths, "\n - "))
	for _, ip := range targetPaths {
		conf.Import(ip)
	}

	fmt.Println(`Loading program...
Warnings may appear here, just ignore them if hexya-generate doesn't crash.`)

	program, _ := conf.Load()
	fmt.Println("Ok")

	fmt.Print("Generating symlinks...")
	modules := generate.GetModulePackages(program)
	cleanModuleSymlinks()
	for _, m := range modules {
		if m.ModType != generate.Base {
			continue
		}
		pkg, err := build.Import(m.Pkg.Path(), "", 0)
		if err != nil {
			panic(err)
		}
		createModuleSymlinks(pkg)
	}
	fmt.Println("Ok")

	fmt.Print("Generating pool...")
	generate.CreatePool(program, poolDir)
	fmt.Println("Ok")

	fmt.Print("Checking the generated code...")
	conf.AllowErrors = false
	_, err := conf.Load()
	if err != nil {
		fmt.Println("FAIL", err)
		os.Exit(1)
	}
	fmt.Println("Ok")

	fmt.Println("Pool generated successfully")
}

// cleanPoolDir removes all files in the given directory and leaves only
// one empty file declaring package 'pool'.
func cleanPoolDir(dirName string) {
	os.RemoveAll(dirName)
	modelsDir := filepath.Join(dirName, generate.PoolModelPackage)
	queryDir := filepath.Join(dirName, generate.PoolQueryPackage)
	os.MkdirAll(modelsDir, 0755)
	os.MkdirAll(queryDir, 0755)
	generate.CreateFileFromTemplate(filepath.Join(modelsDir, TempEmpty), emptyPoolTemplate, generate.PoolModelPackage)
	generate.CreateFileFromTemplate(filepath.Join(queryDir, TempEmpty), emptyPoolTemplate, generate.PoolQueryPackage)
}

// createModuleSymlinks create the symlinks of the given module in the
// server directory.
func createModuleSymlinks(mod *build.Package) {
	for _, dir := range symlinkDirs {
		srcPath := filepath.Join(mod.Dir, dir)
		dstPath := filepath.Join(generate.HexyaDir, "hexya", "server", dir, mod.Name)
		if _, err := os.Stat(srcPath); err != nil {
			// Subdir doesn't exist, so we don't symlink
			continue
		}
		if err := os.Symlink(srcPath, dstPath); err != nil {
			panic(err)
		}
	}
}

// cleanModuleSymlinks removes all symlinks in the server symlink directories.
// Note that this function actually removes and recreates the symlink directories.
func cleanModuleSymlinks() {
	for _, dir := range symlinkDirs {
		dirPath := filepath.Join(generate.HexyaDir, "hexya", "server", dir)
		os.RemoveAll(dirPath)
		os.Mkdir(dirPath, 0775)
	}
}

var emptyPoolTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package {{ . }}
`))
