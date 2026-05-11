package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hurleybirdjr/album2vid-gui/internal/engine"
	"github.com/hurleybirdjr/album2vid-gui/internal/ffmpegbin"
)

func extractEmbedded(tmpd string) (string, string) {
	ffmpegName := "ffmpeg"
	ffprobeName := "ffprobe"
	if runtime.GOOS == "windows" {
		ffmpegName = "ffmpeg.exe"
		ffprobeName = "ffprobe.exe"
	}

	ffmpegPath := filepath.Join(tmpd, ffmpegName)
	ffprobePath := filepath.Join(tmpd, ffprobeName)

	if err := os.WriteFile(ffmpegPath, ffmpegbin.FFmpeg, 0755); err != nil {
		throwError("Failed to extract ffmpeg: " + err.Error())
	}
	if err := os.WriteFile(ffprobePath, ffmpegbin.FFprobe, 0755); err != nil {
		throwError("Failed to extract ffprobe: " + err.Error())
	}

	return ffmpegPath, ffprobePath
}

func throwError(text string) {
	_, err := fmt.Fprintln(os.Stderr, "ERROR: "+text)
	if err != nil {
		return
	}
	os.Exit(1)
}

func cleanup(tempDir string) {
	err := os.RemoveAll(tempDir)
	if err != nil {
		return
	}
}

func main() {
	fast := flag.Bool("f", false, "Enables fast mode, may cause rendering errors")
	flag.BoolVar(fast, "fast", false, "Enables fast mode, may cause rendering errors")
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Usage: album2vid [options] <path>\n")
		if err != nil {
			return
		}
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	srcPath := flag.Arg(0)
	info, err := os.Stat(srcPath)
	if err != nil || !info.IsDir() {
		throwError("The directory could not be found")
	}

	tempDir, err := os.MkdirTemp("", "album2vid")
	if err != nil {
		throwError("Could not create temp directory: " + err.Error())
	}
	defer cleanup(tempDir)

	ffmpegPath, ffprobePath := extractEmbedded(tempDir)
	fmt.Println("Welcome to album2vid!")

	eng := &engine.Engine{
		FFmpegBin:  ffmpegPath,
		FFprobeBin: ffprobePath,
		OnProgress: func(msg string) {
			fmt.Println(msg)
		},
	}

	if err := eng.ProcessAlbum(context.Background(), srcPath, *fast, tempDir); err != nil {
		throwError(err.Error())
	}
}
