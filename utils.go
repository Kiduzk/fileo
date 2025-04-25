package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"

	"gopkg.in/yaml.v3"
)

var sampleConfig string =
`# This is a sample config file

# Filters out all documents (txt, pdf and docx) which have dates in their names
folders:
- name: 'dated_documents'
  recurse: True
  patterns: ['\d{4}-\d{2}-\d{2}']
  extensions: ['txt', 'pdf', 'docx']


# You can also nest folders within folders. A more complex use case might look like the following. 

# Filters out all documents (txt, pdf and docx)
- name: 'all_documents'
  extensions: ['txt', 'pdf', 'docx']

  # Create a sub-folders to filter out the txt and PDFs into their own folder.
  # represents the directory: all_documents/text_and_PDFS 
  folders:
    - name: "text_and_pdf"
      extensions: ['txt', 'pdf']

      # Puts the text files further into their own folder
      # represents the directory: all_documents/text_and_PDFS/text 
      folders:
        - name: "text"
          extensions: ["txt"]

    # Puts the word files into their own directory
    - name: "words"
      extensions: ["docx"]
  ` 


 
func copyMatchedFiles(fileList []string, outputPath string) error {
  for _, file := range fileList { 
    copyFile(file, outputPath)
  }
  return nil
}

// This functin organizes file using the name pattern
// INPUT: pattern -> the regex pattern we want to match
//      : outputPath -> the path of where we want the new files to be at
// OUTPUT: list of the file names that matched, it does not actually copy them
func getRegexMatches(regexPattern string) []string {
  dir, err := os.Getwd()
  HandleError(err)

  re := regexp.MustCompile(regexPattern)

  files, err := os.ReadDir(dir)
  HandleError(err)

  matched := []string{}

  for _, file := range files {

    if file.IsDir() {
      continue
    } 

    // Match the file names with the pattern 
    r := re.MatchString(file.Name())
    HandleError(err)

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
      r, _ := regexp.MatchString(regexPattern, d.Name())
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
  return getRegexMatches(".*\\." + extension + "$") 
}

func getExtensionMatchesRecursive(extension string) []string {
  return getRegexMatchesRecursive(".*\\." + extension + "$") 
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
  ChildFolders []Folder `yaml:"folders"`
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
  
  // Ensuring the config is valid
  if data.Folders == nil {
    return errors.New("Make sure your config has a folders directory")
  }

  // we enter here, there must always be a folders key in the yaml files
  applyConfigRecurse("", data.Folders, []string{}, true)
  return nil
}


// NOTE: General behavior now: if the user specifies a folder within a folder in the config file,
// then the inner folder will only match the files from the ones that matched with the parent file.
// NOTE: also, if a file matches in multiple patterns, the default behavior will create a copy of a file for each match. 
// (both the above can be modified but thats the current implementation)
func applyConfigRecurse(parentDir string, folders []Folder, parentMatches []string, firstRun bool) []string {
  currTotalMatches := []string{}

  for _, folder := range folders {
    
    extensionMatches := []string{}
    patternMatches := []string{}

    // Handle the extensions
    for _, extension := range folder.Extensions {
      if folder.Recurse {
        extensionMatches = append(extensionMatches, getExtensionMatchesRecursive(extension)...)
      } else {
        extensionMatches = append(extensionMatches, getExtensionMatches(extension)...) 
      }
    }

    for _, pattern := range folder.Patterns {
      if folder.Recurse {
        patternMatches = append(patternMatches, getRegexMatchesRecursive(pattern)...)
      } else {
        patternMatches = append(patternMatches, getRegexMatches(pattern)...) 
      }
    }

    // TODO: this part could use some work 
    currMatches := []string{}
    for _, patternMatch := range patternMatches {
      if slices.Contains(extensionMatches, patternMatch) {
        currMatches = append(currMatches, patternMatch)
      }
    }

    if len(folder.Extensions) == 0 {
      currMatches = patternMatches
    }

    if len(folder.Patterns) == 0 {
      currMatches = extensionMatches
    }

    // Look through only the parent matches
    matchesParentCommon := []string{}
    if !firstRun {
      for _, item1 := range currMatches {
        if slices.Contains(parentMatches, item1) {
          matchesParentCommon = append(matchesParentCommon, item1)
        }
      } 
    } else {
      matchesParentCommon = currMatches 
    }

    newPath := path.Join(parentDir, folder.Name)

    // If a file has been covered by a subfolder, just skip it 
    matches := []string{}
    if len(folder.ChildFolders) == 0 {
      matches = matchesParentCommon
    } else {
      childrenMatches := applyConfigRecurse(newPath, folder.ChildFolders, matchesParentCommon, false)
      currTotalMatches = append(currTotalMatches, childrenMatches...)

      for _, match := range matchesParentCommon {
        if !slices.Contains(childrenMatches, match) {
          matches = append(matches, match)
        }
      }
    }

    copyMatchedFiles(matches, newPath) 
    currTotalMatches = append(currTotalMatches, matches...)
  }

  return currTotalMatches 
}


// General error handler function
func HandleError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
