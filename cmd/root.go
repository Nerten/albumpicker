package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var osExit = os.Exit

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "albumpicker",
	Short: "Utility for randomly selecting and copying FLAC albums",
	Long: `A console utility written in Go for randomly selecting and copying music albums
in FLAC format. The utility allows specifying the number of albums to copy,
provides functionality to remove image metadata from FLAC files, and
copies album covers with size optimization.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		osExit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/albumpicker/config.yaml)")

	rootCmd.PersistentFlags().StringP("source", "s", "", "source directory with FLAC albums")
	rootCmd.PersistentFlags().StringP("destination", "d", "", "destination directory for copied albums")
	rootCmd.PersistentFlags().Int("height", 0, "cover image height in pixels (default 240)")
	rootCmd.PersistentFlags().String("cover-name", "", "output cover file name")

	m := map[string]string{
		"source":                "source",
		"destination":           "destination",
		"cover_height":          "height",
		"output_cover_filename": "cover-name",
	}
	for key, name := range m {
		err := viper.BindPFlag(key, rootCmd.PersistentFlags().Lookup(name))
		if err != nil {
			panic(err.Error())
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("source", "")
	viper.SetDefault("destination", "")
	viper.SetDefault("albums_count", 10)
	viper.SetDefault("cover_filenames", []string{"album.jpg", "album.png", "cover.jpg", "cover.png"})
	viper.SetDefault("output_cover_filename", "cover.jpg")
	viper.SetDefault("cover_height", 240)

	if cfgFile != "" {
		// use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %s\n", err)
			os.Exit(1)
		}
		// search config in home directory
		configDir := filepath.Join(home, ".config", "albumpicker")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creation config directory: %s\n", err)
				os.Exit(1)
			}
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// save default config if not exists
		if err := viper.SafeWriteConfig(); err != nil {
			var configFileAlreadyExistsError viper.ConfigFileAlreadyExistsError
			if !errors.As(err, &configFileAlreadyExistsError) {
				fmt.Fprintf(os.Stderr, "Warning: Could not create default config: %s\n", err)
			}
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stdout, "Config file: %s\n", viper.ConfigFileUsed())
	}
}
