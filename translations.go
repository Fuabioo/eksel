package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

var (
	translations = viper.New()
)

func t(key string, args ...interface{}) string {
	return fmt.Sprintf(translations.GetString(key), args...)
}

//go:embed translations.yml
var translationFile string

func init() {
	translations.SetConfigType("yaml")
	err := translations.ReadConfig(strings.NewReader(translationFile))
	if err != nil {
		log.Fatal("Failed to read translations file", err)
	}
}
