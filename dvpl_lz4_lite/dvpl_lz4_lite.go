//go:generate goversioninfo -64

package main

import (
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pierrec/lz4/v4"
)

// ANSI escape codes for text coloring
const (
	RedColor    = "\033[31m"
	GreenColor  = "\033[32m"
	YellowColor = "\033[33m"
	ResetColor  = "\033[0m"
)

// Info variables
const Dev = "RifsxD"
const Name = "DVPL_LZ4 CLI TOOL"
const Version = "1.2.4-lite"
const Repo = "https://github.com/rifsxd/dvpl_lz4"
const Web = "https://rxd-mods.xyz"
const Commit = "08/03/2024"
const Info = "A CLI Tool Coded In GoLang To Convert WoTB ( Dava ) SmartDLC DVPL File Based On LZ4 High Compression."

// Constants related to DVPL format
const (
	dvplFooterSize = 20
	dvplTypeNone   = 0
	dvplTypeLZ4    = 2
	dvplFooter     = "DVPL"
)

// createDVPLFooter creates a DVPL footer from the provided data.
func createDVPLFooter(inputSize, compressedSize, crc32, typeVal uint32) []byte {
	result := make([]byte, dvplFooterSize)
	writeLittleEndianUint32(result, inputSize, 0)
	writeLittleEndianUint32(result, compressedSize, 4)
	writeLittleEndianUint32(result, crc32, 8)
	writeLittleEndianUint32(result, typeVal, 12)
	copy(result[16:], dvplFooter)
	return result
}

// readDVPLFooter reads the DVPL footer data from a DVPL buffer.
func readDVPLFooter(buffer []byte) (*DVPLFooter, error) {
	footerBuffer := buffer[len(buffer)-dvplFooterSize:]
	if string(footerBuffer[16:]) != dvplFooter || len(footerBuffer) != dvplFooterSize {
		return nil, errors.New(RedColor + "InvalidDVPLFooter" + ResetColor)
	}

	footerData := &DVPLFooter{}
	footerData.OriginalSize = readLittleEndianUint32(footerBuffer, 0)
	footerData.CompressedSize = readLittleEndianUint32(footerBuffer, 4)
	footerData.CRC32 = readLittleEndianUint32(footerBuffer, 8)
	footerData.Type = readLittleEndianUint32(footerBuffer, 12)
	return footerData, nil
}

// writeLittleEndianUint32 writes a little-endian uint32 value to a byte slice at the specified offset.
func writeLittleEndianUint32(b []byte, v uint32, offset int) {
	b[offset+0] = byte(v)
	b[offset+1] = byte(v >> 8)
	b[offset+2] = byte(v >> 16)
	b[offset+3] = byte(v >> 24)
}

// readLittleEndianUint32 reads a little-endian uint32 value from a byte slice at the specified offset.
func readLittleEndianUint32(b []byte, offset int) uint32 {
	return uint32(b[offset]) | uint32(b[offset+1])<<8 | uint32(b[offset+2])<<16 | uint32(b[offset+3])<<24
}

var GlobalPath string

const (
	dvplExtension = ".dvpl"
)

// DVPLFooter represents the footer structure of a DVPL file
type DVPLFooter struct {
	OriginalSize   uint32 // Original size of the data
	CompressedSize uint32 // Compressed size of the data
	CRC32          uint32 // CRC32 checksum of the data
	Type           uint32 // Type of compression used (0 - None, 2 - LZ4)
}

// CompressDVPL compresses a buffer and returns the processed DVPL file buffer.
func CompressDVPL(buffer []byte) ([]byte, error) {
	// Calculate the maximum possible compressed block size
	compressedBlockSize := lz4.CompressBlockBound(len(buffer))
	compressedBlock := make([]byte, compressedBlockSize)

	// Compress the data
	n, err := lz4.CompressBlock(buffer, compressedBlock, nil)
	if err != nil {
		return nil, err
	}

	// Trim the slice to actual compressed size
	compressedBlock = compressedBlock[:n]

	// Create DVPL footer
	footerBuffer := createDVPLFooter(uint32(len(buffer)), uint32(n), crc32.ChecksumIEEE(compressedBlock), dvplTypeLZ4)

	// Append footer to the compressed data
	return append(compressedBlock, footerBuffer...), nil
}

// DecompressDVPL decompresses a DVPL buffer and returns the uncompressed file buffer.
func DecompressDVPL(buffer []byte) ([]byte, error) {
	// Read DVPL footer
	footerData, err := readDVPLFooter(buffer)
	if err != nil {
		return nil, err
	}

	// Extract compressed block
	targetBlock := buffer[:len(buffer)-dvplFooterSize]

	// Check if compressed size matches the footer
	if uint32(len(targetBlock)) != footerData.CompressedSize {
		return nil, errors.New(RedColor + "DVPLSizeMismatch" + ResetColor)
	}

	// Check CRC32 checksum
	if crc32.ChecksumIEEE(targetBlock) != footerData.CRC32 {
		return nil, errors.New(RedColor + "DVPLCRC32Mismatch" + ResetColor)
	}

	// Decompress based on compression type
	if footerData.Type == dvplTypeNone {
		// No compression applied, return the block as is
		if footerData.OriginalSize != footerData.CompressedSize || footerData.Type != dvplTypeNone {
			return nil, errors.New(RedColor + "DVPLTypeSizeMismatch" + ResetColor)
		}
		return targetBlock, nil
	} else if footerData.Type == dvplTypeLZ4 {
		// LZ4 compression, decompress the block
		deDVPLBlock := make([]byte, footerData.OriginalSize)
		n, err := lz4.UncompressBlock(targetBlock, deDVPLBlock)
		if err != nil {
			return nil, err
		}

		// Check if decompressed size matches the footer
		if uint32(n) != footerData.OriginalSize {
			return nil, errors.New(RedColor + "DVPLDecodeSizeMismatch" + ResetColor)
		}

		return deDVPLBlock, nil
	}

	// Unknown compression type
	return nil, errors.New(RedColor + "UNKNOWN DVPL FORMAT" + ResetColor)
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

	if info.IsDir() {
		dirList, err := os.ReadDir(directoryOrFile)
		if err != nil {
			return 0, 0, 0, err
		}

		for _, dirItem := range dirList {
			succ, fail, ignored, err := VerifyDVPLFiles(filepath.Join(directoryOrFile, dirItem.Name()), config)
			if err != nil {
				fmt.Printf("Error processing directory %s: %v\n", dirItem.Name(), err)
			}
			successCount += succ
			failureCount += fail
			ignoredCount += ignored
		}
	} else {
		// Ignore non-.dvpl files during verification
		if !strings.HasSuffix(directoryOrFile, dvplExtension) {
			fmt.Printf("%sIgnoring%s file %s\n", YellowColor, ResetColor, directoryOrFile)
			ignoredCount++
			return successCount, failureCount, ignoredCount, nil
		}

		filePath := directoryOrFile
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("%sError%s reading file %s: %v\n", RedColor, ResetColor, directoryOrFile, err)
			return 0, 0, 0, err
		}

		_, err = DecompressDVPL(fileData)
		if err != nil {
			fmt.Printf("%sFile%s %s %sfailed to verify due to %v%s\n", RedColor, ResetColor, directoryOrFile, RedColor, err, ResetColor)
			return 0, 1, 0, nil // Return failure count as 1 for this file
		}

		fmt.Printf("%sFile%s %s has been successfully %s\n", GreenColor, ResetColor, filePath, getAction(config.Mode))

		successCount++
	}

	return successCount, failureCount, ignoredCount, nil
}

// Config represents the configuration for the program.
type Config struct {
	Mode          string
	KeepOriginals bool
	Path          string // New field to specify the directory path.
	Ignore        string
	IgnoreExt     bool
}

func PrintElapsedTime(elapsedTime time.Duration) {
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

func ParseCommandLineArgs() (*Config, error) {
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

func PrintHelpMessage() {
	fmt.Println(`dvpl_lz4 [-mode] [-keep-originals] [-path]

    • mode can be one of the following:

        compress: compresses files into dvpl.
        decompress: decompresses dvpl files into standard files.
        help: show this help message.

	• flags can be one of the following:

    	-keep-originals flag keeps the original files after compression/decompression.
		-path specifies the directory/files path to process. Default is the current directory.
		-ignore specifies comma-separated file extensions to ignore during compression.

	• usage can be one of the following examples:

		$ dvpl_lz4 -mode help

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

func getAction(mode string) string {
	if mode == "compress" {
		return GreenColor + "compressed" + ResetColor
	}
	return GreenColor + "decompressed" + ResetColor
}

func ProcessFiles(directoryOrFile string, config *Config) (successCount, failureCount, ignoredCount int, err error) {
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
			succ, fail, ignored, err := ProcessFiles(filepath.Join(directoryOrFile, dirItem.Name()), config)
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
				processedBlock, err = CompressDVPL(fileData)
				newName = directoryOrFile + dvplExtension
			} else {
				processedBlock, err = DecompressDVPL(fileData)
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

func main() {

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

	config, err := ParseCommandLineArgs()
	if err != nil {
		log.Printf("%sError%s parsing command-line arguments: %v", RedColor, ResetColor, err)

		return
	}

	switch config.Mode {
	case "compress", "decompress":
		successCount, failureCount, ignoredCount, err := ProcessFiles(config.Path, config)
		if err != nil {
			log.Printf("%s%s FAILED%s: %v", RedColor, strings.ToUpper(config.Mode), ResetColor, err)
		} else {
			log.Printf("%s%s FINISHED%s. Successful conversions: %s%d%s, Failed conversions: %s%d%s, Ignored conversions: %s%d%s", GreenColor, strings.ToUpper(config.Mode), ResetColor, GreenColor, successCount, ResetColor, RedColor, failureCount, ResetColor, YellowColor, ignoredCount, ResetColor)
		}
	case "verify":
		successCount, failureCount, ignoredCount, err := VerifyDVPLFiles(config.Path, config)
		if err != nil {
			log.Printf("%s%s FAILED%s: %v", RedColor, strings.ToUpper(config.Mode), ResetColor, err)
		} else {
			log.Printf("%s%s FINISHED%s. Successful verifications: %s%d%s, Failed verifications: %s%d%s, Ignored files: %s%d%s", GreenColor, strings.ToUpper(config.Mode), ResetColor, GreenColor, successCount, ResetColor, RedColor, failureCount, ResetColor, YellowColor, ignoredCount, ResetColor)
		}
	case "help":
		PrintHelpMessage()
	default:
		log.Fatalf("%sIncorrect mode selected. Use '-help' for information.%s", RedColor, ResetColor)
	}

	elapsedTime := time.Since(startTime) // Calculate elapsed time
	PrintElapsedTime(elapsedTime)
}
