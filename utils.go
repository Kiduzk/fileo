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

    if file.IsDir() {
      continue
    } 

    // Match the file names with the pattern 
    r, _ := regexp.MatchString(regexPattern, file.Name())

    if r {
      fmt.Println(r, regexPattern, file.Name())
      copyFile(file.Name(), outputPath)
    }
  }

  return nil
}


// TODO: this kind of feels repeated code as the non-recursive version so maybe put them together. But
// I kind of like that it is repeated since it is more clear for me to understand 

// Function to recursively search for a regex pattern
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
      }
    }

		return nil
	})
}


// Organizes using file extension. 
func OrganizeFilesByExtension(outputPath, extension string) error {
  return OrganizeFilesByRegex(".*\\." + extension, outputPath)
}

// Organizes using file extension recursively. 
func OrganizeFilesByExtensionRecursive(outputPath, extension string) error {
  return OrganizeFilesByRegexRecursive(".*\\." + extension, outputPath)
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

  // Get the file name
  _, fileName := filepath.Split(src)

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
