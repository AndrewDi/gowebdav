# WebDAV CLI Tool Usage Documentation

## Overview

`webdav-cli` is a command-line tool for performing basic WebDAV operations, such as listing directory contents, uploading, downloading, creating directories, and deleting files.

## Installation

```bash
# Build from source
cd /path/to/gowebdav
go build -o webdav-cli ./cmd

# Or use go install
go install github.com/AndrewDi/gowebdav/cmd
```

## Basic Usage

```bash
webdav-cli -endpoint <endpoint> -username <username> -password <password> -command <command> -remote <remote_path> [-local <local_path>]
```

## Proxy Configuration

### Using Environment Variables for Proxy

```bash
# Set HTTP proxy
export http_proxy=http://127.0.0.0:11111

# Set HTTPS proxy (if needed)
export https_proxy=http://127.0.0.0:11111

# Execute command
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command ls -remote /
```

## Command Details

### 1. ls - List directory contents

**Function**: List the contents of a specified WebDAV directory.

**Parameters**:
- `-remote`: Remote directory path

**Example**:
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command ls -remote /
```

### 2. get - Download file

**Function**: Download a file from the WebDAV server to local.

**Parameters**:
- `-remote`: Remote file path
- `-local`: Local save path

**Example**:
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command get -remote /document.txt -local ./document.txt
```

### 3. put - Upload file

**Function**: Upload a local file to the WebDAV server.

**Parameters**:
- `-local`: Local file path
- `-remote`: Remote save path

**Example**:
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command put -local ./document.txt -remote /document.txt
```

### 4. mkdir - Create directory

**Function**: Create a directory on the WebDAV server.

**Parameters**:
- `-remote`: Directory path to create

**Example**:
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command mkdir -remote /new-folder
```

### 5. rm - Delete file

**Function**: Delete a file from the WebDAV server.

**Parameters**:
- `-remote`: File path to delete

**Example**:
```bash
export http_proxy=http://127.0.0.0:11111
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command rm -remote /document.txt
```

### 6. console - Interactive Console Mode

**Function**: Start an interactive console for running multiple WebDAV commands.

**Parameters**: Supports all global options (--config, --endpoint, --username, --password, etc.)

**Example**:
```bash
# Start with config file
webdav-cli console --config ~/.webdav.yaml

# Start with command-line parameters
webdav-cli console --endpoint http://192.168.1.5:5005 --username webcli --password @Andrewtomy9
```

**Console Built-in Commands**:
- `help` - Show available commands
- `exit` / `quit` - Exit the console
- `clear` - Clear the screen
- `cd [path]` - Change directory (cd, cd .., cd /path)
- `pwd` - Print working directory
- `ll` - List directory contents with long format, sorted by time (alias for `ls -lrt`)

**Usage Example**:
```
$ webdav-cli console --config ~/.webdav.yaml
Welcome to WebDAV CLI Console!
Type 'help' for available commands, 'exit' or 'quit' to exit.
Use Tab for auto-completion.

webdav> ls
webdav> ll
webdav> cd /documents
webdav:/documents> ls
webdav:/documents> ll
webdav:/documents> pwd
/documents
webdav:/documents> exit
```

**Features**:
- **Tab Completion**: Press Tab to auto-complete commands and file paths
- **Directory Navigation**: Use cd and pwd commands to navigate directories
- **Command History**: Use up/down arrow keys to browse command history

### 7. cat - Display File Content (with Format Support)

**Function**: Display file content from WebDAV server, with automatic formatting for JSON and YAML files.

**Parameters**:
- `path`: File path

**Example**:
```bash
# View plain text file
webdav-cli cat /document.txt

# View JSON file (auto-formatted)
webdav-cli cat /config.json

# View YAML file (auto-formatted)
webdav-cli cat /config.yaml
```

**Format Support**:
- `.json` files are automatically formatted with indentation
- `.yaml` / `.yml` files are automatically formatted with proper indentation

## Complete Example

### Scenario: Upload file to server and verify

```bash
# Set proxy
export http_proxy=http://127.0.0.0:11111

# Create test file
echo "Hello WebDAV" > test.txt

# Upload file
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command put -local test.txt -remote /test.txt

# List directory contents to confirm file has been uploaded
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command ls -remote /

# Download file for verification
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command get -remote /test.txt -local downloaded.txt

# View downloaded file content
cat downloaded.txt

# Clean up files
webdav-cli -endpoint http://192.168.1.5:5005 -username webcli -password @Andrewtomy9 -command rm -remote /test.txt
rm test.txt downloaded.txt
```

## Common Issues

### 1. Proxy connection failure

**Symptom**: Connection timeout or proxy error when executing commands.

**Solution**:
- Confirm the proxy server is running
- Check if the proxy address and port are correct
- Verify if the proxy server allows access to the target WebDAV server

### 2. Authentication failure

**Symptom**: 401 Unauthorized error when executing commands.

**Solution**:
- Confirm the username and password are correct
- Verify if the user has permission to access the specified WebDAV resource

### 3. Path error

**Symptom**: 404 Not Found error when executing commands.

**Solution**:
- Confirm the remote path exists
- For upload operations, ensure the target directory exists
