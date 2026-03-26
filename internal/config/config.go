package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	logger "zcli/internal/logging"
	"zcli/internal/zosmf"
)

// ZcliConfig represents the top-level zcli.json structure.
type ZcliConfig struct {
	Profiles map[string]ProfileEntry `json:"profiles"`
	Defaults []map[string]any        `json:"defaults"`
}

// ProfileEntry represents a single profile in zcli.json.
type ProfileEntry struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// Global configuration instance.
var Conf *ZcliConfig

// ConfigDir returns the path to the zcli config directory.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("Failed to get user home directory", "error", err)
		fmt.Fprintf(os.Stderr, "ZCLI-CONFIG-001S Unable to determine home directory: %v\n", err)
		os.Exit(16)
	}
	return filepath.Join(home, ".config", "zcli")
}

// LoadConfig reads zcli.json from the config directory.
func LoadConfig() *ZcliConfig {
	configPath := filepath.Join(ConfigDir(), "zcli.json")
	logger.Log.Info("Loading zcli configuration", "file", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Log.Error("Failed reading configuration", "file", configPath, "error", err)
		fmt.Fprintf(os.Stderr, "ZCLI-CONFIG-003S Unable to read %s, terminating rc = 16\n", configPath)
		os.Exit(16)
	}

	conf := &ZcliConfig{}
	if err := json.Unmarshal(data, conf); err != nil {
		logger.Log.Error("Failed decoding config file", "error", err)
		fmt.Fprintf(os.Stderr, "ZCLI-CONFIG-004S Unable to parse %s: %v\n", configPath, err)
		os.Exit(16)
	}

	Conf = conf
	return conf
}

// GetDefaultProfileName returns the name of the default profile for the given type.
func GetDefaultProfileName(profileType string) (string, error) {
	if Conf == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	for _, entry := range Conf.Defaults {
		if profilesRaw, ok := entry["profiles"]; ok {
			if profiles, ok := profilesRaw.([]interface{}); ok {
				for _, p := range profiles {
					if pm, ok := p.(map[string]interface{}); ok {
						if name, ok := pm[profileType]; ok {
							return fmt.Sprintf("%v", name), nil
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("no default profile found for type %s", profileType)
}

// GetZcliProperty returns a zcli property from the defaults section.
func GetZcliProperty(propName string) string {
	if Conf == nil {
		return ""
	}
	for _, entry := range Conf.Defaults {
		if zcliRaw, ok := entry["zcli"]; ok {
			if zcli, ok := zcliRaw.(map[string]interface{}); ok {
				if propsRaw, ok := zcli["properties"]; ok {
					if props, ok := propsRaw.(map[string]interface{}); ok {
						if val, ok := props[propName]; ok {
							return fmt.Sprintf("%v", val)
						}
					}
				}
			}
		}
	}
	return ""
}

// GetProfileProperty returns a specific property from a named profile.
func GetProfileProperty(profileName, profileType, key string) string {
	if Conf == nil {
		return ""
	}

	profile, ok := Conf.Profiles[profileName]
	if !ok {
		return ""
	}
	if profile.Type != profileType {
		return ""
	}
	if val, ok := profile.Properties[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// GetTSOProperty returns a TSO property from the defaults section.
func GetTSOProperty(propName string) string {
	if Conf == nil {
		return ""
	}
	for _, entry := range Conf.Defaults {
		if tsoRaw, ok := entry["tso"]; ok {
			if tso, ok := tsoRaw.(map[string]interface{}); ok {
				if propsRaw, ok := tso["properties"]; ok {
					if props, ok := propsRaw.(map[string]interface{}); ok {
						if val, ok := props[propName]; ok {
							return fmt.Sprintf("%v", val)
						}
					}
				}
			}
		}
	}
	return ""
}

// GetSoftwareProperty returns a software property from the defaults section.
func GetSoftwareProperty(propName string) string {
	if Conf == nil {
		return ""
	}
	for _, entry := range Conf.Defaults {
		if swRaw, ok := entry["software"]; ok {
			if sw, ok := swRaw.(map[string]interface{}); ok {
				if propsRaw, ok := sw["properties"]; ok {
					if props, ok := propsRaw.(map[string]interface{}); ok {
						if val, ok := props[propName]; ok {
							return fmt.Sprintf("%v", val)
						}
					}
				}
			}
		}
	}
	return ""
}

// ProfileData holds all connection parameters resolved from a profile.
type ProfileData struct {
	ProfileName string
	Protocol    string
	HostName    string
	Port        string
	Encoding    string
	User        string
	Password    string
	CertPath    string
	Verify      bool
}

// ResolveProfile resolves a profile name into connection data.
// If profileName is empty, the default zosmf profile is used.
func ResolveProfile(profileName string, verify bool) (*ProfileData, error) {
	if Conf == nil {
		LoadConfig()
	}

	if profileName == "" {
		def, err := GetDefaultProfileName("zosmf")
		if err != nil {
			return nil, fmt.Errorf("ZCLI-MAIN-001S No default profile and --profile-name is empty: %w", err)
		}
		profileName = def
	}

	protocol := GetProfileProperty(profileName, "zosmf", "protocol")
	if protocol == "" {
		protocol = "https"
	}

	hostName := GetProfileProperty(profileName, "zosmf", "host")
	if hostName == "" {
		return nil, fmt.Errorf("ZCLI-MAIN-002S Property \"host\" missing in profile %s", profileName)
	}

	port := GetProfileProperty(profileName, "zosmf", "port")
	if port == "" {
		port = "443"
	}

	encoding := GetProfileProperty(profileName, "zosmf", "encoding")
	if encoding == "" {
		encoding = "IBM-1047"
	}

	user := GetProfileProperty(profileName, "zosmf", "user")
	if user == "" {
		return nil, fmt.Errorf("ZCLI-MAIN-002S Property \"user\" missing in profile %s", profileName)
	}

	password := GetProfileProperty(profileName, "zosmf", "password")
	if password == "" {
		return nil, fmt.Errorf("ZCLI-MAIN-002S Property \"password\" missing in profile %s", profileName)
	}

	certPath := GetZcliProperty("cert_path")

	return &ProfileData{
		ProfileName: profileName,
		Protocol:    protocol,
		HostName:    hostName,
		Port:        port,
		Encoding:    encoding,
		User:        user,
		Password:    password,
		CertPath:    certPath,
		Verify:      verify,
	}, nil
}

// NewZosmfClient creates a z/OSMF client from profile data.
func (p *ProfileData) NewZosmfClient() *zosmf.Client {
	return zosmf.NewClient(p.Protocol, p.HostName, p.Port, p.User, p.Password, p.Verify)
}

func init() {
	LoadConfig()
}
