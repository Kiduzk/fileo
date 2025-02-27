package main

import (
	"fmt"
	"log"
	"os"
  "io/fs"
	"path/filepath"
	"regexp"
)


// TODO: add option for recursive file search
// This functin organizes file using the name pattern
// INPUT: pattern -> the regex pattern we want to match
//      : outputPath -> the path of where we want the new files to be at
func OrganizeFilesByRegex(regexPattern, outputPath string) error {
  dir, err := os.Getwd()
  HandleError(err)

  files, err := os.ReadDir(dir)
  HandleError(err)

  for _, file := range files {

    // Match the file names with the pattern 
    r, _ := regexp.MatchString(regexPattern, file.Name())

    if r {
      copyFile(file.Name(), outputPath)
    }
  }

  return nil
}

func OrganizeFilesByRegexRecursive(regexPattern, outputPath string) error {
  dir, err := os.Getwd()
  HandleError(err)

  return fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
    HandleError(err)

    // Look at the files and if they match our pattern copy them over to our output path
    if (!d.IsDir()) {

      // Match the file names with the pattern 
      r, _ := regexp.MatchString(regexPattern, path)
      if r {
        copyFile(path, outputPath)
		    fmt.Println(path)
      }
    }

		return nil
	})
}


// Organizes using file extension. Ensures the extension is the ones we want 
func OrganizeFilesByExtension(outputPath, extension string) error {
  return OrganizeFilesByRegex(".*\\." + extension, outputPath)
}

// Copies a source file to the destination folder
func copyFile(src, dst string) {
  
  // Create the destination folder if it does not exist already
  if _, err := os.Stat(src); !os.IsNotExist(err) {
    err := os.MkdirAll(dst, os.ModePerm)
    HandleError(err)
  }

  data, err := os.ReadFile(src)
  HandleError(err)

  // Add the file name to the path
  fullDstPath := filepath.Join(dst, src)
  os.WriteFile(fullDstPath, data, 0644)

  HandleError(err)
  wd, _ := os.Getwd()
  fmt.Println("the dir rn is",wd, src, dst)
}


// General error handler function
func HandleError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
