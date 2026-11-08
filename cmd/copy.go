package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nerten/albumpicker/pkg/config"
	"github.com/nerten/albumpicker/pkg/processor"
)

func printCopyUsage() {
	fmt.Print(`Copy a single music album

Usage:
  albumpicker copy [flags] [album-path]

Flags:
  -c, --config string        config file (default is $HOME/.config/albumpicker/config.yaml)
      --cover-name string    output cover file name
  -d, --destination string   destination directory for copied albums
      --embed-cover          embed resized cover art into audio files (for dual-boot/iTunes)
      --height int           cover image height in pixels (default 240)
  -h, --help                 help for copy
  -s, --source string        source directory with music albums
`)
}

// runCopyCommand executes the copy command
func runCopyCommand(args []string, rootCF ...*commonFlags) error {
	fs := flag.NewFlagSet("copy", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = printCopyUsage

	var cf commonFlags
	registerCommonFlags(fs, &cf)

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	positionalArgs := fs.Args()
	if len(positionalArgs) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(positionalArgs))
	}

	cfgFile := cf.cfgFile
	if cfgFile == "" && len(rootCF) > 0 && rootCF[0] != nil {
		cfgFile = rootCF[0].cfgFile
	}

	conf, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	// Apply root common flags first
	if len(rootCF) > 0 && rootCF[0] != nil {
		applyCommonFlags(conf, rootCF[0])
	}
	// Apply copy common flags
	applyCommonFlags(conf, &cf)

	if err := conf.Validate(); err != nil {
		return err
	}

	targetPath := positionalArgs[0]
	// ensure album path is absolute
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(conf.Source, targetPath)
	}

	albums, err := processor.FindAllAlbums(targetPath)
	if err != nil {
		return fmt.Errorf("error scanning %s directory: %s", targetPath, err)
	}
	// process the album
	return processor.ProcessAlbums(albums, conf)
}
