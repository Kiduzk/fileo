package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name: "output",
        Usage: "The path for the output directory",
      },
    },
    Name: "FileOrganizer",
    Usage: "Organize files",
    Action: cliActionHandler,
  }
  if err := app.Run(os.Args); err != nil {
    log.Fatal(err)
  }
}


func cliActionHandler(cCtx *cli.Context) error {
  if cCtx.NArg() > 0 {
    pattern := cCtx.Args().Get(0)
    organizeFilesByPattern(pattern, "/cool")
  }

  if cCtx.String("pattern") != "" {
    fmt.Println("pattern here")
  }
  
  return nil
}

func organizeFilesByPattern(pattern, outputPath string) error {
  dir, err := os.Getwd()
  HandleError(err)

  files, err := os.ReadDir(dir)
  HandleError(err)

  for _, file := range files {
    if strings.Contains(file.Name(), pattern) {
      fmt.Println("Copying", file, "into", outputPath)
    }
  }

  
  return nil
}

func HandleError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
