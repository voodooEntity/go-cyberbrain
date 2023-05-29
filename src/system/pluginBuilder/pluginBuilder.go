package pluginBuilder

import (
	"errors"
	"github.com/voodooEntity/archivist"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var shell string

func BuildPlugins(pluginsPath string, abilitiesPath string) {
	archivist.Info("| Running cyberbrain PluginBuilder.")
	// running a detection for what shell to use to build the plugins
	detectShell()

	abilities := getAbilityDirectories(abilitiesPath)
	if 0 == len(abilities) {
		archivist.Error("There are no available ability plugins found at abilitiesPath directory '" + abilitiesPath + "'")
		os.Exit(0)
	}

	for _, ability := range abilities {
		archivist.Info("++ Plugin: " + ability)
		err := BuildPlugin(pluginsPath, abilitiesPath, ability)
		if nil != err {
			archivist.Error("- Error building plugin '" + ability + "' - Error: " + err.Error())
		}
	}
	archivist.Info("PluginBuilder finished - exiting.")
}

func BuildPlugin(pluginsPath string, basePath string, abilityPath string) error {
	archivist.Info("++ Plugin: " + abilityPath)
	archivist.Info("#go build -buildmode plugin -o " + pluginsPath + abilityPath + ".so " + basePath + abilityPath + "/" + abilityPath + ".go")
	_, err := custExec("go build -buildmode plugin -o " + pluginsPath + abilityPath + ".so " + basePath + abilityPath + "/" + abilityPath + ".go")
	if nil != err {
		archivist.Error("- Error building plugin '" + abilityPath + "' - Error: " + err.Error())
		return err
	}
	return nil
}

func custExec(cmd string) (string, error) {
	archivist.Debug("Executing following command: " + cmd)
	output, err := exec.Command(shell, "-c", cmd).Output()
	if nil != err {
		return "", errors.New("Error executin command'" + err.Error() + "'")
	}
	return strings.TrimRight(string(output), "\n"), nil
}

func getAbilityDirectories(abilitiesPath string) []string {
	var abilities []string
	dirs, err := ioutil.ReadDir(abilitiesPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range dirs {
		abilities = append(abilities, f.Name())
	}
	return abilities
}

func detectShell() {
	// detect shell
	executer, exists := os.LookupEnv("SHELL")
	if !exists {
		// Print the value of the environment variable
		archivist.Error("No SHELL env given, exiting")
		os.Exit(1)
	}
	shell = executer
}
