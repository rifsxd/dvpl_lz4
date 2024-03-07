package cmd

import (
	"fmt"
	"log"
	"strings"
	"time" // Import the time package for time tracking

	"github.com/fatih/color"
	"github.com/rifsxd/dvpl_lz4/common/colors"
	"github.com/rifsxd/dvpl_lz4/common/meta"
	"github.com/rifsxd/dvpl_lz4/common/utils"
)

func Cli() {

	cyan := color.New(color.FgCyan)

	fmt.Println()
	cyan.Println("• Name:", meta.Name)
	cyan.Println("• Version:", meta.Version)
	cyan.Println("• Commit:", meta.Commit)
	cyan.Println("• Dev:", meta.Dev)
	cyan.Println("• Repo:", meta.Repo)
	cyan.Println("• Web:", meta.Web)
	cyan.Println("• Info:", meta.Info)
	fmt.Println()

	startTime := time.Now() // Record start time

	config, err := utils.ParseCommandLineArgs()
	if err != nil {
		log.Printf("%sError%s parsing command-line arguments: %v -> %sFallback to GUI mode!%s", colors.RedColor, colors.ResetColor, err, colors.YellowColor, colors.ResetColor)
		Gui() // Fallback to GUI mode if parseCommandLineArgs() errors out
		return
	}

	switch config.Mode {
	case "compress", "decompress":
		successCount, failureCount, ignoredCount, err := utils.ProcessFiles(config.Path, config)
		if err != nil {
			log.Printf("%s%s FAILED%s: %v", colors.RedColor, strings.ToUpper(config.Mode), colors.ResetColor, err)
		} else {
			log.Printf("%s%s FINISHED%s. Successful conversions: %s%d%s, Failed conversions: %s%d%s, Ignored conversions: %s%d%s", colors.GreenColor, strings.ToUpper(config.Mode), colors.ResetColor, colors.GreenColor, successCount, colors.ResetColor, colors.RedColor, failureCount, colors.ResetColor, colors.YellowColor, ignoredCount, colors.ResetColor)
		}
	case "gui":
		runGui() // Call the GUI mode
	case "help":
		utils.PrintHelpMessage()
	default:
		log.Fatalf("%sIncorrect mode selected. Use '-help' for information.%s", colors.RedColor, colors.ResetColor)
	}

	elapsedTime := time.Since(startTime) // Calculate elapsed time
	utils.PrintElapsedTime(elapsedTime)
}

func runGui() {
	Gui()
}
