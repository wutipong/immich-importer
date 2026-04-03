package archive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mholt/archives"
	"github.com/wutipong/immich-importer/immich"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

func Process(
	ctx context.Context,
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

	assetIds, err = WalkArchive(ctx, server, albumPath, archiveFile)

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
	server immich.ServerConfig,
	archivePath string,
	archive io.Reader,
) (assetIds []string, err error) {
	if ctx.Err() != nil {
		err = fmt.Errorf("context error: %w", ctx.Err())
		return
	}

	format, stream, err := archives.Identify(ctx, archivePath, archive)
	if err != nil {
		return
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		err = fmt.Errorf("format does not support extraction")
		return
	}

	var decoder *encoding.Decoder = nil

	err = extractor.Extract(
		ctx, stream,
		func(ctx context.Context, f archives.FileInfo) error {
			if ctx.Err() != nil {
				err = fmt.Errorf("context error: %w", ctx.Err())
				return err
			}

			if f.IsDir() {
				return nil
			}

			filename := f.NameInArchive
			if decoder == nil {
				en, charset, err := DetectCharSet(filename)
				if err != nil {
					slog.Warn("failed to detect filename character set",
						slog.String("filename", filename),
						slog.String("archive file", archivePath))

					decoder = &encoding.Decoder{
						Transformer: transform.Nop,
					}
				} else {
					slog.Debug(
						"using character set",
						slog.String("charset", charset),
						slog.String("filename", filename),
						slog.String("archive file", archivePath),
					)
					decoder = en.NewDecoder()
				}
			}

			if decoder != nil {
				filename, err = decoder.String(filename)
				if err != nil {
					slog.Warn(
						"failed to decode filename. using filename as is.",
						slog.String("filename", filename),
						slog.String("error", err.Error()),
					)
				}
			}

			if IsArchiveFile(f.NameInArchive) {
				slog.Info(
					"nested archive found.",
					slog.String("filename", filename),
					slog.String("archive", archivePath),
				)

				file, err := f.Open()

				if err != nil {
					return fmt.Errorf("failed to open nested archive %s: %w", filename, err)
				}
				defer file.Close()

				buffer := bytes.Buffer{}
				_, err = io.Copy(&buffer, file)

				reader := bytes.NewReader(buffer.Bytes())

				nestedAssetIds, err := WalkArchive(
					ctx, server, filepath.Join(archivePath, filename), reader,
				)
				if err != nil {
					return fmt.Errorf("failed to process nested archive %s: %w", filename, err)
				}
				assetIds = append(assetIds, nestedAssetIds...)
				return nil
			}

			if immich.IsMediaFile(f.NameInArchive) {
				asset, err := uploadAsset(ctx, server, archivePath, filename, f)
				if err != nil {
					return err
				}

				slog.Info("uploaded asset", slog.Any("asset", asset))

				assetIds = append(assetIds, asset.ID)

			}

			return nil
		})

	return assetIds, err
}

func uploadAsset(
	ctx context.Context,
	server immich.ServerConfig,
	archivePath string,
	filename string,
	f archives.FileInfo,
) (asset immich.AssetMediaResponseDto, err error) {
	if ctx.Err() != nil {
		err = fmt.Errorf("context error: %w", ctx.Err())
		return
	}

	slog.Info(
		"creating asset",
		slog.String("archive", archivePath),
		slog.String("entry", filename),
	)

	file, err := f.Open()
	if err != nil {
		err = fmt.Errorf("failed to open archive entry %s/%s: %w", archivePath, filename, err)
		return
	}

	defer file.Close()

	asset, err = immich.PostAsset(
		ctx,
		server,
		archivePath,
		filename,
		file,
		f.ModTime(),
	)
	if err != nil {
		err = fmt.Errorf("failed to upload asset %s/%s: %w", archivePath, filename, err)
		return
	}
	return
}
