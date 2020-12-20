package main

import (
	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"regexp"

	"github.com/sh0rez/luxtronik2-exporter/pkg/luxtronik"
)

// Version of the app. To be set by ldflags
var Version = "dev"

// Config holds the configuration structure
type Config struct {
	Verbose bool   `flag:"verbose" short:"v" help:"Show debug logs"`
	Address string `flag:"address" short:"a" help:"IP or hostname of the heatpump"`
	Filters luxtronik.Filters
	Mutes   []struct {
		Domain string
		Field  string
	}
}

func main() {
	cmd := &cobra.Command{
		Use:     "luxtronik2-exporter",
		Short:   "Expose metrics from luxtronik2 based heatpumps in Prometheus format.",
		Version: Version,
	}

	// file config
	viper.SetConfigName("lux")
	viper.AddConfigPath(".")

	// flag config
	for _, s := range structs.Fields(Config{}) {
		if s.Tag("flag") != "" {
			cmd.Flags().StringP(s.Tag("flag"), s.Tag("short"), s.Tag("default"), s.Tag("help"))
		}
	}
	viper.BindPFlags(cmd.Flags())

	// env config
	viper.SetEnvPrefix("lux")
	viper.AutomaticEnv()

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// unmarshal sources
		var config Config
		if err := viper.ReadInConfig(); err != nil {
			log.WithField("err", err).Fatal("Error getting config from sources")
		}
		if err := viper.Unmarshal(&config); err != nil {
			log.Error(err)
			log.WithField("err", err).Fatal("invalid config")
		}

		log.SetLevel(log.InfoLevel)
		if config.Verbose {
			log.SetLevel(log.DebugLevel)
		}

		run(&config)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func run(config *Config) {
	mutes = make(MuteList, len(config.Mutes))
	for i, m := range config.Mutes {
		mutes[i] = Mute{
			domain: regexp.MustCompile(m.Domain),
			field:  regexp.MustCompile(m.Field),
		}
	}

	// connect to the heatpump
	lux := luxtronik.Connect(config.Address, config.Filters)

	// register update handler, gets called by the update routine
	// updates changed metrics
	lux.OnUpdate = func(new []luxtronik.Location) {
	}

	// serve the /metrics endpoint
	log.Fatalln(http.ListenAndServe(":2112", nil))
}
