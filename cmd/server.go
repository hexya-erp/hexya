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

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/controllers"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/menus"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/templates"
	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/hexya-erp/hexya/src/views"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
		runProject(projectDir, "server", args)
	},
}

// runProject builds and run the executable of the given project with the given command and arguments.
func runProject(projectDir, cmd string, args []string) {
	fmt.Println("Please wait, Hexya is starting ...")
	viper.Set("ProjectDir", projectDir)
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		panic(err)
	}

	cmdName := filepath.Base(absProjectDir)
	runCommand("go", "build", "-o", cmdName, absProjectDir)
	runCommand(filepath.Join(absProjectDir, cmdName), append([]string{cmd}, args...)...)
}

// StartServer starts the Hexya server. It is meant to be called from
// a project start file which imports all the project's module.
func StartServer() {
	setupLogger()
	defer log.Sync()
	setupDebug()
	resourceDir, err := filepath.Abs(viper.GetString("ResourceDir"))
	if err != nil {
		log.Panic("Unable to find Resource directory", "error", err)
	}
	server.ResourceDir = resourceDir
	server.PreInit()
	connectToDB()
	i18n.BootStrap()
	models.BootStrap()
	server.LoadTranslations(resourceDir, i18n.Langs)
	server.LoadInternalResources(resourceDir)
	views.BootStrap()
	templates.BootStrap()
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
	models.DBConnect(viper.GetString("DB.Driver"), models.ConnectionParams{
		Host:     viper.GetString("DB.Host"),
		Port:     viper.GetString("DB.Port"),
		User:     viper.GetString("DB.User"),
		Password: viper.GetString("DB.Password"),
		DBName:   viper.GetString("DB.Name"),
		SSLMode:  viper.GetString("DB.SSLMode"),
		SSLCert:  viper.GetString("DB.SSLCert"),
		SSLKey:   viper.GetString("DB.SSLKey"),
		SSLCA:    viper.GetString("DB.SSLCA"),
	})
}

// SetServerFlags adds the server flags to the given command.
func SetServerFlags(c *cobra.Command) {
	c.PersistentFlags().StringP("interface", "i", "", "Interface on which the server should listen. Empty string is all interfaces")
	viper.BindPFlag("Server.Interface", c.PersistentFlags().Lookup("interface"))
	c.PersistentFlags().StringP("port", "p", "8080", "Port on which the server should listen.")
	viper.BindPFlag("Server.Port", c.PersistentFlags().Lookup("port"))
	c.PersistentFlags().StringSliceP("languages", "l", []string{}, "Comma separated list of language codes to load (ex: fr,de,es).")
	viper.BindPFlag("Server.Languages", c.PersistentFlags().Lookup("languages"))
	c.PersistentFlags().StringP("domain", "d", "", "Domain name of the server. When set, interface and port are set to 0.0.0.0:443 and it will automatically get an HTTPS certificate from Letsencrypt")
	viper.BindPFlag("Server.Domain", c.PersistentFlags().Lookup("domain"))
	c.PersistentFlags().StringP("certificate", "C", "", "Certificate file for HTTPS. If neither certificate nor domain is set, the server will run on plain HTTP. When certificate is set, private-key must also be set.")
	viper.BindPFlag("Server.Certificate", c.PersistentFlags().Lookup("certificate"))
	c.PersistentFlags().StringP("private-key", "K", "", "Private key file for HTTPS.")
	viper.BindPFlag("Server.PrivateKey", c.PersistentFlags().Lookup("private-key"))
}

func runCommand(c string, args ...string) error {
	cmdToRun := exec.Command(c, args...)
	cmdToRun.Stdout = os.Stdout
	cmdToRun.Stderr = os.Stderr
	return cmdToRun.Run()
}

func init() {
	SetServerFlags(serverCmd)
	HexyaCmd.AddCommand(serverCmd)
}
