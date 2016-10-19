package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ogier/pflag"
	"github.com/spf13/viper"
)

var (
	config, output, file, port, pw, query, un, urlAddr string
	inter                                              bool
)

const ERROR_USER_INPUT int = 0x1

const DEFAULT_CONFIG_NAME string = "config"
const API_SEARCH_ENDPOINT string = "/services/search/jobs/export"

func main() {
	if !inter {
		// check for the query
		if query == "" {
			fmt.Fprintf(os.Stdout, "No query provided: execution is complete\n")
			return
		}
		requestVals := url.Values{}
		requestVals.Add("search", "search "+query)
		requestVals.Add("output_mode", output)
		requestURL := fmt.Sprintf("%s:%s%s", strings.TrimSuffix(urlAddr, "/"), port, API_SEARCH_ENDPOINT)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		request, _ := http.NewRequest("POST", requestURL, bytes.NewBufferString(requestVals.Encode()))
		request.SetBasicAuth(un, pw)
		resp, err := client.Do(request)
		checkError(err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		checkError(err)

		if file != "" {
			ioutil.WriteFile(file, body, 0644)
		} else {
			// output file not specified; print to console
			fmt.Fprintln(os.Stdout, string(body))
		}
	} else {
		// run the console in interactive mode
		consoleRun()
	}
	return
}

/* Check for an error and panic if it exists
 * @params - e: a error (usually returned from some function) to check
 * @return - none
 */
func checkError(e error) {
	if e != nil {
		panic(e)
	}

	return
}

/* Prints the 'help' output for use in the console; this should print any time
 * the user inputs an incorrect command
 * @params - none
 * @return - none
 */
func consoleHelp() {
	fmt.Println("Ephor Console Commands:")
	fmt.Println(" config filename     Reloads the console with the specified config file.")
	fmt.Println(" exit                Exits the Ephor console.")
	fmt.Println(" file filename       Writes the output to the specified file.")
	fmt.Println(" help                Prints this help message.")
	fmt.Println(" output filetype     Changes the output file type to the specified type (XML/JSON/CSV).")
	fmt.Println(" query querystring   Runs the specified query and outputs the results.")
	return
}

/* Runs the console for interactive mode. This function loops, executing each
 * command as it is received, until it breaks or receives an 'exit' command.
 * @params - none
 * @return - none
 */
func consoleRun() {
	// create the client connection
	fmt.Printf("Connecting to splunk instance '%s'... ", urlAddr)
	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// client := &http.Client{Transport: tr}
	fmt.Println("connected.")
	for {
		fmt.Print("ephor > ")
		r := bufio.NewReader(os.Stdin)
		in, _ := r.ReadString('\n')
		consoleCmd := strings.Split(strings.TrimSuffix(in, "\n"), " ")
		switch consoleCmd[0] {
		case "config":

		case "exit":
			return
		case "file":
			if len(consoleCmd) != 2 {
				consoleUsage()
			} else {
				file = consoleCmd[1]
				fmt.Println("Data will now be output to " + file)
			}
		case "output":
			if len(consoleCmd) != 2 || !validOutputType(consoleCmd[1]) {
				consoleUsage()
			} else {
				output = consoleCmd[1]
				fmt.Println("Data will now be output as " + strings.ToUpper(output))
			}
		case "query":
			if file == "" {
				fmt.Fprint(os.Stdout, "Somthing was output")
			} else {
				ioutil.WriteFile(file, []byte("Something was output"), 0644)
			}
		default:
			consoleHelp()
		}
	}
}

/* Print a usage message and then print the consoleHelp message
 * @params - none
 * @return - none
 */
func consoleUsage() {
	fmt.Println("Incorrect command usage - see help below")
	consoleHelp()
}

/* This init function handles the parsing of the commandline flags and the
 * import of the configuration information from the cofig files.
 * @params - none
 * @return - none
 */
func init() {
	// parse the commandline flags
	pflag.StringVarP(&config, "config", "c", "", "The path to a config file if it is not in your home or execution directories")
	pflag.StringVarP(&file, "file", "f", "", "The path to a file for writing query results")
	pflag.StringVarP(&output, "output", "o", "", "The data type for the results (valid values are XML (default), JSON, and CSV)")
	pflag.StringVarP(&port, "port", "p", "", "The port used by the splunk server instance (if not provided, defaults to 8089)")
	pflag.StringVarP(&query, "query", "q", "", "The search query to execute (required if not using interactive mode)")
	pflag.StringVarP(&urlAddr, "url", "r", "", "The URL of the splunk server instance (required if not using a config file)")
	pflag.StringVarP(&un, "username", "u", "", "The username of a splunk account (required if not using a config file)")
	pflag.StringVarP(&pw, "password", "w", "", "The password to a splunk account (required if not using a config file)")
	pflag.BoolVarP(&inter, "interactive", "i", false, "Use the interactive console for making multiple queries")
	pflag.Parse()

	// get the current user context
	u, _ := user.Current()

	// setup viper config manager
	viper.SetConfigName(DEFAULT_CONFIG_NAME)
	viper.AddConfigPath(u.HomeDir)
	viper.AddConfigPath(".")
	// if a config file was passed in, parse it here
	if config != "" {
		p, _ := filepath.Abs(filepath.Dir(config))
		viper.AddConfigPath(p)

		e := viper.ReadInConfig()
		if e != nil {
			fmt.Fprintf(os.Stderr, "ERROR: a could not read a valid config file: %s\n", e.Error())
			os.Exit(ERROR_USER_INPUT)
		}
		if un == "" && viper.IsSet("username") {
			un = viper.GetString("username")
		}
		if pw == "" && viper.IsSet("password") {
			pw = viper.GetString("password")
		}
		if urlAddr == "" && viper.IsSet("url") {
			urlAddr = viper.GetString("url")
		}
	}
	// make sure minimal information has been parsed
	if un == "" || pw == "" || urlAddr == "" {
		fmt.Fprintf(os.Stderr, "ERROR: one or more pieces of required configuration information were not provided\n")
		os.Exit(ERROR_USER_INPUT)
	}

	// if not provided, set the port to the splunk default :8089
	if port == "" {
		port = "8089"
	}

	if output == "" {
		output = "xml" // splunk docs note the default datatype is XML
	}
	if !validOutputType(output) {
		fmt.Fprintf(os.Stderr, "ERROR: output format ('%s') is invalid; valid values are XML (default), JSON, and CSV\n", output)
		output = "xml" // splunk docs note the default datatype is XML
	}
}

/* Returns true if the provided string is equal to "XML", "JSON", or "CSV",
 * ignoring character case
 * @params - t: a string value to validate
 * @return - a boolean indicating whether the provided sting is a valid
 * 			 output type
 */
func validOutputType(t string) bool {
	return strings.EqualFold("XML", t) || strings.EqualFold("JSON", t) || strings.EqualFold("CSV", t)
}
