package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Track struct {
	Title    string
	Duration float64
}

type AlbumInfo struct {
	Tracks        []Track
	CoverPath     string
	TotalDuration float64
}

type Engine struct {
	FFmpegBin  string
	FFprobeBin string
	OnProgress func(string)
	OnNumeric  func(float64) // 0.0 to 1.0
}

func (e *Engine) log(msg string) {
	if e.OnProgress != nil {
		e.OnProgress(msg)
	}
}

func (e *Engine) setProgress(val float64) {
	if e.OnNumeric != nil {
		e.OnNumeric(val)
	}
}

func (e *Engine) GetRuntime(ctx context.Context, filename string) (float64, error) {
	cmd := exec.CommandContext(ctx, e.FFprobeBin,
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filename,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func GetTimestamp(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (e *Engine) PreprocessFiles(ctx context.Context, infiles []string, tmpd string) ([]string, error) {
	outfiles := make([]string, len(infiles))
	for i, inf := range infiles {
		e.log(fmt.Sprintf("Preprocessing track %d/%d...", i+1, len(infiles)))
		e.setProgress(float64(i) / float64(len(infiles)) * 0.5) // 0% to 50%
		base := filepath.Base(inf)
		name := strings.TrimSuffix(base, filepath.Ext(base)) + ".m4a"
		outf, _ := filepath.Abs(filepath.Join(tmpd, name))
		outfiles[i] = outf

		cmd := exec.CommandContext(ctx, e.FFmpegBin,
			"-i", inf,
			"-map", "0",
			"-map", "-v?",
			"-map", "-V?",
			"-acodec", "aac",
			"-b:a", "320k",
			outf,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to preprocess %s: %w", inf, err)
		}
	}
	return outfiles, nil
}

func GenFilelist(infiles []string, tmpd string) (string, error) {
	filename := filepath.Join(tmpd, "files.txt")
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	for _, file := range infiles {
		escaped := strings.ReplaceAll(file, "'", "'\\''")
		fmt.Fprintf(f, "file '%s'\n", escaped)
	}
	return filename, nil
}

func (e *Engine) GenTracklist(ctx context.Context, infiles []string, outfile string) error {
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	currTime := 0.0
	for _, file := range infiles {
		stem := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		fmt.Fprintf(f, "%s -- %s\n", stem, GetTimestamp(currTime))
		if runtime, err := e.GetRuntime(ctx, file); err == nil {
			currTime += runtime
		}
	}
	return nil
}

func GetCover(srcdir string) (string, error) {
	exts := []string{".jpg", ".jpeg", ".png", ".webp", ".bmp"}
	names := []string{"cover", "folder", "album", "art"}

	entries, err := os.ReadDir(srcdir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fname := strings.ToLower(entry.Name())
		ext := filepath.Ext(fname)
		base := strings.TrimSuffix(fname, ext)

		for _, name := range names {
			if base == name {
				for _, e := range exts {
					if ext == e {
						return filepath.Abs(filepath.Join(srcdir, entry.Name()))
					}
				}
			}
		}
	}

	// fallback if nothing found :(
	for _, name := range []string{"cover.jpg", "cover.png"} {
		candidate := filepath.Join(srcdir, name)
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Abs(candidate)
		}
	}

	return "", fmt.Errorf("the cover photo could not be found")
}

func GetSrcFiles(srcdir string) ([]string, error) {
	exts := map[string]bool{
		".wav": true, ".mp3": true, ".m4a": true, ".ogg": true, ".flac": true,
	}

	entries, err := os.ReadDir(srcdir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if exts[strings.ToLower(filepath.Ext(e.Name()))] {
			abs, _ := filepath.Abs(filepath.Join(srcdir, e.Name()))
			files = append(files, abs)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("the audio files could not be found")
	}

	sort.Slice(files, func(i, j int) bool {
		si := strings.TrimSuffix(filepath.Base(files[i]), filepath.Ext(files[i]))
		sj := strings.TrimSuffix(filepath.Base(files[j]), filepath.Ext(files[j]))
		return si < sj
	})

	return files, nil
}

func (e *Engine) MainFfmpegCall(ctx context.Context, filelist, cover string, totalDuration float64, outfile string) error {
	e.log("Generating final video...")
	e.setProgress(0.6) // start generation at 60%, yes I'm literally just picking numbers out of thin air
	cmd := exec.CommandContext(ctx, e.FFmpegBin,
		"-hwaccel", "auto",
		"-y",
		"-loop", "1",
		"-framerate", "1",
		"-i", cover,
		"-f", "concat",
		"-safe", "0",
		"-i", filelist,
		"-tune", "stillimage",
		"-t", fmt.Sprintf("%.2f", totalDuration),
		"-vf", "format=yuv420p",
		"-s", "1080x1080",
		"-b:a", "320k",
		outfile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *Engine) GetTotalDuration(ctx context.Context, infiles []string) float64 {
	total := 0.0
	for _, file := range infiles {
		if runtime, err := e.GetRuntime(ctx, file); err == nil {
			total += runtime
		}
	}
	return total
}

func (e *Engine) ScanAlbum(ctx context.Context, srcdir string) (*AlbumInfo, error) {
	files, err := GetSrcFiles(srcdir)
	if err != nil {
		return nil, err
	}

	cover, _ := GetCover(srcdir)

	info := &AlbumInfo{
		CoverPath: cover,
	}

	for _, f := range files {
		duration, _ := e.GetRuntime(ctx, f)
		info.Tracks = append(info.Tracks, Track{
			Title:    strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)),
			Duration: duration,
		})
		info.TotalDuration += duration
	}

	return info, nil
}

func (e *Engine) ProcessAlbum(ctx context.Context, srcPath string, fast bool, tempDir string) error {
	sourceFiles, err := GetSrcFiles(srcPath)
	if err != nil {
		return err
	}
	cover, err := GetCover(srcPath)
	if err != nil {
		return err
	}

	var ppfiles []string
	if fast {
		ppfiles = make([]string, len(sourceFiles))
		copy(ppfiles, sourceFiles)
	} else {
		ppfiles, err = e.PreprocessFiles(ctx, sourceFiles, tempDir)
		if err != nil {
			return err
		}
	}

	filelist, err := GenFilelist(ppfiles, tempDir)
	if err != nil {
		return err
	}
	totalDuration := e.GetTotalDuration(ctx, ppfiles)
	if err := e.GenTracklist(ctx, sourceFiles, filepath.Join(srcPath, "tracklist.txt")); err != nil {
		return err
	}
	err = e.MainFfmpegCall(ctx, filelist, cover, totalDuration, filepath.Join(srcPath, "out.mp4"))
	if err == nil {
		e.setProgress(1.0)
	}
	return err
}
