package cmd

import (
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

	albumPath := args[0]
	// ensure album path is absolute
	if !filepath.IsAbs(albumPath) {
		albumPath = filepath.Join(conf.Source, albumPath)
	}

	// process the album
	return processor.ProcessAlbum(albumPath, conf)
}
