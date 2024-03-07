package dvpl

import (
	"errors"
	"hash/crc32"

	"github.com/pierrec/lz4/v4"
)

// ANSI color codes for pretty printing
const (
	RedColor    = "\033[31m"
	GreenColor  = "\033[32m"
	YellowColor = "\033[33m"
	ResetColor  = "\033[0m"
)

// Constants related to DVPL format
const (
	dvplFooterSize = 20
	dvplTypeNone   = 0
	dvplTypeLZ4    = 2
	dvplFooter     = "DVPL"
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
