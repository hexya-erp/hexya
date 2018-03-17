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
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/controllers"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/hexya-erp/hexya/hexya/views"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const startFileName = "start.go"

var serverCmd = &cobra.Command{
	Use:   "server [projectDir]",
	Short: "Start the Hexya server",
	Long: `Start the Hexya server of the project in 'projectDir'.
If projectDir is omitted, defaults to the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := "."
		if len(args) > 0 {
			projectDir = args[0]
		}
		generateAndRunFile(projectDir, startFileName, startFileTemplate)
	},
}

// generateAndRunFile creates the startup file of the project and runs it.
func generateAndRunFile(projectDir, fileName string, tmpl *template.Template) {
	fmt.Println("Please wait, Hexya is starting ...")
	conf := viper.AllSettings()
	delete(conf, "modules")

	modules := viper.GetStringSlice("Modules")

	tmplData := struct {
		Imports []string
		Config  string
	}{
		Imports: modules,
		Config:  fmt.Sprintf("%#v", conf),
	}
	startFileName := filepath.Join(projectDir, fileName)
	generate.CreateFileFromTemplate(startFileName, tmpl, tmplData)
	cmd := exec.Command("go", "run", startFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// StartServer starts the Hexya server. It is meant to be called from
// a project start file which imports all the project's module.
func StartServer(config map[string]interface{}) {
	setupConfig(config)
	setupLogger()
	setupDebug()
	server.PreInit()
	connectToDB()
	models.BootStrap()
	i18n.BootStrap()
	server.LoadTranslations(i18n.Langs)
	server.LoadInternalResources()
	views.BootStrap()
	actions.BootStrap()
	controllers.BootStrap()
	menus.BootStrap()
	server.PostInit()
	srv := server.GetServer()
	address := fmt.Sprintf("%s:%s", viper.GetString("Server.Interface"), viper.GetString("Server.Port"))
	cert := viper.GetString("Server.Certificate")
	key := viper.GetString("Server.PrivateKey")
	domain := viper.GetString("Server.Domain")
	switch {
	case cert != "":
		srv.RunTLS(address, cert, key)
	case domain != "":
		srv.RunAutoTLS(domain)
	default:
		srv.Run(address)
	}
}

// setupConfig takes the given config map and stores it into the viper configuration
func setupConfig(config map[string]interface{}) {
	for key, value := range config {
		viper.Set(key, value)
	}
}

// setupLogger initializes the logger
func setupLogger() {
	logging.Initialize()
	log = logging.GetLogger("init")
}

// setupDebug updates the server for debugging if Debug is enabled
func setupDebug() {
	if !viper.GetBool("Debug") {
		return
	}
	gin.SetMode(gin.DebugMode)
	pprof.Register(server.GetServer().Engine)
}

// connectToDB creates the connection to the database
func connectToDB() {
	connectString := fmt.Sprintf("dbname=%s sslmode=%s", viper.GetString("DB.Name"), viper.GetString("DB.SSLMode"))
	if viper.GetString("DB.User") != "" {
		connectString += fmt.Sprintf(" user=%s", viper.GetString("DB.User"))
	}
	if viper.GetString("DB.Password") != "" {
		connectString += fmt.Sprintf(" password=%s", viper.GetString("DB.Password"))
	}
	if viper.GetString("DB.Host") != "" {
		connectString += fmt.Sprintf(" host=%s", viper.GetString("DB.Host"))
	}
	if viper.GetString("DB.Port") != "5432" {
		connectString += fmt.Sprintf(" port=%s", viper.GetString("DB.Port"))
	}
	models.DBConnect(viper.GetString("DB.Driver"), connectString)
}

func init() {
	serverCmd.PersistentFlags().StringP("interface", "i", "", "Interface on which the server should listen. Empty string is all interfaces")
	viper.BindPFlag("Server.Interface", serverCmd.PersistentFlags().Lookup("interface"))
	serverCmd.PersistentFlags().StringP("port", "p", "8080", "Port on which the server should listen.")
	viper.BindPFlag("Server.Port", serverCmd.PersistentFlags().Lookup("port"))
	serverCmd.PersistentFlags().StringSliceP("languages", "l", []string{}, "Comma separated list of language codes to load (ex: fr,de,es).")
	viper.BindPFlag("Server.Languages", serverCmd.PersistentFlags().Lookup("languages"))
	serverCmd.PersistentFlags().StringP("domain", "d", "", "Domain name of the server. When set, interface and port are set to 0.0.0.0:443 and it will automatically get an HTTPS certificate from Letsencrypt")
	viper.BindPFlag("Server.Domain", serverCmd.PersistentFlags().Lookup("domain"))
	serverCmd.PersistentFlags().StringP("certificate", "C", "", "Certificate file for HTTPS. If neither certificate nor domain is set, the server will run on plain HTTP. When certificate is set, private-key must also be set.")
	viper.BindPFlag("Server.Certificate", serverCmd.PersistentFlags().Lookup("certificate"))
	serverCmd.PersistentFlags().StringP("private-key", "K", "", "Private key file for HTTPS.")
	viper.BindPFlag("Server.PrivateKey", serverCmd.PersistentFlags().Lookup("private-key"))
	HexyaCmd.AddCommand(serverCmd)
}

var startFileTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-server
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package main

import (
	"github.com/hexya-erp/hexya/cmd"
{{ range .Imports }}	_ "{{ . }}"
{{ end }}
)

func main() {
	cmd.StartServer({{ .Config }})
}
`))
