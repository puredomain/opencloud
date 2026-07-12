package preprocessor

// timocloud patch #2 (PATCHES.md): video thumbnails via an ffmpeg exec —
// upstream issue opencloud-eu/opencloud#3119 (video thumbnail requests 404
// because no video mimetype is in SupportedMimeTypes and no decoder
// exists). A pure-Go frame decode isn't realistic across containers/codecs;
// ffmpeg is added to the runtime image (Dockerfile). The input is spooled
// to a temp file first — most containers (mp4 moov atoms at the end, mkv
// seeking) need seekable input, so piping stdin is not reliable.

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
)

// VideoDecoder extracts a poster frame from a video stream using ffmpeg.
type VideoDecoder struct{}

// videoFrameAt runs ffmpeg against the spooled input, asking for one frame
// at the given seek offset, PNG-encoded on stdout.
func videoFrameAt(path string, seek string) ([]byte, error) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	// -ss before -i: fast input seek; -frames:v 1: single frame;
	// image2pipe+png: stdout, no temp output file.
	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-ss", seek,
		"-i", path,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-",
	)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg frame extraction failed: %w (%s)", err, errOut.String())
	}
	return out.Bytes(), nil
}

// Convert spools the video to a temp file and extracts a frame ~1s in
// (matching the web client's own poster-capture heuristic); falls back to
// the very first frame for clips shorter than that.
func (v VideoDecoder) Convert(r io.Reader) (any, error) {
	tmp, err := os.CreateTemp("", "thumb-video-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}

	frame, err := videoFrameAt(tmp.Name(), "1")
	if err != nil || len(frame) == 0 {
		// Clips shorter than the 1s seek produce no output — take frame 0.
		frame, err = videoFrameAt(tmp.Name(), "0")
		if err != nil {
			return nil, err
		}
	}
	if len(frame) == 0 {
		return nil, fmt.Errorf("ffmpeg produced no frame")
	}

	img, _, err := image.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	return img, nil
}
