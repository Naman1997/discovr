# Discovr

[![Go Build & Test](https://github.com/Naman1997/discovr/actions/workflows/main.yml/badge.svg)](https://github.com/Naman1997/discovr/actions/workflows/main.yml)  [![Go Fmt & Commit](https://github.com/Naman1997/discovr/actions/workflows/gofmt.yml/badge.svg)](https://github.com/Naman1997/discovr/actions/workflows/gofmt.yml)

Automated asset discovery

# Development

### Linux

```
make
./discovr -h
```


### Windows

```
# Enable sudo from the Developer Settings page[ms-settings:developers]

# Install Docker Desktop for Windows and WSL2

# Inside WSL2, navigate to the directory containing the repository
# Example
cd /mnt/c/Users/loki/Documents/Github/discovr/

# Get the nmap binary and windows zip
make get_nmap_binary
make get_nmap_win_zip

# Open Windows Command Prompt(cmd) and navigate to the same directory
# Example
cd C:\Users\loki\Documents\Github\discovr\

# Set the environment variables
for /F %A in (.env) do SET %A

# Build (from the same command prompt)
go build -ldflags="-X 'github.com/Naman1997/discovr/internal.NmapVersion=%NMAP_VERSION%'"

# Run the binary
.\discovr.exe -h
```