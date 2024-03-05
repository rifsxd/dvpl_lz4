# Dvpl_Lz4 Cli/Gui Converter
- A Cli/Gui Tool Coded In Golang To Convert WoTB ( Dava ) SmartDLC DVPL File Based On LZ4 High Compression.

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

	- usage can be one of the following examples:

		```
		$ dvpl_lz4 -mode help
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
		$ dvpl_lz4 -mode compress -path /path/to/decompress/compress.yaml -ignore .exe,.dll
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
