package archive

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/mholt/archives"
	"github.com/saintfish/chardet"
	"github.com/wutipong/immich-importer/immich"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

func Process(url *url.URL, apiKey string, sourceDir string, force bool) error {
	var albums []immich.AlbumResponseDto
	var err error

	if !force {
		albums, err = immich.GetAlbums(url, apiKey)
		if err != nil {
			err = fmt.Errorf("unable to retrieved existing albums. => %w", err)
			return err
		}
		slog.Debug("albums", slog.Any("existing albums", albums))
	}
	err = filepath.WalkDir(sourceDir,
		func(path string, d os.DirEntry, err error,
		) error {
			if d.IsDir() {
				return nil
			}

			if err != nil {
				slog.Warn(
					"failed to access path. skipping.",
					slog.String("path", path),
					slog.String("error", err.Error()),
				)
				return nil
			}

			if !slices.Contains(archiveExtensions, filepath.Ext(path)) {
				slog.Info("skipping non-archive file", slog.String("path", path))
				return nil
			}

			archiveFilePath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				slog.Error(
					"failed to determine album name.",
					slog.String("error", err.Error()),
				)
				return nil
			}

			slog.Debug("album name", slog.String("album_name", archiveFilePath))

			if slices.ContainsFunc(albums, func(a immich.AlbumResponseDto) bool {
				return a.AlbumName == archiveFilePath
			}) {
				slog.Warn("album already exists. skipping.", slog.String("album_name", archiveFilePath))
				return nil
			}

			ctx := context.Background()

			archiveFile, err := os.Open(path)
			if err != nil {
				slog.Warn("failed to open archive", slog.String("file", archiveFilePath))
				return nil
			}
			defer archiveFile.Close()

			assetIds := make([]string, 0)

			err = WalkArchive(ctx, path, archiveFile,
				func(
					ctx context.Context, filename string, f archives.FileInfo,
				) error {
					slog.Info(
						"creating asset",
						slog.String("archive", archiveFilePath),
						slog.String("entry", filename),
					)

					file, err := f.Open()
					if err != nil {
						slog.Error("failed to open archive entry",
							slog.String("archive", archiveFilePath),
							slog.String("entry", filename),
							slog.String("error", err.Error()),
						)
						return nil
					}

					defer file.Close()

					asset, err := immich.PostAsset(
						archiveFilePath, filename, file, f.ModTime(), url, apiKey,
					)
					if err != nil {
						slog.Error("failed to oupload asset",
							slog.String("archive", archiveFilePath),
							slog.String("entry", filename),
							slog.String("error", err.Error()),
						)
						return nil
					}

					slog.Info("uploaded asset", slog.Any("asset", asset))

					assetIds = append(assetIds, asset.ID)

					return nil
				})

			if err != nil {
				slog.Error(
					"failed to upload assets",
					slog.String("archive", archiveFilePath),
					slog.String("error", err.Error()),
				)
				return nil
			}

			slog.Info("creating album", slog.String("name", archiveFilePath))

			createdAlbum, err := immich.CreateAlbum(
				url, apiKey, archiveFilePath, assetIds,
			)
			if err != nil {
				slog.Error("failed to create album", slog.String("error", err.Error()))
				return nil
			}

			slog.Info("created album", slog.Any("album", createdAlbum))

			return nil
		})

	if err != nil {
		slog.Error("fails creating albums", slog.Any("error", err))
	}
	return nil
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

	err = extractor.Extract(ctx, stream, func(ctx context.Context, f archives.FileInfo) error {
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

		extension := filepath.Ext(f.NameInArchive)
		if slices.Contains(archiveExtensions, extension) {
			slog.Warn(
				"archive contains nested archived. manually extraction required.",
				slog.String("filename", filename),
			)
		}

		if immich.IsMediaFile(extension) {
			return mediaProcessFn(ctx, filename, f)
		}

		return nil
	})

	return err
}
