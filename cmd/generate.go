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
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/tools/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/packages"
)

const (
	// PoolDirRel is the name of the generated pool directory (relative to the hexya root)
	PoolDirRel = "pool"
	// TempEmpty is the name of the temporary go file in the pool directory for startup
	TempEmpty = "temp.go"
)

var generateCmd = &cobra.Command{
	Use:   "generate [projectDir]",
	Short: "Generate the source code of the model pool",
	Long: `Generate the source code of the pool package which includes the definition of all the models.
This command also creates the resource directory by symlinking all modules resources into the project directory.
This command must be rerun after each source code modification, including module import.`,
	Run: func(cmd *cobra.Command, args []string) {
		if testedModule != "" {
			runGenerate("")
			return
		}
		if len(args) == 0 {
			fmt.Println("You must specify the project directory or a module to test")
			os.Exit(1)
		}
		runGenerate(args[0])
	},
}

var symlinkDirs = []string{"static", "data", "demo", "resources", "i18n"}

var (
	generateEmptyPool bool
	testedModule      string
)

func init() {
	HexyaCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&testedModule, "test", "t", "", "Generate pool for testing the module in the given source directory. When set projectDir is ignored.")
	generateCmd.Flags().BoolVar(&generateEmptyPool, "empty", false, "Generate an empty pool package. When set projectDir is ignored.")
}

func runGenerate(projectDir string) {
	poolDir := filepath.Join(generate.HexyaDir, PoolDirRel)
	cleanPoolDir(poolDir)
	if generateEmptyPool {
		return
	}

	fmt.Println(`Hexya Generate
--------------`)
	fmt.Printf("Detected Hexya root directory at %s.\n", generate.HexyaDir)

	targetPaths := viper.GetStringSlice("Modules")
	if testedModule != "" {
		testDir, _ := filepath.Abs(testedModule)
		importPack, err := packages.Load(&packages.Config{Mode: packages.LoadFiles}, testDir)
		if err != nil {
			panic(err)
		}
		if len(importPack) == 0 {
			panic(fmt.Errorf("no package found at %s", testDir))
		}
		targetPaths = []string{importPack[0].PkgPath}
	}
	fmt.Println("Modules paths:")
	fmt.Println(" -", strings.Join(targetPaths, "\n - "))

	fmt.Print(`Loading program...`)

	conf := packages.Config{
		Mode: packages.LoadAllSyntax,
	}
	packs, _ := packages.Load(&conf, targetPaths...)
	fmt.Println("Ok")

	if projectDir != "" {
		fmt.Print("Generating symlinks...")
		modules := generate.GetModulePackages(packs)
		cleanModuleSymlinks(projectDir)
		for _, m := range modules {
			if m.ModType != generate.Base {
				continue
			}
			createModuleSymlinks(m, projectDir)
		}
		fmt.Println("Ok")
	}

	fmt.Print("Generating pool...")
	generate.CreatePool(packs, poolDir)
	fmt.Println("Ok")

	fmt.Print("Checking the generated code...")
	_, err := packages.Load(&conf, targetPaths...)
	if err != nil {
		fmt.Println("FAIL")
		fmt.Println(err)
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
// project directory.
func createModuleSymlinks(mod *generate.ModuleInfo, projectDir string) {
	for _, dir := range symlinkDirs {
		mDir := filepath.Dir(mod.GoFiles[0])
		srcPath := filepath.Join(mDir, dir)
		dstPath := filepath.Join(projectDir, "res", dir)
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
		dirPath := filepath.Join(projectDir, "res", dir)
		os.RemoveAll(dirPath)
		os.Mkdir(dirPath, 0775)
	}
}

var emptyPoolTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package {{ . }}
`))
