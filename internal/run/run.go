package run

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lorentzforces/fresh-err/fresherr"
)

func FailOut(msg string) {
	fmt.Fprintln(os.Stderr, ErrMsg(msg))
	os.Exit(1)
}

func FailOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, ErrMsg(err.Error()))
		os.Exit(1)
	}
}

func ErrMsg(msg string) string {
	return fresherr.GetFresh() + ": " + msg
}

// Attempt to move a file from one location to another.
//
// This first attempts a simple move operation via os.Rename. However, os.Rename notably does not
// work across filesystem bounaries on Unix (and some other cases). It will attempt to detect if
// the operation failed for one of these reasons and will attempt to move the file via a
// copy-then-delete operation if so.
//
// Rename is always tried first because it is significantly faster (and has fewer filesystem side
// effects) when successful.
//
// The more robust copy-then delete code was closely copied from this StackOverflow answer:
// https://stackoverflow.com/questions/50740902/move-a-file-to-a-different-drive-with-go
func MoveFile(srcPath, destPath string) error {
	err := os.Rename(srcPath, destPath)
	if err == nil { return nil }

	// Attempt to determine if the given error is from a rename operation which is not supported on
	// the current platform for whatever reason (e.g. moving across a filesystem boundary on Unix).
	// These may present very differently on different platforms, so be ready to update this.
	isIncompatibleRenameError := strings.Contains(err.Error(), "invalid cross-device link")
	if !isIncompatibleRenameError { return err }

	inputFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("move file with copy: couldn't open source file: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("move file with copy: couldn't open dest file: %w", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("move file with copy: couldn't copy to dest from source: %w", err)
	}

	// for Windows, close before trying to remove: https://stackoverflow.com/a/64943554/246801
	inputFile.Close()
	err = os.Remove(srcPath)
	if err != nil {
		return fmt.Errorf("move file with copy: couldn't remove source file: %w", err)
	}

	return nil
}
