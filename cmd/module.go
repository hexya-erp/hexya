// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package cmd

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Module development utilities",
	Long:  `Hexya utilities for module development.`,
}

var moduleInitCmd = &cobra.Command{
	Use:   "init MODULE_PATH",
	Short: "Initialize a module",
	Long:  `Initialize and scaffold a new module in the current directory with the given path (e.g. github.com/myuser/my-hexya-module).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create the go.mod file
		if len(args) == 0 {
			fmt.Println("You must specify a module path.")
			os.Exit(1)
		}
		modulePath := args[0]
		if err := runCommand("go", "mod", "init", modulePath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Create the 000hexya.go file
		data := struct {
			ModuleName string
		}{
			ModuleName: path.Base(modulePath),
		}
		if err := writeFileFromTemplate("000hexya.go", hexyaGoTmpl, data); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Create standard directories
		for _, dir := range symlinkDirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println(err)
			}
		}
		runCommand("go", "mod", "tidy")
	},
}

var moduleCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the module directory",
	Long: `Clean the current directory from all generated and test artifacts.
You should use this command before committing your work.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand("go", "mod", "edit", "-dropreplace", "github.com/hexya-erp/pool")
		if err := removeProjectDir(PoolDirRel); err != nil {
			fmt.Println(err)
		}
		if err := removeProjectDir(ResDirRel); err != nil {
			fmt.Println(err)
		}
		runCommand("go", "mod", "tidy")
	},
}

func init() {
	HexyaCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleInitCmd)
	moduleCmd.AddCommand(moduleCleanCmd)
}

var hexyaGoTmpl = template.Must(template.New("").Parse(`
package {{ .ModuleName }}

import (
	"github.com/hexya-erp/hexya/src/server"
	// blank import here this hexya module dependencies 
)

const MODULE_NAME string = {{ .ModuleName }}

func init() {
	server.RegisterModule(&server.Module{
		Name:     MODULE_NAME,
		PostInit: func() {},
	})
}
`))
