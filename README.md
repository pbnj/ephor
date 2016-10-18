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

## Features
+ easy to configure set-up
+ the ability to quickly change credentials/roles by using a different config file
+ an 'interactive mode' where the user can execute multiple queries in a row (coming soon!)
