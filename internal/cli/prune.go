package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/spf13/cobra"
)

const (
	optionPrintOnly = "print-only"
	optionKeepN = "keep"
)

func CreateRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "prunejuice [options] dir-path",
		Long: "prunejuice will read a given directory path and remove old files. By default, it " +
			"will keep the 1 newest file and remove others, ignoring any directories.",
		SilenceUsage: true,
		DisableFlagsInUseLine: true,
		SilenceErrors: true,
		RunE: runPruneJuice,
	}

	rootCmd.InitDefaultHelpFlag()
	rootCmd.Flags().Bool(
		optionPrintOnly,
		false,
		"Print the names of files to be removed only - do not perform any action on them",
	)
	rootCmd.Flags().IntP(
		optionKeepN,
		"N",
		1,
		"Keep only the N newest files. If this is specified in combination with other options,\n" +
			"treat this as the minimum number of files to keep and keep the N newest regardless\n" +
			"of whether or not they would otherwise be removed. is used implicitly. Zero is a\n" +
			"valid value, but you should probably consider with caution if that's really what\n" +
			"you want to use this program for.",
	)

	return rootCmd
}

func runPruneJuice(cmd *cobra.Command, args []string) error {
	keepNumber, err := cmd.Flags().GetInt(optionKeepN)
	if err != nil { return fmt.Errorf("Error while getting command line flags: %w", err) }
	if keepNumber < 0 {
		return fmt.Errorf("Cannot keep a negative number of files (was given %d)", keepNumber)
	}
	printOnly, err := cmd.Flags().GetBool(optionPrintOnly)
	if err != nil { return fmt.Errorf("") }

	if len(args) != 1 { return fmt.Errorf("Expected 1 path argument but found %d", len(args)) }

	files, err := findAllFilesInPath(args[0])
	if err != nil { return err }

	slices.SortStableFunc(files, func(l, r foundDirEntry) int {
		if l.ModifiedTime.After(r.ModifiedTime) {
			return 1
		} else if r.ModifiedTime.After(l.ModifiedTime) {
			return -1
		} else {
			return 0
		}
	})


	lastIndexToRemove := len(files) - keepNumber - 1
	// if we already have fewer files than our minimum, bail
	if lastIndexToRemove < 0 { return nil }

	for i := lastIndexToRemove; i >= 0; i-- {
		if printOnly {
			fmt.Println(files[i].RelativePath)
		} else {
			// TODO: add actual delete here once I'm more confident
			fmt.Println("WOULD DELETE " + files[i].RelativePath)
		}
	}

	return nil
}

type foundDirEntry struct {
	RelativePath string
	FullPath string
	ModifiedTime time.Time
}

func findAllFilesInPath(pathToDir string) ([]foundDirEntry, error) {
	fullPath, err := filepath.Abs(pathToDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to determine canonical path for: %s", pathToDir)
	}
	fileInfo, err := os.Stat(fullPath)
	if err != nil { return nil, fmt.Errorf("Could not find directory path: %s", fullPath) }
	if !fileInfo.IsDir() { return nil, fmt.Errorf("Path is not a directory: %s", fullPath) }

	dirContents, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read contents of directory at \"%s\": %w", fullPath, err)
	}

	results := make([]foundDirEntry, 0, len(dirContents))
	for _, dirEntry := range dirContents {
		if dirEntry.IsDir() { continue }

		childFileInfo, err := dirEntry.Info()
		if err != nil {
			return nil, fmt.Errorf(
				"File \"%s\" removed after reading dir: %s", dirEntry.Name(), fullPath,
			)
		}
		results = append(
			results,
			foundDirEntry{
				RelativePath: path.Join(pathToDir, dirEntry.Name()),
				FullPath: path.Join(fullPath, dirEntry.Name()),
				ModifiedTime: childFileInfo.ModTime(),
			},
		)
	}

	return results, nil
}
