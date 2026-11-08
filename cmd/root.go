package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/nerten/albumpicker/pkg/config"
)

var osExit = os.Exit

type commonFlags struct {
	cfgFile    string
	source     string
	dest       string
	height     int
	coverName  string
	embedCover bool
}

func registerCommonFlags(fs *flag.FlagSet, cf *commonFlags) {
	fs.StringVar(&cf.cfgFile, "config", "", "config file (default is $HOME/.config/albumpicker/config.yaml)")
	fs.StringVar(&cf.cfgFile, "c", "", "config file (shorthand)")

	fs.StringVar(&cf.source, "source", "", "source directory with music albums")
	fs.StringVar(&cf.source, "s", "", "source directory with music albums (shorthand)")

	fs.StringVar(&cf.dest, "destination", "", "destination directory for copied albums")
	fs.StringVar(&cf.dest, "d", "", "destination directory for copied albums (shorthand)")

	fs.IntVar(&cf.height, "height", 0, "cover image height in pixels (default 240)")

	fs.StringVar(&cf.coverName, "cover-name", "", "output cover file name")

	fs.BoolVar(&cf.embedCover, "embed-cover", false, "embed resized cover art into audio files (for dual-boot/iTunes)")
}

func applyCommonFlags(cfg *config.Config, cf *commonFlags) {
	if cf.source != "" {
		cfg.Source = cf.source
	}
	if cf.dest != "" {
		cfg.Destination = cf.dest
	}
	if cf.height != 0 {
		cfg.CoverHeight = cf.height
	}
	if cf.coverName != "" {
		cfg.OutputCoverName = cf.coverName
	}
	if cf.embedCover {
		cfg.EmbedCover = true
	}
}

func printUsage() {
	fmt.Print(`A console utility written in Go for randomly selecting and copying music albums
(FLAC, MP3, ALAC, AAC). The utility allows specifying the number of albums to copy,
provides functionality to remove image metadata from FLAC files, and
copies album covers with size optimization.

Usage:
  albumpicker [command]

Available Commands:
  copy        Copy a single music album
  help        Help about any command
  pick        Randomly select and copy music albums
  version     Print albumpicker version

Flags:
  -c, --config string        config file (default is $HOME/.config/albumpicker/config.yaml)
      --cover-name string    output cover file name
  -d, --destination string   destination directory for copied albums
      --embed-cover          embed resized cover art into audio files (for dual-boot/iTunes)
      --height int           cover image height in pixels (default 240)
  -h, --help                 help for albumpicker
  -s, --source string        source directory with music albums
  -v, --version              version for albumpicker

Use "albumpicker [command] --help" for more information about a command.
`)
}

// Execute parses command line arguments and runs the appropriate command.
func Execute(version string) {
	fs := flag.NewFlagSet("albumpicker", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = printUsage

	var (
		versionFlag bool
		cf          commonFlags
	)

	fs.BoolVar(&versionFlag, "version", false, "version for albumpicker")
	fs.BoolVar(&versionFlag, "v", false, "version for albumpicker (shorthand)")
	registerCommonFlags(fs, &cf)

	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		return
	}

	// Handle standalone subcommands before flag parsing if first arg is not a flag
	switch args[0] {
	case "version":
		fmt.Fprintf(os.Stdout, "albumpicker version %s\n", version)
		return
	case "help":
		if len(args) > 1 {
			switch args[1] {
			case "pick":
				printPickUsage()
				return
			case "copy":
				printCopyUsage()
				return
			}
		}
		printUsage()
		return
	case "pick":
		if err := runPickCommand(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
		return
	case "copy":
		if err := runCopyCommand(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
		return
	}

	// If first arg starts with '-', parse root flags
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		osExit(1)
		return
	}

	if versionFlag {
		fmt.Fprintf(os.Stdout, "albumpicker version %s\n", version)
		return
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		printUsage()
		return
	}

	switch remaining[0] {
	case "pick":
		if err := runPickCommand(remaining[1:], &cf); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	case "copy":
		if err := runCopyCommand(remaining[1:], &cf); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	case "help":
		printUsage()
	case "version":
		fmt.Fprintf(os.Stdout, "albumpicker version %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q for \"albumpicker\"\nRun 'albumpicker --help' for usage.\n", remaining[0])
		osExit(1)
	}
}
