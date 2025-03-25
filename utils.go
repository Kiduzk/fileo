package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)
 
func copyMatchedFiles(fileList []string, outputPath string) error {
  for _, file := range fileList{ 
    copyFile(file, outputPath)
  }
  return nil
}

// TODO: add option for recursive file search
// This functin organizes file using the name pattern
// INPUT: pattern -> the regex pattern we want to match
//      : outputPath -> the path of where we want the new files to be at
// OUTPUT: list of the file names that matched, it does not actually copy them
func getRegexMatches(regexPattern string) []string {
  dir, err := os.Getwd()
  HandleError(err)

  files, err := os.ReadDir(dir)
  HandleError(err)

  matched := []string{}

  for _, file := range files {

    if file.IsDir() {
      continue
    } 

    // Match the file names with the pattern 
    r, _ := regexp.MatchString(regexPattern, file.Name())

    if r {
      matched = append(matched, file.Name())
    }
  }
  return matched 
}

// First gets the matches, then copies them over
func OrganizeFilesByRegex(regexPattern, outputPath string) {
  matches := getRegexMatches(regexPattern)
  copyMatchedFiles(matches, outputPath)
}

func OrganizeFilesByRegexRecursive(regexPattern, outputPath string) {
  matches := getRegexMatchesRecursive(regexPattern)
  copyMatchedFiles(matches, outputPath)
}


// TODO: this kind of feels repeated code as the non-recursive version so maybe put them together. But I kind of like that it is repeated since it is more clear for me to understand 

// Function to recursively search for a regex pattern
func getRegexMatchesRecursive(regexPattern string) []string {
  dir, err := os.Getwd()
  HandleError(err)

  matched := []string{}

  err = fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
    HandleError(err)

    // Look at the files and if they match our pattern copy them over to our output path
    if (!d.IsDir()) {

      // Match the file names with the pattern 
      r, _ := regexp.MatchString(regexPattern, path)
      if r {
        matched = append(matched, path)
      }
    }

    return nil
	})
  HandleError(err)

  return matched
}


func getExtensionMatches(extension string) []string {
  return getRegexMatches(".*\\." + extension) 
}

func getExtensionMatchesRecursive(extension string) []string {
  return getRegexMatches(".*\\." + extension) 
}

// Organizes using file extension. 
func OrganizeFilesByExtension(outputPath, extension string) {
  matches := getExtensionMatches(extension)
  copyMatchedFiles(matches, outputPath)
}

// Organizes using file extension recursively. 
func OrganizeFilesByExtensionRecursive(outputPath, extension string) {
  matches := getExtensionMatchesRecursive(extension)
  copyMatchedFiles(matches, outputPath)
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
  if false {
    fmt.Println("the dir rn is",wd, src, dst)
  }
}

// Struct for how config should look 
type Folder struct {
  Name string `yaml:"name"`;
  Extensions []string `yaml:"extensions"`
  Patterns []string `yaml:"patterns"`
  Recurse bool `yaml:"recurse"`
}


type ConfigData struct{
  Folders []Folder `yaml:"folders"`
}


// Now , we will create a function to read the config file recursively and apply the desired structure
func ApplyConfig(fileName string) error {
  yamlFile, err := os.ReadFile(fileName)
  HandleError(err)

  var data ConfigData 

  err = yaml.Unmarshal(yamlFile, &data)
  HandleError(err)


  // we enter here, there must always be a folders key in the yaml files
  // walkYamlFile(yamlFile)

  
  for _, folder := range data.Folders {
    
    matches := []string{}

    // Handle the extensions
    for _, extension := range folder.Extensions {
      if folder.Recurse {
        matches = append(matches, getExtensionMatches(extension)...)
      } else {
        matches = append(matches, getRegexMatchesRecursive(extension)...) 
      }
    }

    for _, pattern := range folder.Patterns {
      if folder.Recurse {
        matches = append(matches, getRegexMatches(pattern)...)
      } else {
        matches = append(matches, getRegexMatches(pattern)...) 
      }
    }

    copyMatchedFiles(matches, folder.Name)
  }

  // fmt.Println(data.Folders)
  // fmt.Println(fmt.Sprintf("%T", data["fi"]))
  return nil
}


// TODO: to be implemented once done with working with folders that do not have nested folders 
func walkYamlFile(yamlFile map[string]any) error {
  var data map[string]any

  // err := yaml.Unmarshal(yamlFile, &data)
  // HandleError(err)

  fmt.Println(yamlFile, &data)
  return nil
}


// General error handler function
func HandleError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
