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
	"strconv"
	"strings"

	"github.com/ogier/pflag"
	"github.com/spf13/viper"
)

var (
	config, output, file, port, pw, query, un, urlAddr string
	inter                                              bool
)

const ERROR_USER_INPUT int = 0x1000

const DEFAULT_CONFIG_NAME string = "config"
const SINGLE_SEARCH_ENDPOINT string = "/services/search/jobs/export"

func main() {
	if !inter {
		resp := sendHTTPRequest()
		writeHTTPResponse(file, resp)
	} else {
		// run the console in interactive mode
		consoleRun()
	}
	return
}

/* Checks for an error and when there is an error, prints the message out to
 * stderr, then exits
 * @params - e: an error (usually returned from some function) to check
 * @return - none
 */
func checkErrorWithExit(e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, e.Error())
		os.Exit(1)
	}
}

/* Checks for an error and when there is an error, panics
 * @params - e: an error (usually returned from some function) to check
 * @return - none
 */
func checkErrorWithPanic(e error) {
	if e != nil {
		panic(e)
	}
}

/* Runs the console for interactive mode. This function loops, executing each
 * command as it is received, until it breaks or receives an 'exit' command.
 * @params - none
 * @return - none
 */
func consoleRun() {
	for {
		fmt.Print("ephor > ")
		r := bufio.NewReader(os.Stdin)
		in, _ := r.ReadString('\n')
		consoleCmd := strings.Split(strings.TrimSuffix(in, "\n"), " ")
		switch consoleCmd[0] {
		case "config":
			if len(consoleCmd) != 2 {
				printConsoleHelp()
				continue
			}
			fmt.Printf("Reloading config info from '%s' ", consoleCmd[1])
			e := loadConfig(consoleCmd[1])
			if e != nil {
				fmt.Printf("failed. Config info was not changed.")
			} else {
				fmt.Println("done.")
			}
		case "exit":
			return
		case "file":
			if len(consoleCmd) != 2 {
				printConsoleHelp()
				continue
			}
			file = consoleCmd[1]
			fmt.Printf("Data will now be output to '%s'.\n", file)
		case "output":
			if len(consoleCmd) != 2 {
				printConsoleHelp()
				continue
			}
			if !isValidOutputType(consoleCmd[1]) {
				printConsoleHelp()
			} else {
				output = consoleCmd[1]
				fmt.Printf("Data will now be output as %s.\n", strings.ToUpper(output))
			}
		case "port":
			if len(consoleCmd) != 2 {
				printConsoleHelp()
				continue
			}
			tmp := consoleCmd[1]
			if _, e := strconv.Atoi(tmp); e != nil {
				fmt.Printf("The provided port '%s' is not valid. Keeping previous value of %s.\n", tmp, port)
			} else {
				port = consoleCmd[1]
				fmt.Printf("Now using port %s to connect to '%s'.\n", port, urlAddr)
			}
		case "query":
			if len(consoleCmd) < 2 {
				printConsoleHelp()
				continue
			}
			query = strings.Join(consoleCmd[1:], " ")
			resp := sendHTTPRequest()
			writeHTTPResponse(file, resp)
		case "status":
			if len(consoleCmd) != 1 {
				printConsoleHelp()
				continue
			}
			printStatus()
		default:
			printConsoleHelp()
		}
	}
}

/* This init function handles the parsing of the commandline flags and the
 * import of the configuration information from the cofig files.
 * @params - none
 * @return - none
 */
func init() {
	// parse the commandline flags
	pflag.StringVarP(&config, "config", "c", "", "The path to a config file if it is not in your home or execution directories.")
	pflag.StringVarP(&file, "file", "f", "", "The path to a file for writing query results.")
	pflag.StringVarP(&output, "output", "o", "", "The data type for the results (valid values are XML (default), JSON, and CSV).")
	pflag.StringVarP(&port, "port", "p", "", "The port used by the splunk server instance (if not provided, defaults to 8089).")
	pflag.StringVarP(&query, "query", "q", "", "The search query to execute (required if not using interactive mode).")
	pflag.StringVarP(&urlAddr, "url", "r", "", "The URL of the splunk server instance (required if not using a config file).")
	pflag.StringVarP(&un, "username", "u", "", "The username of a splunk account (required if not using a config file).")
	pflag.StringVarP(&pw, "password", "w", "", "The password to a splunk account (required if not using a config file).")
	pflag.BoolVarP(&inter, "interactive", "i", false, "Use the interactive console for making multiple queries.")
	pflag.Parse()

	// parse the base info
	e := loadConfig(config)
	checkErrorWithExit(e)

	// if not provided, set the port to the splunk default :8089
	if port == "" {
		port = "8089"
	}

	if output == "" {
		output = "xml" // splunk docs note the default datatype is XML
	}
	if !isValidOutputType(output) {
		fmt.Fprintf(os.Stderr, "ERROR: output format ('%s') is invalid; valid values are XML (default), JSON, and CSV.\n", output)
		output = "xml" // splunk docs note the default datatype is XML
	}

	return
}

/* Returns true if the provided string is equal to "XML", "JSON", or "CSV",
 * ignoring the character case of the string.
 * @params - t: a string value to validate
 * @return - a boolean indicating whether the provided string is a valid
 * 			 output type
 */
func isValidOutputType(t string) bool {
	return strings.EqualFold("XML", t) || strings.EqualFold("JSON", t) || strings.EqualFold("CSV", t)
}

/* Loads desired information from the provided config file. Commandline args
 * take precedence over those passed in from a file at initialization, but
 * new values should take precedence when reloading the config from the console.
 * @params - f: a config filepath/filename
 * @return - an error indicating if something went wrong during processing
 */
func loadConfig(f string) error {
	// get the current user context
	u, _ := user.Current()

	if f != "" {
		viper.SetConfigName(strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)))
		p, _ := filepath.Abs(filepath.Dir(f))
		viper.AddConfigPath(p)
	} else {
		viper.SetConfigName(DEFAULT_CONFIG_NAME)
		viper.AddConfigPath(u.HomeDir)
		viper.AddConfigPath(".")
	}
	// ignore error return; bad config info will be caught later
	viper.ReadInConfig()
	if un == "" && viper.IsSet("username") {
		un = viper.GetString("username")
	}
	if pw == "" && viper.IsSet("password") {
		pw = viper.GetString("password")
	}
	if urlAddr == "" && viper.IsSet("url") {
		urlAddr = viper.GetString("url")
	}
	// make sure minimal information has been parsed
	if un == "" || pw == "" || urlAddr == "" {
		return fmt.Errorf("ERROR: one or more pieces of required configuration information were not provided.\n")
	}

	return nil
}

/* Prints the 'help' output for use in the console; this should print any time
 * the user inputs an incorrect command.
 * @params - none
 * @return - none
 */
func printConsoleHelp() {
	fmt.Println("Ephor Console Commands:")
	fmt.Println(" config filename     Reloads the console with the specified config file.")
	fmt.Println(" exit                Exits the Ephor console.")
	fmt.Println(" file filename       Writes the output to the specified file.")
	fmt.Println(" help                Prints this help message.")
	fmt.Println(" output filetype     Changes the output file type to the specified type (XML/JSON/CSV).")
	fmt.Println(" port number         Changes the port used to make a connection.")
	fmt.Println(" query querystring   Runs the specified query and outputs the results.")
	fmt.Println(" status              Prints out the current configuration information.")
	return
}

/* Prints the current state of the application parameter values. This function
 * is not accessible from the commandline.
 * @params - none
 * @return - none
 */
func printStatus() {
	// set the output stream - file or out
	o := ""
	if file != "" {
		o = file
	} else {
		o = "Terminal"
	}
	// obfuscate the PW by only printing the first 4 chars
	ob := fmt.Sprintf("%s%s", pw[:4], strings.Repeat("*", len(pw[4:])))
	fmt.Printf("User Info:\n Username: %s\n Password: %s\nInstance Info:\n URL: %s\n Port: %s\nOutput Info:\n Output Type: %s\n Output Location: %s\n", un, ob, urlAddr, port, output, o)
}

/* Sends an HTTP request with the set parameters to the splunk server.
 * @params - none
 * @return - a byte-slice containing the response data
 */
func sendHTTPRequest() []byte {
	// check if the query was sent
	if query == "" {
		fmt.Fprintf(os.Stdout, "No query provided: execution is complete.\n")
		return nil
	}
	requestVals := url.Values{}
	// trim the quotes around querystrings from the commandline
	// querystrings used in the API are expected to be of the form 'search index=...'
	requestVals.Add("search", fmt.Sprintf("search %s", query))
	requestVals.Add("output_mode", output)
	requestVals.Add("exec_mode", "oneshot")
	fmt.Println(query)
	fmt.Println(requestVals)
	requestURL := fmt.Sprintf("%s:%s%s", strings.TrimSuffix(urlAddr, "/"), port, SINGLE_SEARCH_ENDPOINT)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	request, _ := http.NewRequest("POST", requestURL, bytes.NewBufferString(requestVals.Encode()))
	request.SetBasicAuth(un, pw)
	resp, err := client.Do(request)
	checkErrorWithPanic(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErrorWithPanic(err)

	return body
}

/* if the filename is provided, writes the output to file; if not, writes
 * output to Stdout. A write will not be attempted if the byte-slice param
 * is nil.
 * @params - f: a filename string
 *			 o: a byte-slice of output data
 * @return - none
 */
func writeHTTPResponse(f string, o []byte) {
	if o != nil {
		if f != "" {
			ioutil.WriteFile(f, o, 0644)
		} else {
			fmt.Fprintln(os.Stdout, string(o))
		}
	}
}
