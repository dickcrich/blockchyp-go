package blockchyp

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// ConfigSettings contains configuration options for the CLI.
type ConfigSettings struct {
	APIKey          string `json:"apiKey"`
	BearerToken     string `json:"bearerToken"`
	SigningKey      string `json:"signingKey"`
	GatewayHost     string `json:"gateway"`
	TestGatewayHost string `json:"testGateway"`
	Secure          bool   `json:"https"`
	RouteCacheTTL   int    `json:"routeCacheTTL"`
	GatewayTimeout  int    `json:"gatewayTimeout"`
	TerminalTimeout int    `json:"terminalTimeout"`
}

// CommandLineArguments contains arguments which are passed in at runtime.
type CommandLineArguments struct {
	Type            string `arg:"type"`
	ManualEntry     bool   `arg:"manual"`
	ConfigFile      string `arg:"f"`
	GatewayHost     string `arg:"gateway"`
	TestGatewayHost string `arg:"testGateway"`
	Test            bool   `arg:"test"`
	APIKey          string `arg:"apiKey"`
	BearerToken     string `arg:"bearerToken"`
	SigningKey      string `arg:"signingKey"`
	TransactionRef  string `arg:"txRef"`
	Description     string `arg:"desc"`
	TerminalName    string `arg:"terminal"`
	Token           string `arg:"token"`
	Amount          string `arg:"amount"`
	PromptForTip    bool   `arg:"promptForTip"`
	TipAmount       string `arg:"tip"`
	TaxAmount       string `arg:"tax"`
	CurrencyCode    string `arg:"currency"`
	TransactionID   string `arg:"txId"`
	RouteCache      string `arg:"routeCache"`
	OutputFile      string `arg:"out"`
	SigFormat       string `arg:"sigFormat"`
	SigWidth        int    `arg:"sigWidth"`
	SigFile         string `arg:"sigFile"`
	HTTPS           bool   `arg:"secure"`
	Version         bool   `arg:"version"`
}

var defaultSettings = &ConfigSettings{
	GatewayHost:     "https://api.blockchyp.com",
	TestGatewayHost: "https://test.blockchyp.com",
	Secure:          true,
}

// LoadConfigSettings loads settings from the command line and/or the
// configuration file.
func LoadConfigSettings(args CommandLineArguments) (*ConfigSettings, error) {
	fileName := args.ConfigFile
	if fileName == "" {
		var configHome string

		if runtime.GOOS == "windows" {
			configHome = os.Getenv("userprofile")
		} else {
			configHome = os.Getenv("XDG_CONFIG_HOME")
			if configHome == "" {
				user, err := user.Current()
				if err != nil {
					return nil, err
				}
				configHome = user.HomeDir + "/.config"
			}
		}

		fileName = filepath.Join(configHome, ConfigDir, ConfigFile)
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if args.ConfigFile != "" {
			return nil, errors.New(fileName + " not found")
		}
		return defaultSettings, nil
	}

	b, err := ioutil.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	config := &ConfigSettings{}
	err = json.Unmarshal(b, config)

	return config, err
}
