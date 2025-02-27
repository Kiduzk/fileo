package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name: "pattern",
        Usage: "Pattern to match with file name, supports regex.",
      },
      &cli.StringFlag{
        Name: "output",
        Usage: "The directory of output files.",
      },
      &cli.StringFlag{
        Name: "extension",
        Usage: "Matches files with a specific extension.",
      },
      &cli.BoolFlag{
        Name: "recursive",
        Usage: "Option to recursively search a directory.",
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

  pattern := cCtx.String("pattern")
  outputPath := cCtx.String("output")
  
  extension := cCtx.String("extension")

  mimeType := cCtx.String("mime")
  
  recursive := cCtx.Bool("recursive")

  if outputPath == "" {
    log.Fatal("No file output path given.")
    return nil
  }

  if pattern != "" {

    if recursive {
      OrganizeFilesByRegexRecursive(pattern, outputPath)
    } else {
      OrganizeFilesByRegex(pattern, outputPath)
    }

  } else if (extension != "") {

    if recursive {
      OrganizeFilesByExtensionRecursive(outputPath, extension)
    } else {
      OrganizeFilesByExtension(outputPath, extension)
    }

  } else if (mimeType != "") {

  } 
  
  return nil
}

