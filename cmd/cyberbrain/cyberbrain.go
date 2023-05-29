package main

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gitsapi"
	gitsapiConfig "github.com/voodooEntity/gitsapi/src/config"
	"github.com/voodooEntity/go-cyberbrain/src/system/api"
	"github.com/voodooEntity/go-cyberbrain/src/system/cli"
	"github.com/voodooEntity/go-cyberbrain/src/system/core"
	"github.com/voodooEntity/go-cyberbrain/src/system/pluginBuilder"
)

func main() {
	//initially we gonne parse the cli args
	cli.ParseArgs()

	switch cli.Data.Command {
	case "run":
		run()
	case "buildPlugins":
		buildPlugins()
	case "testPlugin":
		archivist.Info("Not implemented yet")
	default:
		cli.PrintHelpText()
	}

}

func run() {

	// init gitsapi the config
	// temporaray hardcoded configs here
	gitsApiCfg := map[string]string{
		"HOST":           "127.0.0.1",
		"PORT":           "1984",
		"PERSISTENCE":    "active",
		"LOG_TARGET":     "stdout",
		"LOG_PATH":       "out.log",
		"LOG_LEVEL":      "info",
		"CORS_HEADER":    "*",
		"CORS_ORIGIN":    "*",
		"SSL_CERT_FILE":  "rsa.crt",
		"SSL_KEY_FILE":   "rsa.key",
		"TOKEN_LIFETIME": "3600",
		"AUTH_ACTIVE":    "no",
	}
	if cli.Data.Verbose {
		gitsApiCfg["LOG_LEVEL"] = "debug"
	}

	gitsapiConfig.Init(gitsApiCfg)

	// init the archivist logger ### maybe will access a different config later on
	// prolly should access the one of bezel not the gitsapi one. for now, we gonne stick with it ###
	archivist.Init(gitsapiConfig.GetValue("LOG_LEVEL"), gitsapiConfig.GetValue("LOG_TARGET"), gitsapiConfig.GetValue("LOG_PATH"))

	// initing some additional application specific endpoints
	api.Extend()

	// than we init the core
	core.Init(gitsapiConfig.Data)

	// start the actual gitsapi
	gitsapi.Start()
}

func buildPlugins() {
	pluginBuilder.BuildPlugins(cli.Data.ProjectPath+"plugins/", cli.Data.PluginSource)
}
