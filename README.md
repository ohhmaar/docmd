# docmd

A CLI tool that syncs local markdown files to Google Docs.

Write in your favorite editor, sync to Google Docs for sharing and collaboration.

## Features

- **Link** markdown files to Google Docs (creates new doc)
- **Push** local changes to Google Docs
- **Watch** mode for automatic syncing on file changes
- **Conflict detection** when the Google Doc has been modified (no auto-resolution)
- Markdown → HTML conversion with support for:
  - Headings, bold, italic, strikethrough
  - Lists (ordered and unordered)
  - Links
  - Code blocks
  - Tables (GitHub Flavored Markdown)
  - Blockquotes

## Installation

```bash
# Clone the repository
git clone https://github.com/ohhmaar/docmd.git
cd docmd

# Install dependencies and build
go mod tidy
go build -o docmd .

# Optionally, install to your PATH
go install
```

## Setup

### 1. Create Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or select an existing one)
3. Enable the **Google Drive API**
4. Go to "APIs & Services" → "Credentials"
5. Click "Create Credentials" → "OAuth client ID"
6. Select "Desktop app" as the application type
7. Download the credentials

### 2. Authenticate

Use the downloaded credentials file to authenticate:

```bash
docmd init ~/Downloads/client_secret_*.json
```

This will open your browser for Google authentication. Your credentials are stored in `~/.docmd/`.

## Usage

### Link a markdown file to a new Google Doc

```bash
docmd link README.md

# With custom title
docmd link README.md --title "Project Documentation"

# In a specific Google Drive folder
docmd link README.md --folder <folder-id>
```

### Push changes to Google Docs

```bash
# Push a specific file
docmd push README.md

# Push all linked files
docmd push --all

# Skip conflict check (force overwrite)
docmd push README.md --force
```

### Watch for changes (auto-sync)

```bash
# Watch a specific file
docmd watch README.md

# Watch all linked files
docmd watch --all

# Custom debounce delay (default: 500ms)
docmd watch README.md --debounce 1000
```

### Check sync status

```bash
docmd status
```

### Unlink a file

```bash
# Unlink (keeps Google Doc)
docmd unlink README.md

# Unlink and delete the Google Doc
docmd unlink README.md --delete
```

## Configuration

Configuration is stored in `~/.docmd/`:

- `config.json` - Linked files and sync metadata
- `token.json` - OAuth credentials (do not share!)

## How It Works

1. **Markdown → HTML**: Your markdown is converted to HTML using [goldmark](https://github.com/yuin/goldmark)
2. **HTML → Google Doc**: The HTML is uploaded via Google Drive API, which automatically converts it to native Google Doc format
3. **Sync tracking**: docmd tracks when each file was last synced and compares with the Google Doc's modification time to detect conflicts (no auto-resolution)

## Limitations

- **One-way sync**: Currently only supports local → Google Docs. Changes made in Google Docs won't sync back.
- **Bidirectional sync**: Planned, with a diff view for reviewing changes before applying them.
- **Full document replacement**: Each push replaces the entire document content (no incremental updates)
- **Images**: Image support is planned for a future release

## Troubleshooting

### "Not authenticated" error

Run `docmd init` to authenticate with Google.

### "Invalid credentials file" error

Make sure you're providing the correct path to the credentials JSON file downloaded from Google Cloud Console.

### "Google Doc not found" error

The linked Google Doc may have been deleted. Use `docmd unlink` to remove the stale link, then `docmd link` to create a new doc.

## License

MIT
