package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/disintegration/imaging"
)

const frameRateRegexStr = `(\d+\.\d+|\d+) fps`

func handleErr(errText string, err error) {
	errString := errText + "\n" + err.Error()
	panic(errString)
}

func parseArgs() (*Args, error) {
	var args Args
	arg.MustParse(&args)
	if !(args.Mode == 1 || args.Mode == 2) {
		return nil, errors.New("Mode must me 1 or 2.")
	}
	if !strings.HasSuffix(args.OutPath, ".webm") {
		return nil, errors.New(`Output file extension must be ".webm".`)
	}
	return &args, nil
}

func makeDirs(path string) error {
	base := filepath.Base(path)
	if base != path {
		err := os.MkdirAll(path, 0755)
		return err
	}
	return nil
}

// Best way without ffprobe that supports all formats.
// Iterate over lines instead of regexing the whole output so we get the right value.
func extractFrameRate(out string) string {
	lines := strings.Split(out, "\n")
	for _, _line := range lines {
		line := strings.TrimSpace(_line)
		if strings.HasPrefix(line, "Stream") {
			regex := regexp.MustCompile(frameRateRegexStr)
			match := regex.FindStringSubmatch(line)
			if match != nil {
				return match[1]
			}
		}
	}
	return ""
}

func extractFrames(vidPath, tempPath string) (string, error) {
	var errBuffer bytes.Buffer
	outPath := filepath.Join(tempPath, "out%04d.png")
	args := []string{"-hide_banner", "-i", vidPath, outPath}
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = &errBuffer
	err := cmd.Run()
	if err != nil {
		errString := fmt.Sprintf("%s\n%s", err, errBuffer.String())
		return "", errors.New(errString)
	}
	frameRate := extractFrameRate(errBuffer.String())
	if frameRate == "" {
		return "", errors.New("No regex match for frame rate.")
	}
	return frameRate, nil
}

func getFrameBases(tempPath string) ([]string, error) {
	var paths []string
	// ReadDir doesn't guarantee order.
	files, err := ioutil.ReadDir(tempPath)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(files)+1; i++ {
		base := filepath.Join(tempPath, fmt.Sprintf("out%04d", i))
		paths = append(paths, base)
	}
	return paths, nil
}

// Between 50 and 1000.
func genRandom() int {
	return rand.Intn(1000) + 50
}

func resizeImages(frameBases []string, mode int) error {
	var (
		x   int
		y   int
		img *image.NRGBA
	)
	for i, base := range frameBases {
		imagePath := base + ".png"
		resImagePath := base + "_r.png"
		f, err := imaging.Open(imagePath)
		if err != nil {
			return err
		}
		if i == 0 {
			b := f.Bounds()
			x = b.Max.X
			y = b.Max.Y
			err := os.Rename(imagePath, resImagePath)
			if err != nil {
				return err
			}
			continue
		}
		switch mode {
		case 1:
			img = imaging.Resize(f, genRandom(), genRandom(), imaging.Lanczos)
		case 2:
			x += 20
			y += 20
			img = imaging.Resize(f, x, y, imaging.Lanczos)
		}
		err = imaging.Save(img, resImagePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func framesToWebms(frameBases []string, frameRate string) error {
	var errBuffer bytes.Buffer
	for _, base := range frameBases {
		args := []string{
			"-hide_banner", "-loglevel", "error", "-framerate", frameRate,
			"-f", "image2", "-i", base + "_r.png", "-c:v",
			"libvpx-vp9", "-pix_fmt", "yuva420p", base + ".webm",
		}
		cmd := exec.Command("ffmpeg", args...)
		cmd.Stderr = &errBuffer
		err := cmd.Run()
		if err != nil {
			errString := fmt.Sprintf("%s\n%s", err, errBuffer.String())
			return errors.New(errString)
		}
	}
	return nil
}

func concatWebms(concatPath, outPath string, frameBases []string) error {
	f, err := os.OpenFile(concatPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, base := range frameBases {
		line := fmt.Sprintf("file '%s.webm'\n", base)
		_, err := f.WriteString(line)
		if err != nil {
			return err
		}
	}
	var errBuffer bytes.Buffer
	args := []string{
		"-hide_banner", "-loglevel", "error", "-f", "concat", "-safe",
		"0", "-i", concatPath, "-c", "copy", "-y", outPath,
	}
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = &errBuffer
	err = cmd.Run()
	if err != nil {
		errString := fmt.Sprintf("%s\n%s", err, errBuffer.String())
		panic(errString)
	}
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	args, err := parseArgs()
	if err != nil {
		handleErr("Failed to parse args.", err)
	}
	err = makeDirs(args.OutPath)
	if err != nil {
		handleErr("Failed to make output path.", err)
	}
	tempPath, err := os.MkdirTemp(os.TempDir(), "")
	concatPath := filepath.Join(tempPath, "concat.txt")
	if err != nil {
		handleErr("Failed to make temp directory.", err)
	}
	defer os.RemoveAll(tempPath)
	fmt.Println("Extracting frames...")
	frameRate, err := extractFrames(args.InPath, tempPath)
	if err != nil {
		panic(err)
	}
	frameBases, err := getFrameBases(tempPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("Resizing frames...")
	err = resizeImages(frameBases, args.Mode)
	if err != nil {
		panic(err)
	}
	fmt.Println("Frames -> WebMs...")
	err = framesToWebms(frameBases, frameRate)
	if err != nil {
		panic(err)
	}
	fmt.Println("Concatting WebMs...")
	err = concatWebms(concatPath, args.OutPath, frameBases)
	if err != nil {
		panic(err)
	}
}
