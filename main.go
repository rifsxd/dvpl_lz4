//go:generate goversioninfo -64

/*
Copyright © 2024 RXD - MODS | support@rxd-mods.xyx
*/

package main

import "github.com/rifsxd/dvpl_lz4/cmd"

func main() {
	cmd.Cli() // change it to 'cmd.Gui()' for it to lauch gui app directly instead from gui mode flag
}
