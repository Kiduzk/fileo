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
        Name: "output",
        Usage: "output directory",
        Aliases: []string{"o"},
      },
      &cli.StringSliceFlag{
        Name: "ext",
        Usage: "match an extension (eg: txt)",
        Aliases: []string{"e"},
      },
      &cli.StringSliceFlag{
        Name: "pattern",
        Usage: "match a regex pattern",
        Aliases: []string{"p"},
      },
      &cli.BoolFlag{
        Name: "recursive",
        Usage: "allow recursive directory search",
        Aliases: []string{"r"},
      },
      &cli.StringFlag{
        Name: "preview",
        Usage: "Edit a config file live and see the changes in real time.",
        Aliases: []string{"v"},
      },
      &cli.BoolFlag{
        Name: "config-create",
        Usage: "creates a sample a config file",
        Aliases: []string{"config-c"},
      },
      &cli.BoolFlag{
        Name: "config-apply",
        Usage: "applies a config file",
        Aliases: []string{"config-a"},
      },
    },
    Name: "fileo",
    Usage: "Highly customizable file organizer",
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

  config_create := cCtx.Bool("config-create")
  config_apply := cCtx.Bool("config-apply")

  // ----- for now live preview is disabled until it works properlt
  previewConfig := cCtx.String("preview")
  if len(previewConfig) != 0 {
    RunLivePreview()
    // RunLivePreview(previewConfig)
  }

  if outputPath == "" && !config_apply && !config_create {
    log.Fatal("No file output path given.")
    return nil
  }

  if config_create { 
    os.WriteFile("fileo.yaml", []byte(sampleConfig), os.ModePerm)
  } else if config_apply {
    ApplyConfig("fileo.yaml")
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

