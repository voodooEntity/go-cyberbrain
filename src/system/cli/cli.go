package cli

import (
	"flag"
	"os"

	"github.com/voodooEntity/archivist"
)

type Args struct {
	ProjectPath string
	Filter      string
	Command     string
	Verbose     bool
}

var Data Args

func ParseArgs() {
	//os.Args = []string{"cyberbrain", "run"} //### debug shit
	if 2 > len(os.Args) {
		PrintHelpText()
	}
	command := os.Args[1]
	os.Args = os.Args[1:]

	if 1 < len(os.Args) {
		if ok := os.Args[1]; ok == "help" {
			PrintHelpText()

		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		archivist.Error("could not retrieve cwd: ", err.Error())
		os.Exit(1)
	}

	// the project path for plugin compiling and config checking
	var projectPath string
	flag.StringVar(&projectPath, "project", cwd+"/", "Your project directory path")

	// filter var
	var filter string
	flag.StringVar(&filter, "filter", "", "Filter string to be applied")

	// verbose output flag
	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "verbose", false, "Verbose logging flag")

	flag.Parse()

	Data = Args{
		ProjectPath: projectPath,
		Filter:      filter,
		Verbose:     verboseFlag,
		Command:     command,
	}
}

func PrintHelpText() {
	helpText := "Cyberbrain Help:\n" +
		"> Cyberbrain is an smart data processing architecture. Using this command you can \n" +
		"  build or test plugins for your cyberbrain application - or start the cyberbrain \n" +
		"  application to run. For further information please check the docs in the reps\n" +
		"   https://github.com/voodooEntity/go-cyberbrain readme.\n\n" +
		"  The cli interface has to be called like\n" +
		"  cyberbrain <command> [+<args>]\n\n" +
		"  Commands: \n" +
		"    build\n" +
		"     - Builds given plugin source codes to cyberbrain plugins. \n" +
		"     - Args: -project\n\n" +
		"    test\n" +
		"     - Tests plugins in a given project. This is meant to test if a plugin runs with\n" +
		"       cyberbrain. To test plugins you need to provide a test.json in the plugin source\n" +
		"       directory. Testing plugins will also result in building the plugins. This is necessary\n" +
		"       for cyberbrain to test-run them. If you prefer to only run a specific test rather than\n" +
		"       testing all plugins you can use the filter param which compares a string to the available\n" +
		"       plugins names.\n" +
		"     - Args: -project -filter\n\n" +
		"    run\n" +
		"     - Args: -project\n\n" +
		"    help\n" +
		"     - Prints the help -text your are just reading.\n\n" +
		"    version\n" +
		"     - Prints the current deployed version of cyberbrain\n\n" +
		"  Args explained: \n" +
		"    -project \"/path/to/projec/\"   | Path your project. This should hold the source codes for\n" +
		"                                      your plugins and also the project config. Plugins will be\n" +
		"                                      build to PROJECTPATH/plugins. The default value will be the\n" +
		"                                      current working directory.\n\n" +
		"    --verbose                     | Activates verbose output mode\n"
	archivist.Info(helpText)
	os.Exit(1)
}
