package cmd

import (
	"embed"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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

	config := &Config{}

	/* Check if command-line arguments were provided
	if len(os.Args) > 1 {
		// Use the provided path as the initial path
		config.Path = os.Args[1]
	} else {
		// Get the current working directory and set it as the initial path
		initialPath, err := os.Getwd()
		if err != nil {
			// Handle the error, e.g., show a message to the user
			initialPath = "" // Default to an empty string if there's an error
		}
		config.Path = initialPath
	} */

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
	pathEntry.SetText(config.Path)
	pathEntry.SetPlaceHolder("Enter directory or file path")
	pathEntry.OnChanged = func(path string) {
		config.Path = path
	}

	// Create a custom success dialog
	successDialog := dialog.NewCustom("Success", "OK", createSuccessContent(), myWindow)
	successDialog.SetDismissText("OK")

	content := container.NewVBox(
		widget.NewLabelWithStyle("DVPL_LZ4 GUI TOOL • "+Version, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
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

func createSuccessContent() fyne.CanvasObject {
	successLabel := widget.NewLabelWithStyle("Conversion completed successfully", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewVBox(
		successLabel,
	)

	return content
}

func convertFiles(myWindow fyne.Window, config *Config) {
	successCount, failureCount, ignoredCount, err := processFiles(config.Path, config)
	if err != nil {
		dialog.NewError(err, myWindow)
		return
	}

	successContent := fmt.Sprintf("Successful conversions: %d\nFailed conversions: %d\nIgnored conversions: %d", successCount, failureCount, ignoredCount)
	successDialog := dialog.NewInformation("Conversion Results", successContent, myWindow)
	successDialog.SetDismissText("OK")

	successDialog.Show()
}
