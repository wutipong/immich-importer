package archive

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mholt/archives"
	"github.com/saintfish/chardet"
	"github.com/wutipong/immich-importer/immich"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

func Process(
	server immich.ServerConfig,
	sourceDir string,
	albumPath string,
) (
	assetIds []string,
	err error,
) {
	if !IsArchiveFile(filepath.Ext(albumPath)) {
		return
	}
	archiveFile, err := os.Open(filepath.Join(sourceDir, albumPath))
	if err != nil {
		err = fmt.Errorf("failed to open archive: %s: %w.",
			albumPath,
			err,
		)
		return
	}
	defer archiveFile.Close()

	ctx := context.Background()
	err = WalkArchive(ctx, albumPath, archiveFile,
		func(
			ctx context.Context, filename string, f archives.FileInfo,
		) error {
			slog.Info(
				"creating asset",
				slog.String("archive", albumPath),
				slog.String("entry", filename),
			)

			file, err := f.Open()
			if err != nil {
				slog.Error("failed to open archive entry",
					slog.String("archive", albumPath),
					slog.String("entry", filename),
					slog.String("error", err.Error()),
				)
				return nil
			}

			defer file.Close()

			asset, err := immich.PostAsset(
				server,
				albumPath, filename, file, f.ModTime())
			if err != nil {
				slog.Error("failed to oupload asset",
					slog.String("archive", albumPath),
					slog.String("entry", filename),
					slog.String("error", err.Error()),
				)
				return nil
			}

			slog.Info("uploaded asset", slog.Any("asset", asset))

			assetIds = append(assetIds, asset.ID)

			return nil
		})
	return
}

func IsArchiveFile(path string) bool {
	return slices.Contains(archiveExtensions, strings.ToLower(filepath.Ext(path)))
}

var archiveExtensions = []string{
	".zip",
	".7z",
	".rar",
}

func WalkArchive(
	ctx context.Context,
	archivePath string,
	archive *os.File,
	mediaProcessFn func(
		ctx context.Context,
		filename string,
		f archives.FileInfo,
	) error,
) error {
	format, stream, err := archives.Identify(ctx, archivePath, archive)
	if err != nil {
		return err
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("format does not support extraction")
	}

	detector := chardet.NewTextDetector()
	var decoder *encoding.Decoder = nil
	var chardetResult *chardet.Result = nil

	err = extractor.Extract(
		ctx, stream,
		func(ctx context.Context, f archives.FileInfo) error {
			if f.IsDir() {
				return nil
			}

			filename := f.NameInArchive
			if decoder == nil {
				chardetResult, err = detector.DetectBest([]byte(filename))
				if err != nil {
					slog.Warn(
						"failed to detect encoding. using filename as is.",
						slog.String("filename", filename),
						slog.String("error", err.Error()),
					)
				} else {
					encoding, err := ianaindex.IANA.Encoding(chardetResult.Charset)
					if err != nil {
						slog.Warn(
							"failed to get encoding. using filename as is.",
							slog.String("filename", filename),
							slog.String("charset", chardetResult.Charset),
							slog.String("error", err.Error()),
						)
					} else {
						slog.Info(
							"detected filename encoding",
							slog.String("filename", filename),
							slog.String("charset", chardetResult.Charset),
						)
						decoder = encoding.NewDecoder()
					}
				}
			}

			if decoder != nil {
				filename, err = decoder.String(filename)
				if err != nil {
					slog.Warn(
						"failed to decode filename. using filename as is.",
						slog.String("filename", filename),
						slog.String("charset", chardetResult.Charset),
						slog.String("error", err.Error()),
					)
				}
			}

			if IsArchiveFile(f.NameInArchive) {
				slog.Warn(
					"archive contains nested archived. manually extraction required.",
					slog.String("filename", filename),
					slog.String("archive", archivePath),
				)
			}

			if immich.IsMediaFile(f.NameInArchive) {
				return mediaProcessFn(ctx, filename, f)
			}

			return nil
		})

	return err
}
