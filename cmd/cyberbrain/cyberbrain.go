package main

import (
	"encoding/json"
	"github.com/voodooEntity/go-cyberbrain/src/system/mapper"
	"github.com/voodooEntity/go-cyberbrain/src/system/observer"
	"github.com/voodooEntity/go-cyberbrain/src/system/registry"
	"github.com/voodooEntity/go-cyberbrain/src/system/scheduler"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gitsapi"
	gitsapiConfig "github.com/voodooEntity/gitsapi/src/config"
	"github.com/voodooEntity/go-cyberbrain/src/system/api"
	"github.com/voodooEntity/go-cyberbrain/src/system/cli"
	"github.com/voodooEntity/go-cyberbrain/src/system/core"
	"github.com/voodooEntity/go-cyberbrain/src/system/pluginBuilder"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
)

const version = "v0.1.2"

func main() {
	//initially we gonne parse the cli args
	cli.ParseArgs()

	// initially set the logger to stdout with given flag
	// run might overwrite this by config later on ### overthink
	logLevel := "info"
	if cli.Data.Verbose {
		logLevel = "debug"
	}
	archivist.Init(logLevel, "stdout", "")

	// dispatch what we do base on the command
	switch cli.Data.Command {
	case "run":
		run()
	case "build":
		buildPlugins(cli.Data.ProjectPath, cli.Data.Filter)
	case "test":
		buildPlugins(cli.Data.ProjectPath, cli.Data.Filter)
		testPlugins(cli.Data.ProjectPath, cli.Data.Filter)
	case "help":
		cli.PrintHelpText()
	case "version":
		archivist.Info("Cyberbrain " + version)
	}
}

func run() {

	// init gitsapi the config
	// temporaray hardcoded configs here
	gitsCfg := map[string]string{
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
		"PROTOCOL":       "http",
	}
	if cli.Data.Verbose {
		gitsCfg["LOG_LEVEL"] = "debug"
	}

	// than we init the core
	core.Init(gitsapiConfig.Data)

	switch mode := cli.Data.Mode; mode {
	case cli.RUN_MODE_CONTINOUUS:
		startContinouus(gitsCfg)
	case cli.RUN_MODE_ONESHOT:
		startOneshot()
	}

}

func startOneshot() {
	if "" == cli.Data.Stdin {
		archivist.Error("Missing stdin content for oneshot execution")
		return
	}

	// unpack the json
	var transportData transport.TransportEntity
	if err := json.Unmarshal([]byte(cli.Data.Stdin), &transportData); err != nil {
		archivist.Error("Invalid json input", err.Error())
		return
	}

	// lets pass the body to our mapper
	// that will recursive map the entities
	mappedData := mapper.MapTransportDataWithContext(transportData, "Data")
	rootType := mappedData.Type
	rootID := mappedData.ID
	scheduler.Run(mappedData, registry.Data)

	obs := observer.New(rootType, rootID)
	obs.Loop()
}

func startContinouus(gitsApiCfg map[string]string) {
	if cli.Data.Verbose {
		gitsApiCfg["LOG_LEVEL"] = "debug"
	}

	gitsapiConfig.Init(gitsApiCfg)

	// init the archivist logger ### maybe will access a different config later on
	// prolly should access the one of bezel not the gitsapi one. for now, we gonne stick with it ###
	archivist.Init(gitsapiConfig.GetValue("LOG_LEVEL"), gitsapiConfig.GetValue("LOG_TARGET"), gitsapiConfig.GetValue("LOG_PATH"))

	// initing some additional application specific endpoints
	api.Extend()

	// start the actual gitsapi
	gitsapi.Start()
}

func buildPlugins(projectPath string, filter string) {
	pluginBuilder.BuildPlugins(projectPath+"plugins/", projectPath+"src/", filter)
}

func testPlugins(projectPath string, filter string) {
	archivist.Info("+ Run cyberbrain tests")
	// first of all we get all the available plugins
	plugins := util.GetAvailablePlugins(projectPath + "plugins/")

	if 0 == len(plugins) {
		archivist.Info("There were no plugins to test found. Exiting")
		os.Exit(1)
	}

	if "" != filter {
		archivist.Info("+ filter applied '" + filter + "'")
	}

	for _, fileName := range plugins {
		if !strings.Contains(fileName, filter) {
			continue
		}
		pluginName := strings.TrimSuffix(fileName, ".so")
		pluginPath := projectPath + "src/" + pluginName + "/"
		if _, err := os.Stat(pluginPath + "test.json"); os.IsNotExist(err) {
			archivist.Warning("No test.json found in plugin directory '" + pluginPath + "' skipping the plugin '" + pluginName + "'")
			continue
		}

		testDataFile, err := os.Open(pluginPath + "test.json")
		if err != nil {
			archivist.Warning("Could not open file at  '"+pluginPath+"' skipping the plugin '"+pluginName+"' with error ", err.Error())
			continue
		}

		// Read the file contents
		contents, err := ioutil.ReadAll(testDataFile)
		if err != nil {
			archivist.Warning("Could not read file at  '"+pluginPath+"' skipping the plugin '"+pluginName+"' with error ", err.Error())
			continue
		}
		testDataFile.Close()

		// Unmarshal JSON content into []TransportEntity
		var data []transport.TransportEntity
		err = json.Unmarshal(contents, &data)
		if err != nil {
			archivist.Warning("Could json decode file at  '"+pluginPath+"' skipping the plugin '"+pluginName+"' with error ", err.Error())
			continue
		}

		testPlugin(projectPath, fileName, data)
	}

}

func testPlugin(projectPath string, pluginName string, testData []transport.TransportEntity) {
	archivist.Info("+ Testing plugin " + pluginName)
	plug, err := util.LoadPlugin(projectPath+"plugins/", pluginName)
	if nil != err {
		archivist.Error("Could not load given plugin "+pluginName+" with error: ", err.Error())
		return
	}

	// ### consider validating the input data for dependency string existing and for structure fitting the defined
	// dependency - probably worth building a generate command that will create a dependency json based on the getConfig()
	// results. tho this is further fancy stuff

	for _, dependency := range testData {
		archivist.Info("++ running tests cases for dependency " + dependency.Value)
		for _, testCase := range dependency.Children() {
			archivist.DebugF("> given input: %+v", testCase)

			// lets get a new instance of our plugin
			start := time.Now()
			plugInstance := plug.New()
			ret, execErr := plugInstance.Execute(testCase, dependency.Value, "testing")
			if nil != execErr {
				archivist.Error("++ plugin execution failed with error:", execErr.Error())
				continue
			}
			archivist.InfoF("++ plugin result: %+v", ret)
			archivist.DebugF("++ Execution time: %+v", time.Since(start))
		}
	}

}
