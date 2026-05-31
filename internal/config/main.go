package config

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type BigIpExporterParameter struct {
	Config        *string
	Listen        *string
	ScrapeTimeout *int
	TLSTimeout    *int
	TLSInsecure   *bool
}
type (
	Target string
)
type BigIpExporterConfig struct {
	AuthKeys      AuthKeys
	Listen        string
	ScrapeTimeout int
	TLSTimeout    int
	TLSInsecure   bool
}

type AuthKeys map[Target]TargetAuth

type TargetAuth struct {
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
}

var (
	parameter = BigIpExporterParameter{
		Config:        flag.String("config", "bigip-config.yaml", "file containing the authentication map to use when connecting to a F5 device"),
		Listen:        flag.String("listen", ":11000", "address to listen on"),
		ScrapeTimeout: flag.Int("scrape-timeout", 30, "max seconds to allow a scrape to take"),
		TLSTimeout:    flag.Int("https-timeout", 10, "TLS Handshake timeout in seconds"),
		TLSInsecure:   flag.Bool("insecure", false, "Skip TLS certificate verification"),
	}

	savedConfig *BigIpExporterConfig
)

func Init() error {
	if savedConfig != nil {
		return nil
	}
	return ReInit()
}

func MustReInit() {
	if err := ReInit(); err != nil {
		log.Fatalf("config.ReInit failed: %+v", err)
	}
}

func ReInit() error {
	flag.Parse()

	config, err := os.ReadFile(*parameter.Config)
	if err != nil {
		log.Fatalf("Failed to read API authentication map file: %v", err)
		return err
	}
	var AuthKeys AuthKeys
	if err := yaml.Unmarshal(config, &AuthKeys); err != nil {
		log.Fatalf("Failed to parse API authentication map file: %v", err)
		return err
	}
	savedConfig = &BigIpExporterConfig{
		AuthKeys:      AuthKeys,
		Listen:        *parameter.Listen,
		ScrapeTimeout: *parameter.ScrapeTimeout,
		TLSTimeout:    *parameter.TLSTimeout,
		TLSInsecure:   *parameter.TLSInsecure,
	}
	log.Printf("Loaded configuration from %q", *parameter.Config)

	return nil
}

func GetConfig() BigIpExporterConfig {
	return *savedConfig
}
