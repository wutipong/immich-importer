**note:** *README.md included AI-generated content.*

# immich-importer

A small CLI tool to import media assets from local directories or archive files into an Immich server.

## Features

- Scan a source directory recursively
- Upload assets from subdirectories as Immich albums
- Upload assets from archives (zip, tar, etc.)
- Create albums automatically from relative folder paths
- Supports dry-run for previewing behavior
- Logging to console and JSON file

## Installation

Make sure you have Go installed (1.26+ recommended).

```bash
go install github.com/wutipong/immich-importer@latest
```

Or build locally:

```bash
go build -o immich-importer .
```

## Configuration

Create a YAML config file at `~/.immich-importer/config.yaml`.

Example config:

```yaml
default:
    immich_url: "https://your-immich-instance"
    immich_api_key: "YOUR_API_KEY"
```

The API key requires appropriete permission. At least it should have *Album*, *Assets* and *Album Assets* permission.

## Profile

The user can add multiple configurations by using different profile. For example:

```yaml
default:
    immich_url: "https://your-immich-instance"
    immich_api_key: "YOUR_API_KEY"

another:
    immich_url: "https://your-another-immich-instance"
    immich_api_key: "YOUR-ANOTHER_API_KEY"
```

Then use `--profile=another` to import to another server.

## Usage

```bash
immich-importer --source /path/to/import
```

### Flags

- `--source, --src` (required): source directory to import from
- `--profile` (default `default`): profile name
- `--force`: force processing even if albums already exist
- `--dry-run`: do not upload to Immich, just simulate processing
- `--disable-directory`: skip scanning subdirectories
- `--disable-archive`: skip processing archive files
- `--display-log` (default `warn`): console log level (`debug`, `info`, `warn`, `error`)
- `--file-log` (default `info`): file log level

### Example

```bash
immich-importer --source ~/Pictures --dry-run --display-log debug
```

## How It Works

1. Reads `~/.immich-importer/config.yaml` for Immich server URL and API key.
2. Walks through the source directory recursively.
3. Uploads assets in directories (if enabled) using `directory.Process`.
4. Uploads assets in archive files (if enabled) using `archive.Process`.
5. Creates Immich albums with the relative path as album name.

## Logs

A log file is created beside the current working directory named `immich-importer.<timestamp>.log`.

## Development

Run build with:

```bash
go build
```

## Notes

- Be cautious running without `--force`; the importer skips albums with existing names by default.
- Ensure your Immich API key has permissions to create albums and upload assets.

## Troubleshooting

### Common issues

- `unable to load configuration`: verify `~/.immich-importer/config.yaml` exists and has valid YAML.
- `invalid immich url`: ensure `immich_url` starts with `http://` or `https://` and is reachable.
- `failed upload assets`: check API key permissions and Immich server availability.

### Validate your config quickly

```bash
cat ~/.immich-importer/config.yaml
```

Confirm it includes:

```yaml
default:
    immich_url: "https://your-immich-instance"
    immich_api_key: "YOUR_API_KEY"
```

If your Immich instance uses self-signed TLS, ensure your system trusts the certificate or use an appropriate CA bundle.
