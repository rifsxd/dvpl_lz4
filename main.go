//go:generate goversioninfo -64

package main

import "github.com/rifsxd/dvpl_lz4/cmd"

func main() {
	cmd.Cli() // change it to 'cmd.Gui()' for it to launch gui app directly instead from gui mode flag
}
