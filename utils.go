package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)


// TODO: add option for recursive file search
// This functin organizes file using the name pattern
// INPUT: pattern -> the regex pattern we want to match
//      : outputPath -> the path of where we want the new files to be at
func OrganizeFilesByRegex(regexPattern, outputPath string,) error {
  dir, err := os.Getwd()
  HandleError(err)

  files, err := os.ReadDir(dir)
  HandleError(err)

  for _, file := range files {

    // Match the file names with the pattern 
    r, _ := regexp.MatchString(regexPattern, file.Name())

    if r {
      copyFile(file.Name(), outputPath, file.Name())
    }
  }

  return nil
}

// Organizes using file extension. Ensures the extension is the ones we want 
func OrganizeFilesByExtension(outputPath, extension string) error {
  return OrganizeFilesByRegex(".*\\." + extension, outputPath)
}

// Copies a source file to the destination folder
func copyFile(src, dst, fileName string) {
  
  // Create the destination folder if it does not exist already
  if _, err := os.Stat(src); !os.IsNotExist(err) {
    err := os.MkdirAll(dst, os.ModePerm)
    HandleError(err)
  }

  data, err := os.ReadFile(src)
  HandleError(err)

  // Add the file name to the path
  fullDstPath := filepath.Join(dst, fileName)
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
