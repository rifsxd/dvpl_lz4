package utils

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rifsxd/dvpl_lz4/common/colors"
	"github.com/rifsxd/dvpl_lz4/common/dvpl"
)

var GlobalPath string

const (
	dvplExtension = ".dvpl"
)

// Config represents the configuration for the program.
type Config struct {
	Mode          string
	KeepOriginals bool
	Path          string // New field to specify the directory path.
	Ignore        string
	IgnoreExt     bool
	Verbose       bool // New field to specify verbose mode.
}

// DVPLFooter represents the DVPL file footer data.
type DVPLFooter struct {
	OriginalSize   uint32
	CompressedSize uint32
	CRC32          uint32
	Type           uint32
}

func PrintElapsedTime(elapsedTime time.Duration) {
	var colorCode string

	// Determine the time unit and color
	switch {
	case elapsedTime.Seconds() < 1:
		colorCode = colors.GreenColor // Milliseconds
		fmt.Printf("\nProcessing took %s%d ms%s\n\n", colorCode, int(elapsedTime.Round(time.Millisecond).Milliseconds()), colors.ResetColor)
		return
	case elapsedTime.Minutes() < 1:
		colorCode = colors.YellowColor // Seconds
		fmt.Printf("\nProcessing took %s%d s%s\n\n", colorCode, int(elapsedTime.Round(time.Second).Seconds()), colors.ResetColor)
		return
	default:
		colorCode = colors.RedColor // Minutes
		fmt.Printf("\nProcessing took %s%d min%s\n\n", colorCode, int(elapsedTime.Round(time.Minute).Minutes()), colors.ResetColor)
		return
	}
}

// ParseCommandLineArgs parses the command-line arguments and returns the configuration.
func ParseCommandLineArgs() (*Config, error) {
	config := &Config{}
	flag.StringVar(&config.Mode, "mode", "", "Mode can be 'compress' / 'decompress' / 'help' (for an extended help guide).")
	flag.BoolVar(&config.KeepOriginals, "keep-originals", false, "Keep original files after compression/decompression.")
	flag.StringVar(&config.Path, "path", "", "directory/files path to process. Default is the current directory.")
	flag.StringVar(&config.Ignore, "ignore", "", "Comma-separated list of file extensions to ignore during compression.")
	flag.BoolVar(&config.Verbose, "verbose", false, "Run in verbose mode (prints detailed log messages).")

	flag.Parse()

	if config.Mode == "" {
		return nil, errors.New("no mode selected. Use '-help' for usage information")
	}

	// Check if the -path flag was provided
	if config.Path == "" {
		// If not, set the path to the current directory
		if initialPath, err := os.Getwd(); err == nil {
			config.Path = initialPath
		}
	}

	// Set the global variable to the value of config.Path
	GlobalPath = config.Path

	return config, nil
}

func PrintHelpMessage() {
	fmt.Println(`dvpl_lz4 [-mode] [-keep-originals] [-path]

    • mode can be one of the following:

        compress: compresses files into dvpl.
        decompress: decompresses dvpl files into standard files.
		verify: verify compressed dvpl files to determine valid compression.
		gui: opens the graphical user interface window.
        help: show this help message.

	• flags can be one of the following:

    	-keep-originals flag keeps the original files after compression/decompression.
		-path specifies the directory/files path to process. Default is the current directory.
		-ignore specifies comma-separated file extensions to ignore during compression.
		-silent disables all file processing verbose information

	• usage can be one of the following examples:

		$ dvpl_lz4 -mode help
		
		$ dvpl_lz4 -mode gui

		$ dvpl_lz4 -mode gui -path /path/to/gui

		$ dvpl_lz4 -mode decompress -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode decompress -keep-originals -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode compress -keep-originals -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode decompress -path /path/to/decompress/compress.yaml.dvpl
		
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress.yaml
		
		$ dvpl_lz4 -mode decompress -keep-originals -path /path/to/decompress/compress.yaml.dvpl
		
		$ dvpl_lz4 -mode dcompress -keep-originals -path /path/to/decompress/compress.yaml

		$ dvpl_lz4 -mode compress -path /path/to/decompress -ignore .exe,.dll

		$ dvpl_lz4 -mode verify -path /path/to/verify/compress.yaml.dvpl

		$ dvpl_lz4 -mode verify -path /path/to/verify/

		$ dvpl_lz4 -mode dcompress -silent

	`)
}

// ProcessFiles process files in the directory or file specified in the config.
func ProcessFiles(directoryOrFile string, config *Config) (successCount, failureCount, ignoredCount int, err error) {
	// Initialize counters
	successCount = 0
	failureCount = 0
	ignoredCount = 0

	info, err := os.Stat(directoryOrFile)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get the path of the currently running executable
	executablePath, err := os.Executable()
	if err != nil {
		return 0, 0, 0, err
	}

	if info.IsDir() {
		dirList, err := os.ReadDir(directoryOrFile)
		if err != nil {
			return 0, 0, 0, err
		}

		for _, dirItem := range dirList {
			succ, fail, ignored, err := ProcessFiles(filepath.Join(directoryOrFile, dirItem.Name()), config)
			if err != nil {
				if config.Verbose {
					fmt.Printf("\n%sError%s processing directory %s: %v\n", colors.RedColor, colors.ResetColor, dirItem.Name(), err)
				}
			}
			successCount += succ
			failureCount += fail
			ignoredCount += ignored
		}
	} else {
		// Check if the file is the executable itself
		if directoryOrFile == executablePath {
			if config.Verbose {
				fmt.Printf("\n%sIgnoring%s own executable file %s\n", colors.YellowColor, colors.ResetColor, directoryOrFile)
			}
			ignoredCount++
			return successCount, failureCount, ignoredCount, nil
		}

		isDecompression := config.Mode == "decompress" && strings.HasSuffix(directoryOrFile, dvplExtension)
		isCompression := config.Mode == "compress" && !strings.HasSuffix(directoryOrFile, dvplExtension)

		ignoreExtensions := make(map[string]bool)
		if config.Ignore != "" {
			extensions := strings.Split(config.Ignore, ",")
			for _, ext := range extensions {
				ignoreExtensions[ext] = true
			}
		}

		shouldIgnore := ignoreExtensions[filepath.Ext(directoryOrFile)]

		if !shouldIgnore && (isDecompression || isCompression) {
			filePath := directoryOrFile
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				if config.Verbose {
					fmt.Printf("\n%sError%s reading file %s: %v\n", colors.RedColor, colors.ResetColor, directoryOrFile, err)
				}
				return 0, 0, 0, err
			}

			var processedBlock []byte
			newName := ""

			if isCompression {
				processedBlock, err = dvpl.CompressDVPL(fileData)
				newName = directoryOrFile + dvplExtension
			} else {
				processedBlock, err = dvpl.DecompressDVPL(fileData)
				newName = strings.TrimSuffix(directoryOrFile, dvplExtension)
			}

			if err != nil {
				if config.Verbose {
					fmt.Printf("\n%sFile%s %s %sfailed to convert due to %v%s\n", colors.RedColor, colors.ResetColor, directoryOrFile, colors.RedColor, err, colors.ResetColor)
				}
				return 0, 1, 0, nil // Return failure count as 1 for this file
			}

			err = os.WriteFile(newName, processedBlock, 0644)
			if err != nil {
				if config.Verbose {
					fmt.Printf("\n%sError%s writing file %s: %v\n", colors.RedColor, colors.ResetColor, newName, err)
				}
				return 0, 0, 0, err
			}

			if config.Verbose {
				fmt.Printf("\n%sFile%s %s has been successfully %s into %s%s%s\n", colors.GreenColor, colors.ResetColor, filePath, getAction(config.Mode), colors.GreenColor, newName, colors.ResetColor)
			}

			if !config.KeepOriginals {
				err := os.Remove(filePath)
				if err != nil {
					if config.Verbose {
						fmt.Printf("\n%sError%s deleting file %s: %v\n", colors.RedColor, colors.ResetColor, filePath, err)
					}
				}
			}

			successCount++
		} else {
			if config.Verbose {
				fmt.Printf("\n%sIgnoring%s file %s\n", colors.YellowColor, colors.ResetColor, directoryOrFile)
			}
			ignoredCount++
		}
	}

	return successCount, failureCount, ignoredCount, nil
}

func getAction(mode string) string {
	if mode == "compress" {
		return colors.GreenColor + "compressed" + colors.ResetColor
	}
	return colors.GreenColor + "decompressed" + colors.ResetColor
}

func VerifyDVPLFiles(directoryOrFile string, config *Config) (successCount, failureCount, ignoredCount int, err error) {
	// Initialize counters
	successCount = 0
	failureCount = 0
	ignoredCount = 0

	info, err := os.Stat(directoryOrFile)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get the path of the currently running executable
	executablePath, err := os.Executable()
	if err != nil {
		return 0, 0, 0, err
	}

	if info.IsDir() {
		dirList, err := os.ReadDir(directoryOrFile)
		if err != nil {
			return 0, 0, 0, err
		}

		for _, dirItem := range dirList {
			succ, fail, ignored, err := VerifyDVPLFiles(filepath.Join(directoryOrFile, dirItem.Name()), config)
			if err != nil {
				if config.Verbose {
					fmt.Printf("\n%sError%s processing directory %s: %v\n", colors.RedColor, colors.ResetColor, dirItem.Name(), err)
				}
			}
			successCount += succ
			failureCount += fail
			ignoredCount += ignored
		}
	} else {
		// Check if the file is the executable itself
		if directoryOrFile == executablePath {
			if config.Verbose {
				fmt.Printf("\n%sIgnoring%s own executable file %s\n", colors.YellowColor, colors.ResetColor, directoryOrFile)
			}
			ignoredCount++
			return successCount, failureCount, ignoredCount, nil
		}

		// Ignore non-.dvpl files during verification
		if !strings.HasSuffix(directoryOrFile, dvplExtension) {
			if config.Verbose {
				fmt.Printf("\n%sIgnoring%s file %s\n", colors.YellowColor, colors.ResetColor, directoryOrFile)
			}
			ignoredCount++
			return successCount, failureCount, ignoredCount, nil
		}

		filePath := directoryOrFile
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			if config.Verbose {
				fmt.Printf("\n%sError%s reading file %s: %v\n", colors.RedColor, colors.ResetColor, directoryOrFile, err)
			}
			return 0, 0, 0, err
		}

		_, err = dvpl.DecompressDVPL(fileData)
		if err != nil {
			if config.Verbose {
				fmt.Printf("\n%sFile%s %s %sfailed to verify due to %v%s\n", colors.RedColor, colors.ResetColor, directoryOrFile, colors.RedColor, err, colors.ResetColor)
			}
			return 0, 1, 0, nil // Return failure count as 1 for this file
		}

		if config.Verbose {
			fmt.Printf("\n%sFile%s %s has been successfully %s\n", colors.GreenColor, colors.ResetColor, filePath, getAction(config.Mode))
		}

		successCount++
	}

	return successCount, failureCount, ignoredCount, nil
}
