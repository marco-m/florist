// Package qh contains test helpers to be used with go-quicktest/qt.
package qh

import (
	"os"

	"github.com/go-quicktest/qt"
)

// FileEqualsString returns a qt.Checker checking equality of the contents of file
// fpath with string want.
func FileEqualsString(fpath string, want string) qt.Checker {
	return &fileEqualsChecker{fpath, want}
}

type fileEqualsChecker struct {
	fpath string
	want  string
}

func (fc *fileEqualsChecker) Args() []qt.Arg {
	return []qt.Arg{{"fpath", fc.fpath}, {"want", fc.want}}
}

func (fc *fileEqualsChecker) Check(note func(key string, value any)) error {
	have, err := os.ReadFile(fc.fpath)
	if err != nil {
		return err
	}
	return qt.Equals(string(have), fc.want).Check(note)
}

// FileContains returns a qt.Checker checking whether string want is contained in the file at fpath.
func FileContains(fpath string, want string) qt.Checker {
	return &fileContainsChecker{fpath, want}
}

type fileContainsChecker struct {
	fpath string
	want  string
}

func (fc *fileContainsChecker) Args() []qt.Arg {
	return []qt.Arg{{"fpath", fc.fpath}, {"want", fc.want}}
}

func (fc *fileContainsChecker) Check(note func(key string, value any)) error {
	have, err := os.ReadFile(fc.fpath)
	if err != nil {
		return err
	}
	return qt.StringContains(string(have), fc.want).Check(note)
}
