# ephor
Ephor is a splunk-query CLI implemented in Go. It is designed to allow queries to be run from the commandline, enabling splunk querying to be built into scripts and other automation.

## Configuration
Ephor expects a minimum of three pieces of information.
1. A splunk instance URL
2. A splunk username
3. A splunk password

These three pieces of information can be passed in via the commandline, using the appropriate flags, or via a simple configuration file.

Sample TOML config file:
  ```
  username="admin"
  password="changeme"
  url="https://my.splunk.url.domain.com/"
  ```
Sample JSON config file:
  ```
  {
    "username": "admin",
    "password": "changeme",
    "url": "https://my.splunk.url.domain.com/"
  }
  ```
The information can be in any order, as long as it exists. More configuration options will likely be available in the future.

## Console
The Ephor 'console' is a simple mode where queries can be excuted quickly in sequence, without having to re-provide credentials and setting query options. The console requires the same basic information (see Configuration above), and it can be passed in via a config file or the commandline, as with the one-time mode. Config information can be changed for subsequent queries by loading a new config file.
