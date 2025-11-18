package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Usage:   "output directory",
				Aliases: []string{"o"},
			},
			&cli.StringSliceFlag{
				Name:    "ext",
				Usage:   "match an extension (eg: txt)",
				Aliases: []string{"e"},
			},
			&cli.StringSliceFlag{
				Name:    "pattern",
				Usage:   "match a regex pattern",
				Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Usage:   "allow recursive directory search",
				Aliases: []string{"r"},
			},
			&cli.StringFlag{
				Name:    "preview",
				Usage:   "Edit a config file live and see the changes in real time.",
				Aliases: []string{"v"},
			},
			&cli.BoolFlag{
				Name:    "config-create",
				Usage:   "creates a sample a config file",
				Aliases: []string{"config-c"},
			},
			&cli.BoolFlag{
				Name:    "config-apply",
				Usage:   "applies a config file",
				Aliases: []string{"config-a"},
			},
		},
		Name:   "fileo",
		Usage:  "Highly customizable file organizer",
		Action: cliActionHandler,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cliActionHandler(cCtx *cli.Context) error {
	if cCtx.NArg() > 0 {
		// pattern := cCtx.Args().Get(0)
	}

	// Get some of the cli arguments
	outputPath := cCtx.String("output")

	patternSlice := cCtx.StringSlice("pattern")
	extensionSlice := cCtx.StringSlice("extension")

	recursive := cCtx.Bool("recursive")

	configCreate := cCtx.Bool("config-create")
	configApply := cCtx.Bool("config-apply")

	previewConfig := cCtx.String("preview")
	if len(previewConfig) != 0 {
		// Make sure the config file exists in the first place
		if stat, err := os.Stat(previewConfig); err != nil {
			return fmt.Errorf("config file not found: %w", err)
		} else if stat.IsDir() {
			return fmt.Errorf("config filepath must be a directory not a file")
		}
		RunLivePreview(previewConfig)
		return nil
	}

	if outputPath == "" && !configApply && !configCreate {
		return fmt.Errorf("no file output path given")
	}

	if configCreate {
		if err := os.WriteFile("fileo.yaml", []byte(sampleConfig), 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		fmt.Println("Created fileo.yaml")
		return nil
	} else if configApply {
		if err := ApplyConfigFromFile("fileo.yaml"); err != nil {
			return fmt.Errorf("failed to apply config: %w", err)
		}
		return nil
	} else if len(patternSlice) != 0 {
		var organizeFunction func(string, string)

		if recursive {
			organizeFunction = OrganizeFilesByRegexRecursive
		} else {
			organizeFunction = OrganizeFilesByRegex
		}

		for _, pattern := range patternSlice {
			organizeFunction(string(pattern), outputPath)
		}

	} else if len(extensionSlice) != 0 {

		var organizeFunction func(string, string)

		if recursive {
			organizeFunction = OrganizeFilesByExtension
		} else {
			organizeFunction = OrganizeFilesByExtensionRecursive
		}

		for _, extension := range extensionSlice {
			organizeFunction(outputPath, string(extension))
		}
	}

	return nil
}
