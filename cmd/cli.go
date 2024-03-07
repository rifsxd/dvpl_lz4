package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time" // Import the time package for time tracking

	"github.com/fatih/color"
	"github.com/rifsxd/dvpl_lz4/common/dvpl_logic"
)

var GlobalPath string

const (
	dvplExtension = ".dvpl"
)

// ANSI escape codes for text coloring
const (
	RedColor    = "\033[31m"
	GreenColor  = "\033[32m"
	YellowColor = "\033[33m"
	ResetColor  = "\033[0m"
)

// Config represents the configuration for the program.
type Config struct {
	Mode          string
	KeepOriginals bool
	Path          string // New field to specify the directory path.
	Ignore        string
	IgnoreExt     bool
}

// DVPLFooter represents the DVPL file footer data.
type DVPLFooter struct {
	OriginalSize   uint32
	CompressedSize uint32
	CRC32          uint32
	Type           uint32
}

// Info variables
const Dev = "RifsxD"
const Name = "DVPL_LZ4 CLI TOOL"
const Version = "1.2.0"
const Repo = "https://github.com/rifsxd/dvpl_lz4"
const Web = "https://rxd-mods.xyz"
const Commit = "07/03/2024"
const Info = "A CLI/GUI Tool Coded In GoLang To Convert WoTB ( Dava ) SmartDLC DVPL File Based On LZ4 High Compression."

func Cli() {

	cyan := color.New(color.FgCyan)

	fmt.Println()
	cyan.Println("• Name:", Name)
	cyan.Println("• Version:", Version)
	cyan.Println("• Commit:", Commit)
	cyan.Println("• Dev:", Dev)
	cyan.Println("• Repo:", Repo)
	cyan.Println("• Web:", Web)
	cyan.Println("• Info:", Info)
	fmt.Println()

	startTime := time.Now() // Record start time

	config, err := parseCommandLineArgs()
	if err != nil {
		log.Printf("%sError%s parsing command-line arguments: %v -> %sFallback to GUI mode!%s", RedColor, ResetColor, err, YellowColor, ResetColor)
		Gui() // Fallback to GUI mode if parseCommandLineArgs() errors out
		return
	}

	switch config.Mode {
	case "compress", "decompress":
		successCount, failureCount, ignoredCount, err := processFiles(config.Path, config)
		if err != nil {
			log.Printf("%s%s FAILED%s: %v", RedColor, strings.ToUpper(config.Mode), ResetColor, err)
		} else {
			log.Printf("%s%s FINISHED%s. Successful conversions: %s%d%s, Failed conversions: %s%d%s, Ignored conversions: %s%d%s", GreenColor, strings.ToUpper(config.Mode), ResetColor, GreenColor, successCount, ResetColor, RedColor, failureCount, ResetColor, YellowColor, ignoredCount, ResetColor)
		}
	case "gui":
		runGui() // Call the GUI mode
	case "help":
		printHelpMessage()
	default:
		log.Fatalf("%sIncorrect mode selected. Use '-help' for information.%s", RedColor, ResetColor)
	}

	elapsedTime := time.Since(startTime) // Calculate elapsed time
	printElapsedTime(elapsedTime)
}

func printElapsedTime(elapsedTime time.Duration) {
	var colorCode string

	// Determine the time unit and color
	switch {
	case elapsedTime.Seconds() < 1:
		colorCode = GreenColor // Milliseconds
		fmt.Printf("Processing took %s%d ms%s\n", colorCode, int(elapsedTime.Round(time.Millisecond).Milliseconds()), ResetColor)
		return
	case elapsedTime.Minutes() < 1:
		colorCode = YellowColor // Seconds
		fmt.Printf("Processing took %s%d s%s\n", colorCode, int(elapsedTime.Round(time.Second).Seconds()), ResetColor)
		return
	default:
		colorCode = RedColor // Minutes
		fmt.Printf("Processing took %s%d min%s\n", colorCode, int(elapsedTime.Round(time.Minute).Minutes()), ResetColor)
		return
	}
}

func parseCommandLineArgs() (*Config, error) {
	config := &Config{}
	flag.StringVar(&config.Mode, "mode", "", "Mode can be 'compress' / 'decompress' / 'help' (for an extended help guide).")
	flag.BoolVar(&config.KeepOriginals, "keep-originals", false, "Keep original files after compression/decompression.")
	flag.StringVar(&config.Path, "path", "", "directory/files path to process. Default is the current directory.")
	flag.StringVar(&config.Ignore, "ignore", "", "Comma-separated list of file extensions to ignore during compression.")

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

func printHelpMessage() {
	fmt.Println(`dvpl_lz4 [-mode] [-keep-originals] [-path]

    • mode can be one of the following:

        compress: compresses files into dvpl.
        decompress: decompresses dvpl files into standard files.
		gui: opens the graphical user interface window.
        help: show this help message.

	• flags can be one of the following:

    	-keep-originals flag keeps the original files after compression/decompression.
		-path specifies the directory/files path to process. Default is the current directory.
		-ignore specifies comma-separated file extensions to ignore during compression.

	• usage can be one of the following examples:

		$ dvpl_lz4 -mode help
		
		$ dvpl_lz4 -mode gui

		$ dvpl_lz4 -mode gui -path /path/to/decompress

		$ dvpl_lz4 -mode decompress -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode decompress -keep-originals -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode compress -keep-originals -path /path/to/decompress/compress
		
		$ dvpl_lz4 -mode decompress -path /path/to/decompress/compress.yaml.dvpl
		
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress.yaml
		
		$ dvpl_lz4 -mode decompress -keep-originals -path /path/to/decompress/compress.yaml.dvpl
		
		$ dvpl_lz4 -mode dcompress -keep-originals -path /path/to/decompress/compress.yaml

		$ dvpl_lz4 -mode compress -path /path/to/decompress -ignore .exe,.dll
	`)
}

func processFiles(directoryOrFile string, config *Config) (successCount, failureCount, ignoredCount int, err error) {
	// Initialize counters
	successCount = 0
	failureCount = 0
	ignoredCount = 0

	info, err := os.Stat(directoryOrFile)
	if err != nil {
		return 0, 0, 0, err
	}

	if info.IsDir() {
		dirList, err := os.ReadDir(directoryOrFile)
		if err != nil {
			return 0, 0, 0, err
		}

		for _, dirItem := range dirList {
			succ, fail, ignored, err := processFiles(filepath.Join(directoryOrFile, dirItem.Name()), config)
			if err != nil {
				fmt.Printf("Error processing directory %s: %v\n", dirItem.Name(), err)
			}
			successCount += succ
			failureCount += fail
			ignoredCount += ignored
		}
	} else {
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
				fmt.Printf("%sError%s reading file %s: %v\n", RedColor, ResetColor, directoryOrFile, err)
				return 0, 0, 0, err
			}

			var processedBlock []byte
			newName := ""

			if isCompression {
				processedBlock, err = dvpl_logic.CompressDVPL(fileData)
				newName = directoryOrFile + dvplExtension
			} else {
				processedBlock, err = dvpl_logic.DecompressDVPL(fileData)
				newName = strings.TrimSuffix(directoryOrFile, dvplExtension)
			}

			if err != nil {
				fmt.Printf("%sFile%s %s %sfailed to convert due to %v%s\n", RedColor, ResetColor, directoryOrFile, RedColor, err, ResetColor)
				return 0, 1, 0, nil // Return failure count as 1 for this file
			}

			err = os.WriteFile(newName, processedBlock, 0644)
			if err != nil {
				fmt.Printf("%sError%s writing file %s: %v\n", RedColor, ResetColor, newName, err)
				return 0, 0, 0, err
			}

			fmt.Printf("%sFile%s %s has been successfully %s into %s%s%s\n", GreenColor, ResetColor, filePath, getAction(config.Mode), GreenColor, newName, ResetColor)

			if !config.KeepOriginals {
				err := os.Remove(filePath)
				if err != nil {
					fmt.Printf("%sError%s deleting file %s: %v\n", RedColor, ResetColor, filePath, err)
				}
			}

			successCount++
		} else {
			fmt.Printf("%sIgnoring%s file %s\n", YellowColor, ResetColor, directoryOrFile)
			ignoredCount++
		}
	}

	return successCount, failureCount, ignoredCount, nil
}

func getAction(mode string) string {
	if mode == "compress" {
		return GreenColor + "compressed" + ResetColor
	}
	return GreenColor + "decompressed" + ResetColor
}

func runGui() {
	Gui()
}
