package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hurleybirdjr/album2vid-gui/internal/engine"
	"github.com/hurleybirdjr/album2vid-gui/internal/ffmpegbin"
	"github.com/nfnt/resize"
)

func extractEmbedded(tmpd string) (string, string, error) {
	ffmpegName := "ffmpeg"
	ffprobeName := "ffprobe"
	if runtime.GOOS == "windows" {
		ffmpegName = "ffmpeg.exe"
		ffprobeName = "ffprobe.exe"
	}

	ffmpegPath := filepath.Join(tmpd, ffmpegName)
	ffprobePath := filepath.Join(tmpd, ffprobeName)

	if err := os.WriteFile(ffmpegPath, ffmpegbin.FFmpeg, 0755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(ffprobePath, ffmpegbin.FFprobe, 0755); err != nil {
		return "", "", err
	}

	return ffmpegPath, ffprobePath, nil
}

func loadResizedIcon(path string, maxSize uint) (fyne.Resource, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := uint(bounds.Dx()), uint(bounds.Dy())

	if width > maxSize || height > maxSize {
		img = resize.Thumbnail(maxSize, maxSize, img, resize.Lanczos3)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return fyne.NewStaticResource(filepath.Base(path), buf.Bytes()), nil
}

func main() {
	mainApp := app.NewWithID("com.album2vid.gui")
	mainWindow := mainApp.NewWindow("album2vid-gui")
	mainWindow.Resize(fyne.NewSize(1200, 800))
	if appIcon, err := loadResizedIcon("assets/icon.png", 128); err == nil {
		mainWindow.SetIcon(appIcon)
	}

	// persist temp dir for tools
	toolDir, err := os.MkdirTemp("", "album2vid.tools")
	if err != nil {
		dialog.ShowError(err, mainWindow)
	} else {
		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {

			}
		}(toolDir)
	}
	ffmpegPath, ffprobePath, _ := extractEmbedded(toolDir)

	eng := &engine.Engine{
		FFmpegBin:  ffmpegPath,
		FFprobeBin: ffprobePath,
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("No folder selected")

	statusLabel := widget.NewLabel("Ready")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	fastMode := widget.NewCheck("Fast Mode (may cause rendering errors)", nil)
	runButton := widget.NewButton("Generate Video", nil)

	// album info display widgets
	coverImage := canvas.NewImageFromResource(theme.ViewFullScreenIcon())
	coverImage.FillMode = canvas.ImageFillContain
	coverImage.SetMinSize(fyne.NewSize(400, 400))

	trackList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("Track Title") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)

	updateUI := func(info *engine.AlbumInfo) {
		if info == nil {
			coverImage.File = ""
			coverImage.Resource = theme.FileImageIcon()
			coverImage.Refresh()
			mainWindow.SetIcon(theme.ComputerIcon())
			trackList.Length = func() int { return 0 }
			trackList.Refresh()
			return
		}
		if info.CoverPath != "" {
			coverImage.File = info.CoverPath
			coverImage.Resource = nil
			coverImage.Refresh()

			//	// update window icon to cover art (I originally had this, but left the logo I made instead lol)
			//	if res, err := loadResizedIcon(info.CoverPath, 128); err == nil {
			//		mainWindow.SetIcon(res)
			//	}
		} else {
			coverImage.File = ""
			coverImage.Resource = theme.FileImageIcon()
			coverImage.Refresh()
			mainWindow.SetIcon(theme.ComputerIcon())
		}
		trackList.Length = func() int { return len(info.Tracks) }
		trackList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			t := info.Tracks[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%d. %s (%s)", id+1, t.Title, engine.GetTimestamp(t.Duration)))
		}
		trackList.Refresh()
	}

	scanFolder := func(path string) {
		if path == "" {
			runButton.Importance = widget.MediumImportance
			mainWindow.Content().Refresh()
			return
		}
		info, err := eng.ScanAlbum(context.Background(), path)
		if err != nil {
			runButton.Importance = widget.MediumImportance
			updateUI(info)
			mainWindow.Content().Refresh()
			dialog.ShowError(err, mainWindow)
			return
		}
		runButton.Importance = widget.SuccessImportance
		updateUI(info)
		mainWindow.Content().Refresh()
	}

	pathEntry.OnSubmitted = func(s string) {
		scanFolder(s)
	}

	// folder picker dialog :)
	browseButton := widget.NewButtonWithIcon("Browse", theme.FolderIcon(), func() {
		d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}
			if list == nil {
				return
			}
			pathEntry.SetText(list.Path())
			scanFolder(list.Path())
		}, mainWindow)

		cwd, _ := os.Getwd()
		if uri, err := storage.ListerForURI(storage.NewFileURI(cwd)); err == nil {
			d.SetLocation(uri)
		}

		d.Resize(fyne.NewSize(1000, 600))
		d.Show()
	})

	var cancelFunc context.CancelFunc
	var isRunning bool

	runButton.OnTapped = func() {
		if isRunning {
			if cancelFunc != nil {
				cancelFunc()
			}
			return
		}

		srcPath := pathEntry.Text
		if srcPath == "" {
			dialog.ShowInformation("Error", "Please select an album folder first.", mainWindow)
			return
		}

		isRunning = true
		runButton.SetText("Cancel")
		runButton.Importance = widget.DangerImportance
		browseButton.Disable()
		pathEntry.Disable()
		fastMode.Disable()

		var ctx context.Context
		ctx, cancelFunc = context.WithCancel(context.Background())

		eng.OnProgress = func(msg string) {
			statusLabel.SetText(msg)
		}
		eng.OnNumeric = func(val float64) {
			progressBar.SetValue(val)
		}

		go func() {
			progressBar.Show()
			progressBar.SetValue(0)
			defer func() {
				isRunning = false
				runButton.SetText("Generate Video")
				runButton.Importance = widget.MediumImportance
				runButton.Enable()
				browseButton.Enable()
				pathEntry.Enable()
				fastMode.Enable()
				progressBar.Hide()
				cancelFunc = nil
			}()

			tempDir, err := os.MkdirTemp("", "album2vid.run")
			if err != nil {
				statusLabel.SetText("Error: " + err.Error())
				return
			}
			defer func(path string) {
				err := os.RemoveAll(path)
				if err != nil {

				}
			}(tempDir)

			err = eng.ProcessAlbum(ctx, srcPath, fastMode.Checked, tempDir)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					statusLabel.SetText("Rendering cancelled.")
				} else {
					statusLabel.SetText("Error: " + err.Error())
					dialog.ShowError(err, mainWindow)
				}
			} else {
				statusLabel.SetText("Success! Video generated.")

				// button to take you to the album folder after rendering, for those that want that :)
				openFolderBtn := widget.NewButtonWithIcon("Open Album Folder", theme.FolderOpenIcon(), func() {
					parsed, _ := url.Parse("file://" + srcPath)
					err := mainApp.OpenURL(parsed)
					if err != nil {
						return
					}
				})

				d := dialog.NewCustom("Done", "Close", container.NewVBox(
					widget.NewLabel("Video and tracklist.txt generated successfully in the album folder."),
					openFolderBtn,
				), mainWindow)
				d.Show()
			}
		}()
	}

	infoSection := container.NewHSplit(
		container.NewBorder(
			widget.NewLabelWithStyle("Album Art", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			container.NewCenter(coverImage),
		),
		container.NewBorder(
			widget.NewLabelWithStyle("Tracklist", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			trackList,
		),
	)
	infoSection.Offset = 0.4

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("album2vid", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Album Folder Path:"),
			container.NewBorder(nil, nil, nil, browseButton, pathEntry),
			fastMode,
		),
		container.NewVBox(
			widget.NewSeparator(),
			runButton,
			progressBar,
			container.NewCenter(statusLabel),
		),
		nil, nil,
		infoSection,
	)

	mainWindow.SetContent(container.NewPadded(content))

	// refresh content after loading, this seemed to fix graphical bugs I had with x11 wayland windows
	go func() {
		mainWindow.Content().Refresh()
	}()

	mainWindow.ShowAndRun()
}
