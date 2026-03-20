package immich

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func PostAsset(
	server ServerConfig,
	archivePath string,
	path string,
	reader io.Reader,
	modDate time.Time,
) (result AssetMediaResponseDto, err error) {
	if server.DryRun {
		slog.Debug(
			"Dry run: skipping asset upload",
			slog.String("path", path),
		)
		result = AssetMediaResponseDto{
			ID:     uuid.NewString(),
			Status: "created",
		}

		return
	}

	assetFileName := filepath.Join(archivePath, path)

	h := fnv.New64()
	h.Write([]byte(assetFileName))
	deviceAssetId := h.Sum64()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("deviceAssetId", strconv.FormatUint(deviceAssetId, 10))
	_ = writer.WriteField("deviceId", "WEB")
	_ = writer.WriteField("fileCreatedAt", modDate.Format(time.RFC3339))
	_ = writer.WriteField("fileModifiedAt", modDate.Format(time.RFC3339))
	_ = writer.WriteField("filename", path)

	part, err := writer.CreateFormFile("assetData", path)
	if err != nil {
		err = fmt.Errorf("failed to create form file: %w", err)
		return
	}

	_, err = io.Copy(part, reader)
	if err != nil {
		err = fmt.Errorf("failed to write data to form file: %w", err)
		return
	}
	_ = writer.Close()

	return DoRequestWithReturnObject[AssetMediaResponseDto](
		"POST", server, "/api/assets", &body, writer.FormDataContentType(),
	)
}

func IsMediaFile(path string) bool {
	return slices.Contains(mediaExtensions, filepath.Ext(path))
}

var mediaExtensions = []string{
	// Image formats
	".avif",
	".bmp",
	".gif",
	".heic",
	".heif",
	".jp2",
	".jpeg",
	".jpg",
	".jpe",
	".insp",
	".jxl",
	".png",
	".psd",
	".raw",
	".rw2",
	".svg",
	".tif",
	".tiff",
	".webp",
	// video format
	".3gp",
	".3gpp",
	".avi",
	".flv",
	".m4v",
	".mkv",
	".mts",
	".m2ts",
	".m2t",
	".mp4",
	".insv",
	".mpg",
	".mpe",
	".mpeg",
	".mov",
	".webm",
	".wmv",
}
