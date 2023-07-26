package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/onflow/flow-go-sdk/access/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/viper"
)

const (
	DefaultServerPort   = 8000
	DefaultFlowNetwork  = "emulator"
	DefaultAdminAccount = "0xf8d6e0586b0a20c7"
	PrefixForge4Flow    = "forge4flow"
	ConfigFileName      = "forge4flow.yaml"
)

type Forge4FlowConfig struct {
	ServerPort   string `mapstructure:"serverPort"`
	LogLevel     int8   `mapstructure:"logLevel"`
	FlowNetwork  string `mapstructure:"flowNetwork"`
	AdminAccount string `mapstructure:"adminAccount"`
}

var AppConfig *Forge4FlowConfig

func LoadAppConfig(cfgFile string) {
	cfg := ConfigFileName
	if cfgFile != "" {
		cfg = cfgFile
	}

	viper.SetConfigFile(cfg)
	viper.SetDefault("serverPort", DefaultServerPort)
	viper.SetDefault("logLevel", zerolog.DebugLevel)
	viper.SetDefault("flowNetwork", DefaultFlowNetwork)
	viper.SetDefault("adminAccount", DefaultAdminAccount)

	// If config file exists, use it
	_, err := os.ReadFile(cfg)
	if err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Msg("Error while reading forge4flow.yaml. Shutting down.")
		}
	} else {
		if os.IsNotExist(err) {
			log.Info().Msg("Could not find forge4flow.yaml. Attempting to use environment variables.")
		} else {
			log.Fatal().Err(err).Msg("Error while reading forge4flow.yaml. Shutting down.")
		}
	}

	var config Forge4FlowConfig
	// If available, use env vars for config
	for _, fieldName := range getFlattenedStructFields(reflect.TypeOf(config)) {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", PrefixForge4Flow, strings.ReplaceAll(fieldName, ".", "_")))
		envVar := os.Getenv(envKey)
		if envVar != "" {
			viper.Set(fieldName, envVar)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal().Err(err).Msg("Error while creating config. Shutting down.")
	}

	// Configure logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.Level(config.LogLevel))
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Check the selected network host value and assign it to the appropriate constant
	flowNetwork := config.FlowNetwork
	switch flowNetwork {
	case "emulator":
		config.FlowNetwork = http.EmulatorHost
	case "testnet":
		config.FlowNetwork = http.TestnetHost
	case "mainnet":
		config.FlowNetwork = http.MainnetHost
	default:
		log.Fatal().Msgf("Invalid flowNetwork parameter: %s - valid options are: emulator, testnet, mainnet", flowNetwork)
	}

	AppConfig = &config
}

func getFlattenedStructFields(t reflect.Type) []string {
	return getFlattenedStructFieldsHelper(t, []string{})
}

func getFlattenedStructFieldsHelper(t reflect.Type, prefixes []string) []string {
	unwrappedT := t
	if t.Kind() == reflect.Pointer {
		unwrappedT = t.Elem()
	}

	flattenedFields := make([]string, 0)
	for i := 0; i < unwrappedT.NumField(); i++ {
		field := unwrappedT.Field(i)
		fieldName := field.Tag.Get("mapstructure")
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			flattenedFields = append(flattenedFields, getFlattenedStructFieldsHelper(field.Type, append(prefixes, fieldName))...)
		default:
			flattenedField := fieldName
			if len(prefixes) > 0 {
				flattenedField = fmt.Sprintf("%s.%s", strings.Join(prefixes, "."), fieldName)
			}
			flattenedFields = append(flattenedFields, flattenedField)
		}
	}

	return flattenedFields
}