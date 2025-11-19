package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

var sampleConfig string = `# This is a sample config file
rules:
- target: so_cool
  extensions: [txt, pdf, sum]

- target: not_so_cool
  extensions: [go]

- target: not_so_cool/nested
  match: [".*"]`

// Copies a list of file list directories into output path
// TODO: conflict resolution of what happens with the copy
func copyMatchedFiles(fileList []string, outputPath string) error {
	for _, file := range fileList {
		copyFile(file, outputPath)
	}
	return nil
}

// This functin organizes file using the name pattern
// INPUT: pattern -> the regex pattern we want to match
//
//	: outputPath -> the path of where we want the new files to be at
//
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
		if !d.IsDir() {

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
	return getRegexMatches(extensionToRegex(extension))
}

func getExtensionMatchesRecursive(extension string) []string {
	return getRegexMatchesRecursive(extensionToRegex(extension))
}

// converts an extension into a regex pattern that matches it
func extensionToRegex(extension string) string {
	return ".*\\." + extension + "$"
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
// TODO: confict copy handling
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
		fmt.Println("the dir rn is", wd, src, dst)
	}
}

// Struct for how config should look
type Rules struct {
	TargetPath   string   `yaml:"target"`
	Extensions   []string `yaml:"extensions"`
	Patterns     []string `yaml:"match"`
	Recurse      bool     `yaml:"recurse"`
	ChildFolders []Rules  `yaml:"folders"`
}

type ConfigData struct {
	FolderRules []Rules `yaml:"rules"`
	Recurse     bool    `yaml:"recurse"`
}

// Takes in a config text input and outputs a list of strings that match the config file
func ApplyConfig(yamlFile []byte) {

	var config ConfigData
	err := yaml.Unmarshal(yamlFile, &config)
	HandleError(err)

	// ensuring the config is valid
	if config.FolderRules == nil {
		HandleError(errors.New("make sure your config has a folders directory"))
	}

	applyConfig(config, ".", config.Recurse, false)
}

// ApplyConfigPreview returns destination paths for preview (where files will be organized to)
func ApplyConfigPreview(yamlFile []byte) []string {
	var config ConfigData
	err := yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return []string{}
	}

	if config.FolderRules == nil {
		return []string{}
	}
	return applyConfig(config, ".", true, true)
}

// A function to read the config file recursively and apply the desired structure
func ApplyConfigFromFile(fileName string) error {
	yamlFile, err := os.ReadFile(fileName)
	HandleError(err)
	ApplyConfig(yamlFile)
	return nil
}

// Given a regex pattern, a slice of file entries and a slice of boolean list of whether a file has already
// been matched, it returns a new slice of filenames that are valid
// It can also be used for extension matching with the flag matchExtension
func filterFilesBasedOnRegex(pattern string, files []string, isFileMatched []bool) []string {

	re := regexp.MustCompile(pattern)
	matches := []string{}

	for i, file := range files {

		// If file has already been matched, skip it
		if isFileMatched[i] {
			continue
		}

		r := re.MatchString(file)

		if r {
			matches = append(matches, file)
			isFileMatched[i] = true
		}
	}

	return matches
}

// Recursively goes through a directory and returns all of the files. If flag for recursion is false
// then it does not look into directories
// It excludes hidden folders and files
func getAllFiles(dir string, recursive bool) []string {
	allFiles := []string{}
	err := fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		HandleError(err)

		// If recursion is toggled off, then ignore all folders, we only care about files
		if !recursive && d.IsDir() {
			return fs.SkipDir
		}

		// Skip hidden folders
		if d.IsDir() && d.Name()[0] == '.' && len(d.Name()) > 1 {
			return fs.SkipDir
		}

		// Skip hidden files hidden files
		if !d.IsDir() && d.Name()[0] == '.' {
			return nil
		}

		// Add the file with full path
		if !d.IsDir() {
			// Create a custom DirEntry that includes the full path
			fullPath := filepath.Join(dir, path)
			allFiles = append(allFiles, fullPath)
		}

		return nil
	})
	HandleError(err)

	return allFiles
}

// Redesigned apply config recurse
// 1) Get all files in the directory (recursive or non-recursive)
// 2) For each rule, uses that rule to match files. Then for all matched files, move/copy them over
// 3) Remove the used/matched folders
// 4) Use catch all to place the remaining
func applyConfig(config ConfigData, parentDir string, recursive, preview bool) []string {

	// Get all files in this directory
	allFiles := getAllFiles(parentDir, recursive)

	// Go through each rule and apply accordingly

	// We use an array to keep track of whether a file has already been matched or not.
	// The current behavior is that if a file already matched with a prevoius pattern,
	// then it can not be matched to any other place again
	isFileMatched := make([]bool, len(allFiles))

	// List of 2 element lists, the filepath and destination path respectively
	matchedFiles := [][2]string{}

	for _, rule := range config.FolderRules {

		ruleMatches := []string{}

		// Regex matches
		for _, pattern := range rule.Patterns {
			matches := filterFilesBasedOnRegex(pattern, allFiles, isFileMatched)
			ruleMatches = append(ruleMatches, matches...)
		}

		// Extension matches
		for _, extension := range rule.Extensions {
			pattern := extensionToRegex(extension)
			matches := filterFilesBasedOnRegex(pattern, allFiles, isFileMatched)
			ruleMatches = append(ruleMatches, matches...)
		}

		// Add each match into the matched files along with its source and target directories
		for _, file := range ruleMatches {
			matchedFiles = append(matchedFiles, [2]string{filepath.Clean(file), filepath.Clean(rule.TargetPath)})
		}
	}

	// Go through the files and enforce that matches should be scoped by their parents
	// TODO


	// If the call was for a live-run we copy over the directories
	if !preview {
		for _, fileNameAndPath := range matchedFiles {
			copyFile(filepath.Base(fileNameAndPath[0]), fileNameAndPath[1])
		}
	}

	// Return the list of full paths, useful for preview
	fullPathNames := make([]string, len(matchedFiles))
	for i, fileNameAndPath := range matchedFiles {
		fullPath := filepath.Join(fileNameAndPath[1], filepath.Base(fileNameAndPath[0]))
		fullPathNames[i] = fullPath
	}

	// for _, x := range fullPathNames {
	// 	fmt.Println(x)
	// }

	return fullPathNames
}

// General error handler function
// TODO: remove this and make errors more clear. Using now for faster development
func HandleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
