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

func printPickUsage() {
	fmt.Print(`Randomly select and copy music albums

Usage:
  albumpicker pick [flags]

Flags:
  -c, --config string        config file (default is $HOME/.config/albumpicker/config.yaml)
      --cover-name string    output cover file name
  -d, --destination string   destination directory for copied albums
      --embed-cover          embed resized cover art into audio files (for dual-boot/iTunes)
      --height int           cover image height in pixels (default 240)
  -h, --help                 help for pick
  -n, --count int            number of albums to select (default 10)
  -s, --source string        source directory with music albums
      --wipe                 wipe destination directory before copying albums. Attention!!! Destructive action!
`)
}

// runPickCommand executes the pick command
func runPickCommand(args []string, rootCF ...*commonFlags) error {
	fs := flag.NewFlagSet("pick", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = printPickUsage

	var cf commonFlags
	registerCommonFlags(fs, &cf)

	var (
		count int
		wipe  bool
	)
	fs.IntVar(&count, "count", 0, "number of albums to select (default 10)")
	fs.IntVar(&count, "n", 0, "number of albums to select (shorthand)")
	fs.BoolVar(&wipe, "wipe", false, "wipe destination directory before copying albums. Attention!!! Destructive action!")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cfgFile := cf.cfgFile
	if cfgFile == "" && len(rootCF) > 0 && rootCF[0] != nil {
		cfgFile = rootCF[0].cfgFile
	}

	conf, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	// Apply root common flags first (if any)
	if len(rootCF) > 0 && rootCF[0] != nil {
		applyCommonFlags(conf, rootCF[0])
	}
	// Apply pick common flags
	applyCommonFlags(conf, &cf)

	// Apply pick-specific flags
	if count != 0 {
		conf.AlbumsCount = count
	}

	if err := conf.Validate(); err != nil {
		return err
	}

	destination, err := os.ReadDir(conf.Destination)
	if err != nil {
		return fmt.Errorf("error finding destination directory: %s", err)
	}

	// find all albums in source directory
	fmt.Println("Scanning source directory for music albums...")
	albums, err := processor.FindAllAlbums(conf.Source)
	if err != nil {
		return fmt.Errorf("error scanning source directory: %s", err)
	}

	if len(albums) == 0 {
		return fmt.Errorf("no music albums found in source directory")
	}

	fmt.Printf("Found %d albums in total\n", len(albums))

	// select random albums
	fmt.Printf("Selecting %d random albums...\n", conf.AlbumsCount)
	selectedAlbums := processor.SelectRandomAlbums(albums, conf.AlbumsCount)

	// check wipe flag
	if wipe {
		// wipe destination directory
		fmt.Printf("Wiping destination directory: %s\n", conf.Destination)
		for _, d := range destination {
			if err := os.RemoveAll(filepath.Join(conf.Destination, d.Name())); err != nil {
				return fmt.Errorf("failed to wipe destination directory: %s", err)
			}
		}
	}

	// process albums
	fmt.Printf("Processing selected %d albums...\n", len(selectedAlbums))
	return processor.ProcessAlbums(selectedAlbums, conf)
}
