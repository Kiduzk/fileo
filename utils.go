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
	Recurse    bool `yaml:"recurse"`
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
	
	applyConfigNewwww(config, ".", config.Recurse)
}

// ApplyConfigPreview returns destination paths for preview (where files will be organized to)
func ApplyConfigPreview(yamlFile []byte) []string {
	var data ConfigData
	err := yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return []string{}
	}

	if data.FolderRules == nil {
		return []string{}
	}

	return applyConfigRecurse("", data.FolderRules, []string{}, true, true)
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
func filterFilesBasedOnRegex(pattern string, files []fs.DirEntry, isFileMatched *[]bool) []fs.DirEntry {

	re := regexp.MustCompile(pattern)
	matches := []fs.DirEntry{}

	for i, file := range files {

		// If file has already been matchced, skip it
		if (*isFileMatched)[i] {
			continue
		}

		r := re.MatchString(file.Name())

		if r {
			matches = append(matches, file)
			(*isFileMatched)[i] = true
		}
	}

	return matches
}

// Recursively goes through a directory and returns all of the files. If flag for recursion is false
// then it does not look into directories
// It excludes hidden folders and files
func getAllFiles(dir string, recursive bool) []fs.DirEntry {
	allFiles := []fs.DirEntry{}
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

		// Add the file
		if !d.IsDir() {
			allFiles = append(allFiles, d)
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
func applyConfigNew(config ConfigData, parentDir string, recursive bool) {

	// Get all files in this directory
	allFiles := getAllFiles(parentDir, recursive)

	// Go through each rule and apply accordingly

	// We use an array to keep track of whether a file has already been matched or not.
	// The current behavior is that if a file already matched with a prevoius pattern,
	// then it can not be matched to any other place again
	isFileMatched := make([]bool, len(allFiles))

	for i, rule := range config.FolderRules {

		ruleMatches := []fs.DirEntry{}

		// Ignore matched files. Can be changed later or made to some flag
		if isFileMatched[i] {
			continue
		}

		// Regex matches
		for _, pattern := range rule.Patterns {
			matches := filterFilesBasedOnRegex(pattern, allFiles, &isFileMatched)
			ruleMatches = append(ruleMatches, matches...)
		}

		// Extension matches
		for _, extension := range rule.Extensions {
			pattern := extensionToRegex(extension)
			matches := filterFilesBasedOnRegex(pattern, allFiles, &isFileMatched)
			ruleMatches = append(ruleMatches, matches...)
		}

		// Copy over all the matched files
		ruleMatchesString := []string{}
		for _, rule := range ruleMatches {
			ruleMatchesString = append(ruleMatchesString, rule.Name())
		}

		copyMatchedFiles(ruleMatchesString, rule.TargetPath)
	}
}

// NOTE: General behavior now: if the user specifies a folder within a folder in the config file,
// then the inner folder will only match the files from the ones that matched with the parent file.
// NOTE: also, if a file matches in multiple patterns, the default behavior will create a copy of a file for each match.
// (both the above can be modified but thats the current implementation)
func applyConfigRecurse(parentDir string, folders []Rules, parentMatches []string, firstRun bool, preview bool) []string {
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

		newPath := path.Join(parentDir, folder.TargetPath)

		// If a file has been covered by a subfolder, just skip it
		matches := []string{}
		if len(folder.ChildFolders) == 0 {
			matches = matchesParentCommon
		} else {
			childrenMatches := applyConfigRecurse(newPath, folder.ChildFolders, matchesParentCommon, false, preview)
			currTotalMatches = append(currTotalMatches, childrenMatches...)

			for _, match := range matchesParentCommon {
				if !slices.Contains(childrenMatches, match) {
					matches = append(matches, match)
				}
			}
		}

		if preview {
			// Convert source paths to destination paths
			for _, match := range matches {
				destPath := path.Join(newPath, filepath.Base(match))
				currTotalMatches = append(currTotalMatches, destPath)
			}
		} else {
			// Copy files and return source paths
			copyMatchedFiles(matches, newPath)
			currTotalMatches = append(currTotalMatches, matches...)
		}
	}

	return currTotalMatches
}

// General error handler function
// TODO: remove this and make errors more clear. Using now for faster development
func HandleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
