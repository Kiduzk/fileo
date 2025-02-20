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
// INPUT: pattern -> the pattern we want to match, using regex
//      : outputPath -> the path of where we want the new files to be at
func OrganizeFilesByPattern(pattern, outputPath string) error {
  dir, err := os.Getwd()
  HandleError(err)

  files, err := os.ReadDir(dir)
  HandleError(err)

  for _, file := range files {

    // Match the file names with the pattern 
    r, _ := regexp.MatchString(pattern, file.Name())

    if r {
      fmt.Println("Copying", file, "into", outputPath)
      copyFile(file.Name(), outputPath, file.Name())
    }
  }

  
  return nil
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
