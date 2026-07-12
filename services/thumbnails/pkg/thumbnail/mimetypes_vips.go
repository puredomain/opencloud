//go:build enable_vips

package thumbnail

var (
	// SupportedMimeTypes contains an all mimetypes which are supported by the thumbnailer.
	SupportedMimeTypes = map[string]struct{}{
		"image/png":      {},
		"image/jpg":      {},
		"image/jpeg":     {},
		"image/gif":      {},
		"image/bmp":      {},
		"image/x-ms-bmp": {},
		"image/tiff":     {},
		"text/plain":     {},
		"audio/flac":     {},
		"audio/mpeg":     {},
		"audio/ogg":      {},
		// timocloud patch #2: video poster frames via ffmpeg (preprocessor_video.go).
		"video/mp4":                         {},
		"video/webm":                        {},
		"video/quicktime":                   {},
		"video/x-matroska":                  {},
		"video/x-msvideo":                   {},
		"application/vnd.geogebra.slides":   {},
		"application/vnd.geogebra.pinboard": {},
		"image/webp":                        {},
	}
)
