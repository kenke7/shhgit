package core

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MatchFile struct {
	Path      string
	Filename  string
	Extension string
	Contents  []byte
}

func NewMatchFile(path string) MatchFile {
	path = filepath.ToSlash(path)
	_, filename := filepath.Split(path)
	extension := filepath.Ext(path)
	contents, _ := ioutil.ReadFile(path)

	return MatchFile{
		Path:      path,
		Filename:  filename,
		Extension: extension,
		Contents:  contents,
	}
}

func IsSkippableFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))

	for _, skippableExt := range session.Config.BlacklistedExtensions {
		if extension == skippableExt {
			return true
		}
	}

	for _, skippablePathIndicator := range session.Config.BlacklistedPaths {
		skippablePathIndicator = strings.Replace(skippablePathIndicator, "{sep}", string(os.PathSeparator), -1)
		if strings.Contains(path, skippablePathIndicator) {
			return true
		}
	}

	return false
}

func (match MatchFile) CanCheckEntropy() bool {
	if match.Filename == "id_rsa" {
		return false
	}

	for _, skippableExt := range session.Config.BlacklistedEntropyExtensions {
		if match.Extension == skippableExt {
			return false
		}
	}

	return true
}

func GetMatchingFiles(dir string) []MatchFile {
	fileList := make([]MatchFile, 0)
	maxFileSize := *session.Options.MaximumFileSize * 1024

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil || f.IsDir() || uint(f.Size()) > maxFileSize || IsSkippableFile(path) {
			return nil
		}
		fileList = append(fileList, NewMatchFile(path))
		return nil
	})

	return fileList
}

func GetStagedFiles(dir string) []MatchFile {
	fileList := make([]MatchFile, 0)

	out, err := exec.Command("git", "diff", "--staged", "--name-only", "--no-ext-diff", "--diff-filter=ACMRTUXB", "-z", "-C", dir).Output()
	if err != nil {
		return nil
	}

	for _, path := range zsplit(string(out)) {
		path = dir + "/" + path
		if !IsSkippableFile(path) {
			fileList = append(fileList, NewMatchFile(path))
		}
	}

	return fileList
}

func zsplit(str string) []string {
	str = strings.Trim(str, "\000")
	if  len(str) > 0 {
		return strings.Split(str, "\000")
	} else {
		return make([]string, 0)
	}
}