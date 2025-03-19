package main

import (
	"github.com/nerten/albumpicker/internal/command"
	"github.com/nerten/albumpicker/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.1.0"
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "albumpicker",
	Short: "Utility for randomly selecting and copying FLAC albums",
	Long: `A console utility written in Go for randomly selecting and copying music albums
in FLAC format. The utility allows specifying the number of albums to copy,
provides functionality to remove image metadata from FLAC files, and
copies album covers with size optimization.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/albumpicker/config.yaml)")

	rootCmd.PersistentFlags().StringP("source", "s", "", "Source directory with FLAC albums")
	rootCmd.PersistentFlags().StringP("destination", "d", "", "Destination directory for copied albums")
	rootCmd.PersistentFlags().Int("height", 0, "Cover image height in pixels (default 240)")
	rootCmd.PersistentFlags().String("cover-name", "", "Output cover file name")

	viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
	viper.BindPFlag("destination", rootCmd.PersistentFlags().Lookup("destination"))
	viper.BindPFlag("cover_height", rootCmd.PersistentFlags().Lookup("height"))
	viper.BindPFlag("output_cover_filename", rootCmd.PersistentFlags().Lookup("cover-name"))
}

func main() {
	rootCmd.Version = version
	config.InitConfig(cfgFile)

	rootCmd.AddCommand(command.PickCmd)
	rootCmd.AddCommand(command.CopyCmd)
	rootCmd.Execute()
}
