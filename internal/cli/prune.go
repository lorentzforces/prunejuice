package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/lorentzforces/prunejuice/internal/run"
	"github.com/spf13/cobra"
)

const (
	optionPrintOnly = "print-only"
	optionKeepN = "keep"
	optionNoConfirm = "no-confirm"
	optionSinceUnixTime = "since-unix-time"
	optionOperateOnDirectories = "directories"
	optionClassify = "classify"
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
	rootCmd.Flags().Bool(
		optionClassify,
		false,
		"Print the name of every file found, prefixed by either REMOVE or KEEP",
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
	rootCmd.Flags().Bool(
		optionNoConfirm,
		false,
		"If this flag is set, all confirmation checks will be treated as if the user had\n" +
			"confirmed \"yes.\"",
	)
	rootCmd.Flags().Int64(
		optionSinceUnixTime,
		0, // for documentation purposes this is "no default," this value is not used if set
		"A Unix-epoch timestamp, keep only files created at or after this time",
	)
	rootCmd.Flags().Bool(
		optionOperateOnDirectories,
		false,
		"Operate on directories instead of regular files",
	)

	return rootCmd
}

func runPruneJuice(cmd *cobra.Command, args []string) error {
	keepNumber, err := cmd.Flags().GetInt(optionKeepN)
	if err != nil { return fmt.Errorf("Error while getting command line flags: %w", err) }
	if keepNumber < 0 {
		return fmt.Errorf("Cannot keep a negative number of files (was given %d)", keepNumber)
	}

	classifyResults, err := cmd.Flags().GetBool(optionClassify)
	run.FailOnErr(err)

	printOnly, err := cmd.Flags().GetBool(optionPrintOnly)
	run.FailOnErr(err)

	shouldNotConfirm, err := cmd.Flags().GetBool(optionNoConfirm)
	run.FailOnErr(err)
	confirmSetting := confirmOrNotFromBool(!shouldNotConfirm)

	unixTimestampProvided := cmd.Flag(optionSinceUnixTime).Changed
	unixTimestamp, err := cmd.Flags().GetInt64(optionSinceUnixTime)
	run.FailOnErr(err)

	operateOnDirectories, err := cmd.Flags().GetBool(optionOperateOnDirectories)
	run.FailOnErr(err)
	fileTypeToUse := fileTypeFromBool(operateOnDirectories)

	if len(args) != 1 { return fmt.Errorf("Expected 1 path argument but found %d", len(args)) }

	files, err := findAllFilesInPath(args[0], fileTypeToUse)
	if err != nil { return err }

	slices.SortStableFunc(files, func(l, r foundDirEntry) int {
		if l.ModifiedTime.After(r.ModifiedTime) {
			return 1
		} else if l.ModifiedTime.Before(r.ModifiedTime) {
			return -1
		} else {
			return 0
		}
	})

	keepClassifiers := make([]dirEntryClassifier, 0)

	if unixTimestampProvided {
		keepClassifiers = append(keepClassifiers, keepAtOrAfterUnixTime(unixTimestamp))
	}

	firstIndexToKeep := len(files)
	ENTRY_LOOP: for i, dirEntry := range files {
		for _, classifier := range keepClassifiers {
			if classifier(dirEntry) {
				firstIndexToKeep = i
				break ENTRY_LOOP
			}
		}
	}

	// if our first keep index would leave us with fewer than our minimum number of entries to
	// keep, set it lower so we still keep that backstop number
	keepNumberBackstop := len(files) - keepNumber
	firstIndexToKeep = min(keepNumberBackstop, firstIndexToKeep)

	filesToRemove := files[0:firstIndexToKeep]

	if classifyResults {
		doClassifyResults(files, firstIndexToKeep)
	} else if printOnly {
		doPrintFiles(filesToRemove)
	} else {
		err = doDelete(filesToRemove, confirmSetting)
		if err != nil { return err }
	}

	return nil
}

type foundDirEntry struct {
	RelativePath string
	FullPath string
	ModifiedTime time.Time
}

type fileType uint8
const (
	fileTypePlainFile fileType = iota
	fileTypeDirectory
)

func findAllFilesInPath(pathToDir string, typeToFind fileType) ([]foundDirEntry, error) {
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
		if dirEntry.IsDir() && typeToFind == fileTypePlainFile { continue }
		if !dirEntry.IsDir() && typeToFind == fileTypeDirectory { continue }

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

func doDelete(files []foundDirEntry, confirmSetting confirmOrNot) error {
	var filesPrompt strings.Builder
	filesPrompt.WriteString("The following files will be deleted: ")
	for _, file := range files {
		filesPrompt.WriteString("\n  ")
		filesPrompt.WriteString(file.RelativePath)
	}

	if confirmSetting == doConfirm {
		err := run.PromptConfirm(filesPrompt.String())
		if err != nil { return err }
	}

	errs := make([]error, 0)
	for _, file := range files {
		err := os.RemoveAll(file.FullPath)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("Encountered errors when deleting: %w", errors.Join(errs...))
	}
	return nil
}

func doClassifyResults(files []foundDirEntry, firstIndexToKeep int) {
	for i, file := range files {
		if i < firstIndexToKeep {
			fmt.Print("REMOVE ")
		} else {
			fmt.Print("KEEP ")
		}
		fmt.Println(file.RelativePath)
	}
}

func doPrintFiles(files []foundDirEntry) {
	for _, file := range files {
		fmt.Println(file.RelativePath)
	}
}

type confirmOrNot uint8
const (
	doConfirm confirmOrNot = iota
	doNotConfirm
)

func confirmOrNotFromBool(boolVal bool) confirmOrNot {
	if boolVal {
		return doConfirm
	}
	return doNotConfirm
}

func fileTypeFromBool(boolVal bool) fileType {
	if boolVal {
		return fileTypeDirectory
	}
	return fileTypePlainFile
}

// classifier function for dir entries: for a given entry, whether or not it should be kept
type dirEntryClassifier func(foundDirEntry) bool

func keepAtOrAfterUnixTime(timestamp int64) dirEntryClassifier {
	return keepAtOrAfterTime(time.Unix(timestamp, 0))
}

func keepAtOrAfterTime(timestamp time.Time) dirEntryClassifier {
	return func(dirEntry foundDirEntry) bool {
		return !dirEntry.ModifiedTime.Before(timestamp)
	}
}
