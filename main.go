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
        Usage: "The directory of output files",
      },

    },
    Name: "FileOrganizer",
    Usage: "Organizes files nicely",
    Action: cliActionHandler,
  }
  if err := app.Run(os.Args); err != nil {
    log.Fatal(err)
  }
}


func cliActionHandler(cCtx *cli.Context) error {
  if cCtx.NArg() > 0 {
    pattern := cCtx.Args().Get(0)
    OrganizeFilesByPattern(pattern, "/cool")
  }

  // Get some of the cli arguments
  pattern := cCtx.String("pattern")
  outputPath := cCtx.String("output")

  if pattern != "" {
    OrganizeFilesByPattern(pattern, outputPath)
  }
  
  return nil
}

