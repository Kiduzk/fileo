package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
  // maintui()
  // err := ApplyConfig("config.yaml")
  // HandleError(err)
  //
  // return 
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name: "output",
        Usage: "The directory of output files.",
        Aliases: []string{"o"},
      },
      &cli.StringSliceFlag{
        Name: "extension",
        Usage: "Matches files with a specific extension.",
        Aliases: []string{"e"},
      },
      &cli.StringSliceFlag{
        Name: "pattern",
        Usage: "Pattern to match with file name, supports regex.",
        Aliases: []string{"p"},
      },
      &cli.BoolFlag{
        Name: "recursive",
        Usage: "Option to recursively search a directory.",
        Aliases: []string{"r"},
      },
      &cli.StringFlag{
        Name: "config",
        Usage: "Applies a config file",
        Aliases: []string{"c"},
      },
    },
    Name: "FileOrganizer",
    Usage: "Organizes files nicely.",
    Action: cliActionHandler,
  }
  if err := app.Run(os.Args); err != nil {
    log.Fatal(err)
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

  mimeType := cCtx.String("mime")

  recursive := cCtx.Bool("recursive")

  config := cCtx.String("config")

  if len(config) != 0 {
    ApplyConfig(config)
  } else if outputPath == "" {
    log.Fatal("No file output path given.")
    return nil
  }

  if len(patternSlice) != 0 {
    var organizeFunction func(string, string) 

    if recursive {
      organizeFunction = OrganizeFilesByRegexRecursive
    } else {
      organizeFunction = OrganizeFilesByRegex
    }


    for _, pattern := range patternSlice {
      organizeFunction(string(pattern), outputPath)
    }

  } else if (len(extensionSlice) != 0) {

    var organizeFunction func(string, string) 

    if recursive {
      organizeFunction = OrganizeFilesByExtension
    } else {
      organizeFunction = OrganizeFilesByExtensionRecursive
    }

    for _, extension := range extensionSlice {
      organizeFunction(outputPath, string(extension))
    }

  } else if (mimeType != "") {

  } 

  return nil
}

