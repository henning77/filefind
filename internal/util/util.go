package util

import (
	"path"
	"regexp"
	"strings"
)

var filenameRegexToExclude = []*regexp.Regexp{
	regexp.MustCompile(`\.git`),
	regexp.MustCompile(`\.svn`),
	regexp.MustCompile(`\.idea`),
	regexp.MustCompile(`\.vscode`),
	regexp.MustCompile(`.*\.class`),
	regexp.MustCompile(`\.DS_Store`),
	regexp.MustCompile(`.*Temporary.*Items`),
	regexp.MustCompile(`\.@__thumb`),
	regexp.MustCompile(`Network Trash Folder`),
}

func LowercaseExtension(filename string) string {
	ext := path.Ext(filename)
	// Remove dot
	if len(ext) > 1 {
		ext = ext[1:]
	}
	return strings.ToLower(ext)
}

func FileToExclude(filename string) bool {
	for _, re := range filenameRegexToExclude {
		if re.MatchString(filename) {
			return true
		}
	}
	return false
}
