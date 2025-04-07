package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nerten/albumpicker/pkg/config"
	"github.com/nerten/albumpicker/pkg/processor"
)

// Pick command
var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Randomly select and copy FLAC albums",
	RunE:  runPickCommand,
}

func init() {
	rootCmd.AddCommand(pickCmd)
	// local flags
	pickCmd.Flags().IntP("count", "n", 0, "number of albums to select (default 10)")
	pickCmd.Flags().Bool("wipe", false, "wipe destination directory before copying albums. Attention!!! Destructive action!")

	// bind flags to viper
	err := viper.BindPFlag("albums_count", pickCmd.Flags().Lookup("count"))
	if err != nil {
		panic(err.Error())
	}
}

// runPickCommand executes the pick command
func runPickCommand(cmd *cobra.Command, _ []string) error {
	// load configuration
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	destination, err := os.ReadDir(conf.Destination)
	if err != nil {
		return fmt.Errorf("error finding destination directory: %s", err)
	}

	// find all albums in source directory
	fmt.Println("Scanning source directory for FLAC albums...")
	albums, err := processor.FindAllAlbums(conf.Source)
	if err != nil {
		return fmt.Errorf("error scanning source directory: %s", err)
	}

	if len(albums) == 0 {
		return fmt.Errorf("no FLAC albums found in source directory")
	}

	fmt.Printf("Found %d albums in total\n", len(albums))

	// select random albums
	fmt.Printf("Selecting %d random albums...\n", conf.AlbumsCount)
	selectedAlbums := processor.SelectRandomAlbums(albums, conf.AlbumsCount)

	// check wipe flag
	if wipe, _ := cmd.Flags().GetBool("wipe"); wipe {
		// wipe destination directory
		fmt.Printf("Wiping destination directory: %s\n", conf.Destination)
		for _, d := range destination {
			if err := os.RemoveAll(path.Join(conf.Destination, d.Name())); err != nil {
				return fmt.Errorf("failed to wipe destination directory: %s", err)
			}
		}
	}

	// process albums
	fmt.Printf("Processing selected %d albums...\n", len(selectedAlbums))
	return processor.ProcessAlbums(selectedAlbums, conf)
}
