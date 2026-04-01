# immich-importer

A command-line importer for Immich that uploads media from local directories or archive files into Immich albums.

## Features

- Recursively scan a source directory
- Upload media from directories as Immich albums
- Upload media from archive files
- Create album names from relative folder or archive paths
- Support dry-run mode for previewing behavior
- JSON file logging plus console logging
- Multiple profiles via YAML config

## Requirements

- Go 1.26+

## Installation

Install from the repository path:

```bash
go install github.com/wutipong/immich-importer@latest
```

Or build locally from source:

```bash
go build -o immich-importer .
```

## Configuration

The application stores configuration in `~/.immich-importer/config.yaml`.

Example config:

```yaml
default:
  immich_url: "https://your-immich-instance"
  immich_api_key: "YOUR_API_KEY"
```

### Multiple profiles

You can define multiple profiles in the same YAML file:

```yaml
default:
  immich_url: "https://your-immich-instance"
  immich_api_key: "YOUR_API_KEY"

work:
  immich_url: "https://work-immich-instance"
  immich_api_key: "WORK_API_KEY"
```

Use `--profile=<name>` to select a profile for any command.

## Commands

### `setup`

Interactively create or update the profile configuration.

```bash
immich-importer setup
immich-importer setup --profile=myprofile
```

### `run`

Import both directories and archive files from a source directory.

```bash
immich-importer run --source-dir /path/to/import
```

Flags:
- `--source-dir`, `--src`, `--source` (required): source directory to scan
- `--force`: force processing even if an album with the same name already exists
- `--dry-run`: run without uploading to Immich
- `--disable-directory`: disable processing media inside directories
- `--disable-archive`: disable processing archive files

### `archive`

Create an album from a single archive file.

```bash
immich-importer archive --source-dir /path/to --archive photos.zip
```

Flags:
- `--source-dir`, `--src`, `--source` (required): directory containing the archive
- `--archive` (required): archive path relative to `source-dir`
- `--dry-run`: run without uploading to Immich

### `directory`

Create an album from a specific directory inside a source directory.

```bash
immich-importer directory --source-dir /path/to --directory vacation
```

Flags:
- `--source-dir`, `--src`, `--source` (required): root source directory
- `--directory` (required): directory path relative to `source-dir`
- `--dry-run`: run without uploading to Immich

### `merge`

Merge multiple existing Immich albums into a new album.

```bash
immich-importer merge --album merged-album --pattern '^2024.*$'
```

Flags:
- `--album` (required): name of the album to create
- `--pattern` (required): Go RE2 regex to match source album names
- `--disable-deletion`: keep source albums instead of deleting them after merge
- `--dry-run`: run without modifying Immich

### `log`

Manage log file output.

```bash
immich-importer log location
immich-importer log latest
immich-importer log purge --keep-latest 2
```

## Global flags

These flags work for all commands:

- `--display-log` (default `warn`): minimum console log level (`debug`, `info`, `warn`, `error`)
- `--file-log` (default `info`): minimum log file level
- `--profile` (default `default`): profile name from `config.yaml`

## Examples

Setup configuration with debug logging:

```bash
immich-importer setup --display-log debug
```

Import a directory tree as albums:

```bash
immich-importer run --source-dir ~/Pictures --dry-run --display-log debug
```

Import using a different profile:

```bash
immich-importer run --source-dir ~/Pictures --profile=work
```

Create an album from an archive:

```bash
immich-importer archive --source-dir ~/Downloads --archive vacation.zip
```

## Log storage

Log files are stored under `~/.immich-importer/logs`.
Use `immich-importer log latest` to print the latest log file path.

## Notes

- `setup` writes `~/.immich-importer/config.yaml`
- `run` skips existing albums unless `--force` is used
- `merge` can delete matching albums by default; use `--disable-deletion` to keep them
- `dry-run` is available on commands that interact with Immich

## Troubleshooting

- `unable to load configuration`: verify `~/.immich-importer/config.yaml` exists and contains the requested profile
- `invalid immich url`: ensure `immich_url` is a valid `http://` or `https://` URL
- `failed upload assets`: check Immich server availability and API key permissions

## Development

Build the project locally:

```bash
go build
```
