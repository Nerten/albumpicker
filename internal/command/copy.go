package command

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nerten/albumpicker/internal/config"
)

// Copy command
var CopyCmd = &cobra.Command{
	Use:   "copy [album-path]",
	Short: "Copy a single FLAC album",
	Args:  cobra.ExactArgs(1),
	RunE:  runCopyCommand,
}

func init() {
}

// runCopyCommand executes the copy command
func runCopyCommand(cmd *cobra.Command, args []string) error {
	// Load configuration
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	albumPath := args[0]
	// Ensure album path is absolute
	if !filepath.IsAbs(albumPath) {
		albumPath = filepath.Join(conf.Source, albumPath)
	}

	// Process the album
	return processAlbum(albumPath, conf)
}
