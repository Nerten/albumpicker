package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nerten/albumpicker/pkg/config"
	"github.com/nerten/albumpicker/pkg/processor"
)

// Copy command
var copyCmd = &cobra.Command{
	Use:   "copy [album-path]",
	Short: "Copy a single FLAC album",
	Args:  cobra.ExactArgs(1),
	RunE:  runCopyCommand,
}

func init() {
	rootCmd.AddCommand(copyCmd)
}

// runCopyCommand executes the copy command
func runCopyCommand(_ *cobra.Command, args []string) error {
	// load configuration
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	path := args[0]
	// ensure album path is absolute
	if !filepath.IsAbs(path) {
		path = filepath.Join(conf.Source, path)
	}

	albums, err := processor.FindAllAlbums(path)
	if err != nil {
		return fmt.Errorf("error scanning %s directory: %s", path, err)
	}
	// process the album
	return processor.ProcessAlbums(albums, conf)
}
