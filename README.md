# Dvpl_Lz4 Cli/Gui Converter
- A Cli/Gui Tool Coded In Golang To Convert WoTB ( Dava ) SmartDLC DVPL File Based On LZ4 High Compression.

<div align="center">
	
[![GitHub license](https://img.shields.io/github/license/rifsxd/dvpl_lz4?logo=apache&label=License&style=flat)](https://github.com/rifsxd/dvpl_lz4/blob/master/LICENSE)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/rifsxd/dvpl_lz4/total?logo=github&label=Downloads&style=flat)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/rifsxd/dvpl_lz4?style=flat&label=Code%20Size)
[![GitHub Debug CI Status](https://img.shields.io/github/actions/workflow/status/rifsxd/dvpl_lz4/build.yml?logo=github&label=Beta%20CI&style=flat)](https://github.com/rifsxd/dvpl_lz4/actions/workflows/build.yml)

</div>

Usage :

  - dvpl_lz4 [-mode] [-keep-originals] [-path] [-ignore]

    - mode can be one of the following:

        compress: compresses files into dvpl.
        decompress: decompresses dvpl files into standard files.
		gui: opens the graphical user interface window.
        help: show this help message.

	- flags can be one of the following:

    	-keep-originals flag keeps the original files after compression/decompression.
		-path specifies the directory/files path to process. Default is the current directory.
		-ignore specifies comma-separated file extensions to ignore during compression.
		-silent disables all file processing verbose information
		
	- usage can be one of the following examples:

		```
		$ dvpl_lz4 -mode help
		```
		```
		$ dvpl_lz4 -mode gui
		```
		```
		$ dvpl_lz4 -mode gui -path /path/to/decompress
		```
		```
		$ dvpl_lz4 -mode decompress -path /path/to/decompress/compress
		```
		```
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress
		```
		```
		$ dvpl_lz4 -mode decompress -keep-originals -path /path/to/decompress/compress
		```
		```
		$ dvpl_lz4 -mode compress -keep-originals -path /path/to/decompress/compress
		```
		```
		$ dvpl_lz4 -mode decompress -path /path/to/decompress/compress.yaml.dvpl
		```
		```
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress.yaml
		```
		```
		$ dvpl_lz4 -mode decompress -keep-originals -path /path/to/decompress/compress.yaml.dvpl
		```
		```
		$ dvpl_lz4 -mode dcompress -keep-originals -path /path/to/decompress/compress.yaml
		```
		```
		$ dvpl_lz4 -mode compress -path /path/to/decompress -ignore .exe,.dll
		```
		```
		$ dvpl_lz4 -mode verify -path /path/to/verify/compress.yaml.dvpl
		```
		```
		$ dvpl_lz4 -mode verify -path /path/to/verify/
		```
		```
		$ dvpl_lz4 -mode dcompress -silent
		```
Building :

- go 1.20+ required!

```
$ git clone https://github.com/rifsxd/dvpl_lz4.git
```

```
$ cd dvpl_lz4
```

```
$ go build
```
