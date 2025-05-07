// config.go
package config

import (
	// import standard Go libraries
	"encoding/json" // decoding json to go
	"fmt"           // printing
	"os"            // for os file access
	"path/filepath" // pilepath without str interpolation
)

// package-wide constants
const configFileName = ".gatorconfig.json"

// . = makes it hidden on system! standard gopher practice for config files!

// config struct
type Config struct {
	URL  *string `json:"db_url"`            // url of DB
	Name *string `json:"current_user_name"` // username
}

// String method to format the Config struct when printing
func (c Config) String() string {
	url := "" // default
	// if ptr is not nil
	if c.URL != nil {
		url = *c.URL // set url to config's value
	}

	name := "" // default
	// if ptr is not nil
	if c.Name != nil {
		name = *c.Name // set url to config's value
	}

	// pre-formatted output of config file!
	return fmt.Sprintf("Config{URL: %q, Name: %q}", url, name)
}

// read gatorconfig & return struct
func Read() (Config, error) {
	// get config file path using helper
	configPath, err := getConfigPath()

	// configpath check
	if err != nil {
		return Config{}, fmt.Errorf("error getting filepath from helper: %w", err)
	}

	// read raw json data
	file, err := os.Open(configPath)

	// read check
	if err != nil {
		// check if the file doesn't exist
		if os.IsNotExist(err) {
			// doesn't exist? OK! just create an empty one
			return Config{}, nil
		} else {
			// otherwise some error occured, handle it
			return Config{}, fmt.Errorf("error reading config file: %w", err)
		}
	}

	defer file.Close() // end reading after return for filesafety

	// create nil slice for external data use
	var cfg Config

	// decode (unmarshal less efficient, negligible for config) raw json data to Go struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	// reminder on unmarhsal: err := json.Unmarshal(file, &cfg)

	// decode check
	if err != nil {
		return cfg, fmt.Errorf("error decoding config file: %w", err)
	}

	// return Go config
	return cfg, nil
}

// set username in gatorconfig json
func SetUser(userName string) (Config, error) {
	// read the config file
	cfg, err := Read() // we already built this function! no need to decode etc!

	// read check
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	// update the username field
	// *cfg.Name = "gator bites!" - this will cause Go panic - field is still nil!
	cfg.Name = &userName // safe way to update field

	// write the updated config file using helper
	err = write(cfg)

	// write check
	if err != nil {
		return cfg, fmt.Errorf("error writing updated config file: %w", err)
	}

	// return Json config
	return cfg, nil
}

// get config file path helper function
func getConfigPath() (string, error) {
	// get home path
	homePath, err := os.UserHomeDir()

	// homepath check
	if err != nil {
		return "", fmt.Errorf("error getting home dir: %w", err)
	}

	// gatorconfig filepath using home dir (don't interpolate raw strs!)
	configPath := filepath.Join(homePath, configFileName)
	// NOTE: filepath.Join() handles /'s for you

	// return the config filepath
	return configPath, nil
}

// write config file helper function
func write(cfg Config) error {
	// get config file path using helper
	configPath, err := getConfigPath()

	// configpath check
	if err != nil {
		return fmt.Errorf("error getting filepath from helper: %w", err)
	}

	// marshal the struct back to json (marshalindent prettifies it with newlines!)
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	// prefix = "" ie nothing , indent = "  " ie 2 spaces

	// marhsalindent check
	if err != nil {
		return fmt.Errorf("error marshalling cfg back to json: %w", err)
	}

	// write to fille the updated json
	err = os.WriteFile(configPath, jsonData, 0644) // writefile only returns ERR
	// 0644 = octal persmission code for owner can read & write

	// writefile check
	if err != nil {
		return fmt.Errorf("error writing updated json: %w", err)
	}

	// return success
	return nil
}
