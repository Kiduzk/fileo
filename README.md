# fileo (FileOrganizer)
 
Easy to use and customizable file organizer cli written in Go. Supports the following:
- Matching files using a regex pattern
- Matching files using an extension
- Option for a recursive search to match files within all nested directories
- Ability to specify a config file for batch processes (default config provided)

As of right now, `fileo` only copies files and does not move/delete anything. 

### Installation

If you have go installed, run:
```bash
go install github.com/kiduzk/fileo@latest
```
If you do not have go installed, you can download the appropriate binary from the releases section and add fileo to your path. 

### How to use 

For example, lets put all PDFs into a new folder named pdf_documents.
```bash
fileo -ext pdf -output pdf_documents
```
You can also use the shorthand:
```bash
fileo -e pdf -o pdf_documents
```
Next, lets filter out PDFs with a date in their names using regex flag `-pattern` (or simply `-p`):
```bash
fileo -e pdf -o pdf_documents -p "\d{4}-\d{2}-\d{2}"
```

Lastly, we also have option to recursively consider files within subfolders using the `-recurse` flag (or simply `-r`).
```bash
fileo -e pdf -o pdf_documents -r
```

Fileo also provides an option to specify a config file. You can generate a config by running the following command which will create a default one for you. 
```bash
fileo -config-create
```
This file will look as follows:
```yaml
# This is a sample config file


# Filters out all documents (txt, pdf and docx) which have dates in their names
- name: 'dated_documents'
  recurse: True
  patterns: ['\d{4}-\d{2}-\d{2}']
  extensions: ['txt', 'pdf', 'docx']


# You can also nest folders within folders. A more complex use case might look like the following. 

# Filters out all documents (txt, pdf and docx)
folders:
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
 
```
The above config will generate the following structure provided that the current directory contains .txt, .pdf and .docx files. Unmatched folders will not be created (can be changed if needed).
```
all_documents/                # Matches: txt, pdf, docx
├── text_and_pdfs/            # Matches: txt, pdf
│   ├── text/                 # Matches: txt only
│   └── (pdfs here)           # Matches: pdf only
└── words/                    # Matches: docx only
```

Finally, to apply the config file to the current directory simply use:
```bash
fileo -config-apply
```
**Note**: A file will be copied to the deepest matching directory only within a branch. If it matches multiple sibling subdirectories, it will be copied to all of them. This behavior is the current default but can be changed/modified. Any feedback is appreciated!

Some additional feature ideas:
- Support the option for a live preview of what a config would do before actually applying it
- Have support for a move functionality instead of just copy (a bit risky, but if there is a live preview feature then it might make it more safer)

