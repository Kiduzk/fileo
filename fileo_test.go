package main

import (
	"errors"
	"os"
	"path"
	"testing"
)

var subDirName, tempDir string
var err error


func TestMain(m *testing.M) {
  // setup
  tempDir, err = os.MkdirTemp("", "fileo-testing")
  HandleError(err)

  os.Chdir(tempDir)
  files := []string{
    "python1.py",
    "python2.py",
    "python3.py",
    "some_book.pdf",
    "important_document.pdf",
    "interesting_file.txt",
    "interesting_file2.txt",
    "interesting_file2.pdf.py.txt",
  }

  subDirFiles := []string{
    "wow.txt",
    "magnificent.py.txt",
    "magnificent.txt",
    "magnificent.pdf",
    "highly_critical.pdf",
    "finallyworks.py",
  }

  subDirName, err = os.MkdirTemp(tempDir, "sub-directory")
  HandleError(err)


  for _, entry := range files {
    err = os.WriteFile(entry, []byte{}, 0755)
    HandleError(err)
  }

  for _, subdirFile := range subDirFiles {
    os.WriteFile(path.Join(subDirName, subdirFile), []byte{}, 0644)
  }
  //
  // entries, _ := os.ReadDir(".")
  // for _, entry := range entries {
  //   fmt.Println(entry.Name())
  // }


  code := m.Run()

  // cleanup
  os.RemoveAll(tempDir)
  os.Exit(code)
}

func TestGetRegexMatches(t *testing.T) {
  case1 := getRegexMatches(".py")
  case2 := getRegexMatches("inter.*t$")
  case3 := getRegexMatches("^.{3}hon")

  if len(case1) != 4 {
    t.Error("Failed regex matching case 1")
  }

  if len(case2) != 3 {
    t.Error("Failed regex matching case 2")
  }

  if len(case3) != 3 {
    t.Error("Failed regex matching case 3")
  }
}


func TestGetRegexMatchesRecursive(t *testing.T) {
  case1 := getRegexMatchesRecursive(".py")
  case2 := getRegexMatchesRecursive("t$")
  case3 := getRegexMatchesRecursive("^mag")

  if len(case1) != 6 {
    t.Error("Failed regex matching case 1")
  }

  if len(case2) != 6 {
    t.Error("Failed regex matching case 2")
  }

  if len(case3) != 3 {
    t.Error("Failed regex matching case 3")
  }

}


func TestGetExtensionMatches(t *testing.T) {
  pdfMatches := getExtensionMatches("pdf") 
  txtMatches := getExtensionMatches("txt") 
  pyMatches := getExtensionMatches("py") 

  if len(pdfMatches) != 2 {
    t.Error("Failed extension matching for PDF files")
  }
  if len(txtMatches) != 3 {
    t.Error("Failed extension matching for txt files")
  }
  if len(pyMatches) != 3 {
    t.Error("Failed extension matching for py files")
  }
}

func TestGetExtensionMatchesRecursive(t *testing.T) {
  pdfMatches := getExtensionMatchesRecursive("pdf") 
  txtMatches := getExtensionMatchesRecursive("txt") 
  pyMatches := getExtensionMatchesRecursive("py") 

  if len(pdfMatches) != 4 {
    t.Error("Failed extension matching for PDF files")
  }
  if len(txtMatches) != 6 {
    t.Error("Failed extension matching for txt files")
  }
  if len(pyMatches) != 4 {
    t.Error("Failed extension matching for py files")
  }
}

func TestApplyConfig(t *testing.T) {

  sampleConfig :=
  `
  folders:
  - name: "documents"
    recurse: True
    extensions: 
      - "txt"
    patterns:
      - "^magni"

  - name: "code"
    extensions: 
      - "py"
    patterns:
      - "uti"
    folders:
      - name: "only_python"
        patterns:
          - "python1" 

  - name: "broad_documents"
    recurse: True
    extensions:
      - "pdf" 
      - "txt" 
    folders:
      - name: "all_documents" 
        recurse: True  
        patterns:
          - ".*"
  ` 
  err := os.WriteFile("test_config.yaml", []byte(sampleConfig), os.ModePerm)
  HandleError(err)

  err = ApplyConfig("test_config.yaml")
  HandleError(err)


  documentsFiles, err := os.ReadDir("documents")
  HandleError(err)
  if len(documentsFiles) != 7 {
    t.Errorf("ApplyConfig not working. Number of files in 'documents' does not match what was expected: %d != 7", len(documentsFiles))
  }

  codeFiles, err := os.ReadDir("code")
  HandleError(err)
  if len(codeFiles) != 4 {
    t.Errorf("ApplyConfig not working. Number of files in 'code' does not match what was expected: %d != 4", len(codeFiles))
  }

  onlyPythonFiles, err := os.ReadDir("code/only_python")
  HandleError(err)
  if len(onlyPythonFiles) != 1 {
    t.Errorf("ApplyConfig not working. Number of files in 'code' does not match what was expected: %d != 1", len(codeFiles))
  }

  broadDocuments, err := os.ReadDir("broad_documents/all_documents")
  HandleError(err)
  if len(broadDocuments) != 10 {
    t.Errorf("ApplyConfig not working. Number of files in 'broad_documents' does not match what was expected: %d != 10", len(broadDocuments))
  }



  pathExists(t, "documents")
  pathExists(t, "code")
  pathExists(t, "code/only_python")
  pathExists(t, "broad_documents/all_documents")

}

func TestCopyFile(t *testing.T) {

  copyFile(path.Join(tempDir, "python1.py"), subDirName) 
  copyFile(path.Join(subDirName, "finallyworks.py"), tempDir) 

  
  if _, err := os.Stat(path.Join(tempDir, "python1.py")); err != nil {
    if errors.Is(err, os.ErrNotExist) {
      t.Error("CopyFile failed. New file not created for the first test case")
    } else {
      t.Error("Copyfile failed. File created but error with accessing or other error") 
    }
  } 

  if _, err := os.Stat(path.Join(subDirName, "finallyworks.py")); err != nil {
    if errors.Is(err, os.ErrNotExist) {
      t.Error("CopyFile failed. New file not created for the second test case")
    } else {
      t.Error("Copyfile failed. File created but error with accessing or other error") 
    }
  } 
}


// helper function, checks if a folder/file exists
func pathExists(t *testing.T, pathName string) {
  if _, err := os.Stat(path.Join(tempDir, pathName)); err != nil {
    if errors.Is(err, os.ErrNotExist) {
      t.Errorf("ApplyConfig not working. Error creating folder/file: %s", pathName)
    } else {
      t.Error("ApplyConfig not working. Folder/file created, but not able to access it")
    }
  } 

} 



// NOT yet implemented
func TestMovefile(t *testing.T) {
}


