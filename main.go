package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ogier/pflag"
	"github.com/spf13/viper"
)

var (
	file, pw, un, url string
	inter             bool
)

const DEFAULT_CONFIG_NAME string = "config"
const DEFAULT_CONFIG_DIR string = "/etc/ephor/"

func main() {
	if !inter {

	} else {
		// run the console for interactive mode
		runConsole()
	}

	return
}

func init() {
	// parse the commandline flags
	pflag.StringVarP(&file, "config", "c", "", "The path to a config file.")
	pflag.BoolVarP(&inter, "interactive", "i", false, "Use interactive mode.")
	pflag.Parse()

	// get the current user context
	u, _ := user.Current()

	// setup viper config manager
	viper.SetConfigName(DEFAULT_CONFIG_NAME)
	viper.AddConfigPath(u.HomeDir)
	viper.AddConfigPath(".")
	// if the config file was passed via the commandline, add its path here
	if file != "" {
		p, _ := filepath.Abs(filepath.Dir(file))
		viper.AddConfigPath(p)
	}
	e := viper.ReadInConfig()
	if e != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not read a valid config file: %s\n", e.Error())
		os.Exit(1)
	}
	un = viper.GetString("username")
	pw = viper.GetString("password")
	url = viper.GetString("url")
}

/* Runs the console for interactive mode. This function loops, executing each
 * command as it is received, until it breaks or receives an 'exit' command.
 * @params - none
 * @return - none
 */
func runConsole() {
	for {
		fmt.Print("ephor > ")
		r := bufio.NewReader(os.Stdin)
		in, _ := r.ReadString('\n')
		if in == "exit\n" {
			os.Exit(0)
		} else {
			// process the command input
			fmt.Print(in)
		}
	}
}
