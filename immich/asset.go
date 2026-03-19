package immich

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

func PostAsset(
	archivePath string,
	path string,
	reader io.Reader,
	modDate time.Time,
	url *url.URL,
	apiKey string,
) (result AssetMediaResponseDto, err error) {
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

	return DoRequestWithResult[AssetMediaResponseDto](
		"POST", url.JoinPath("/api/assets"), &body, writer.FormDataContentType(), apiKey,
	)
}

func IsMediaFile(extension string) bool {
	return slices.Contains(mediaExtensions, extension)
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
