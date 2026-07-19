# Clip Installation Guide

## Recommended Installation Method

To ensure the Go Clip binary is always executed and doesn't conflict with Python's clip command:

### 1. Build the Go binary

```bash
go build -o clip ./cmd/clip
```

### 2. Install globally (PATH-safe setup)

```bash
sudo mv clip /usr/local/bin/clip
```

### 3. Verify installation

```bash
which clip
```

This should output: `/usr/local/bin/clip`

```bash
clip --version
```

This should output: `Clip version dev`

## Alternative Installation Methods

### Using go install (recommended for developers)

```bash
go install github.com/upendra7470/clip/cmd/clip@latest
```

This will install the binary to your `$GOPATH/bin` directory. Make sure this directory is in your PATH.

### Manual installation

1. Clone the repository:
   ```bash
   git clone https://github.com/upendra7470/clip.git
   cd clip
   ```

2. Build the binary:
   ```bash
   go build -o clip ./cmd/clip
   ```

3. Add to PATH:
   ```bash
   sudo cp clip /usr/local/bin/
   ```

## Troubleshooting PATH Conflicts

If `which clip` still shows a Python package instead of the Go binary:

1. Check your PATH order:
   ```bash
   echo $PATH
   ```

2. Ensure `/usr/local/bin` comes before Python's installation directories.

3. Use the full path to the Go binary:
   ```bash
   /usr/local/bin/clip --version
   ```

## Uninstallation

To remove the Clip binary:

```bash
sudo rm /usr/local/bin/clip
```

## Usage Examples

### Basic usage
```bash
clip document.pdf
```

### With exact path
```bash
clip ./files/report.pdf
```

### With absolute path
```bash
clip /full/path/to/document.pdf
```

### Help
```bash
clip --help
```

### Version
```bash
clip --version
```

## Smart File Resolution

Clip will automatically search for files in these locations:
1. Current directory
2. ~/Downloads
3. ~/Desktop
4. ~/Documents

If multiple files with the same name are found, Clip will prompt you to select one.