package cli

import (
	"flag"
	"github.com/voodooEntity/archivist"
	"log"
	"os"
)

type Args struct {
	PluginSource string
	ProjectPath  string
	Command      string
	Verbose      bool
}

var shell string
var Data Args
var loggerOut = log.New(os.Stdout, "", 0)

func ParseArgs() {
	// first we check for the help flag
	if 1 < len(os.Args) {
		if ok := os.Args[1]; ok == "help" {
			PrintHelpText()
			os.Exit(1)
		}
	}

	// plugin source director - to grab the code for new plugins
	var pluginSource string
	flag.StringVar(&pluginSource, "source", "", "The plugin code source directory path")

	// the project path for plugin compiling and config checking
	var projectPath string
	flag.StringVar(&projectPath, "project", "", "Your project directory path")

	// amount of runs in case you dont provide input
	var Command string
	flag.StringVar(&Command, "command", "", "Command to be executed")

	// verbose output flag
	verboseFlag := flag.Bool("verbose", false, "Enable verbose mode")

	flag.Parse()

	Data = Args{
		PluginSource: pluginSource,
		ProjectPath:  projectPath,
		Command:      Command,
		Verbose:      *verboseFlag,
	}

}

func PrintHelpText() {
	helpText := "Cyberbrain Help:\n" +
		"> Cyberbrain is an smart data processing architecture. Using this command you can \n" +
		"  build or test plugins for your cyberbrain application - or start the cyberbrain \n" +
		"  application to run. For further information please check the docs in the reps\n" +
		"   https://github.com/voodooEntity/go-cyberbrain readme.\n\n" +
		"  Args: \n" +
		"    -source \"/path/to/source/\"    | Path to the directory holding your plugin's sourcecode\n" +
		"    -project \"/path/to/source/\"   | Path your project has been inited in. Will hold the builded\n" +
		"                                      plugins for your app and optional a config (not supported yet)\n" +
		"    -command \"commandname\"        | The command you want to execut. Right now you can use\n" +
		"                                      run   (will start the application with the plugins given)\n" +
		"                                      buildPlugins ( will build the plugins from provided -source path\n" +
		"                                                    to provided -project provided path )\n" +
		"                                      testPlugin ( will testrun the plugin source at the given -source \n" +
		"                                                    path using the provided testdata.json )\n" +
		"    --verbose                    | Activates verbose output mode\n"
	archivist.Info(helpText)
}
