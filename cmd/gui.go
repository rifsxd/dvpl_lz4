package cmd

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/rifsxd/dvpl_lz4/common/meta"
	"github.com/rifsxd/dvpl_lz4/common/utils"
)

//go:embed resource/dvpl_lz4.png
var resources embed.FS

func Gui() {
	myApp := app.NewWithID("xyz.rxd.dvpl_lz4")
	myWindow := myApp.NewWindow("DVPL_LZ4 GUI CONVERTER")

	// Load the embedded image
	iconData, _ := resources.ReadFile("resource/dvpl_lz4.png")
	iconResource := fyne.NewStaticResource("dvpl_lz4.png", iconData)
	myWindow.SetIcon(iconResource)

	config := &utils.Config{}

	// Parse command-line arguments
	flag.Parse()

	// Check if the GlobalPath variable is empty
	if utils.GlobalPath == "" {
		// If it is, get the current directory
		if initialPath, err := os.Getwd(); err == nil {
			utils.GlobalPath = initialPath
		}
	}

	// Set the path to the value of the global variable
	config.Path = utils.GlobalPath

	compressButton := widget.NewButton("Compress", func() {
		config.Mode = "compress"
		convertFiles(myWindow, config) // Pass myWindow as a parameter
	})

	decompressButton := widget.NewButton("Decompress", func() {
		config.Mode = "decompress"
		convertFiles(myWindow, config) // Pass myWindow as a parameter
	})

	keepOriginalsCheck := widget.NewCheck("Keep Originals", func(keep bool) {
		config.KeepOriginals = keep
	})

	ignoreCheck := widget.NewCheck("Ignore Extensions", func(ignore bool) {
		config.IgnoreExt = ignore
	})

	ignoreEntry := widget.NewEntry()
	ignoreEntry.SetPlaceHolder("Enter comma-separated extensions to ignore")
	ignoreEntry.OnChanged = func(ext string) {
		config.Ignore = ext
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetText(utils.GlobalPath) // Set the text to the value of the global variable
	pathEntry.SetPlaceHolder("Enter directory or file path")
	pathEntry.OnChanged = func(path string) {
		config.Path = path
	}

	content := container.NewVBox(
		widget.NewLabelWithStyle("DVPL_LZ4 GUI TOOL • "+meta.Version, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewHBox(layout.NewSpacer(), compressButton, decompressButton, layout.NewSpacer()),
		widget.NewForm(
			widget.NewFormItem("Options:", keepOriginalsCheck),
			widget.NewFormItem("Ignore:", ignoreCheck),
			widget.NewFormItem("Extensions:", ignoreEntry),
			widget.NewFormItem("Path:", pathEntry),
		),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(500, 200))
	myWindow.ShowAndRun()
}

func convertFiles(myWindow fyne.Window, config *utils.Config) {
	startTime := time.Now() // Record start time

	successCount, failureCount, ignoredCount, err := utils.ProcessFiles(config.Path, config)
	if err != nil {
		dialog.NewError(err, myWindow)
		return
	}

	elapsedTime := time.Since(startTime) // Calculate elapsed time

	successContent := fmt.Sprintf("Successful conversions: %d\nFailed conversions: %d\nIgnored conversions: %d\n\nTime taken: %s", successCount, failureCount, ignoredCount, formatElapsedTime(elapsedTime))
	successDialog := dialog.NewInformation("Conversion Results", successContent, myWindow)
	successDialog.SetDismissText("OK")

	successDialog.Show()
}

func formatElapsedTime(elapsedTime time.Duration) string {
	seconds := int(elapsedTime.Round(time.Second).Seconds())
	minutes := seconds / 60
	seconds %= 60
	milliseconds := int(elapsedTime.Round(time.Millisecond).Milliseconds())

	if minutes > 0 {
		return fmt.Sprintf("%d min %d sec", minutes, seconds)
	} else if seconds > 0 {
		return fmt.Sprintf("%d sec", seconds)
	}
	return fmt.Sprintf("%d ms", milliseconds)
}
