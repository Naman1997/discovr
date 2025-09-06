# Discovr

[![Go Build & Test](https://github.com/Naman1997/discovr/actions/workflows/main.yml/badge.svg)](https://github.com/Naman1997/discovr/actions/workflows/main.yml)
[![Go Fmt & Commit](https://github.com/Naman1997/discovr/actions/workflows/gofmt.yml/badge.svg)](https://github.com/Naman1997/discovr/actions/workflows/gofmt.yml)
[![OSV-Scanner PR Scan](https://github.com/Naman1997/discovr/actions/workflows/osv-scanner-pr.yml/badge.svg)](https://github.com/Naman1997/discovr/actions/workflows/osv-scanner-pr.yml)

Automated asset discovery

# Development

### Linux

```
make
./discovr -h
```

##### Running passive scans in Linux

```
# Run a passive scan all interfaces
./discovr local passive

# Run a passive scan on the specified interface for a specified amount of time
./discovr local passive -i eth0 -d 20
```

##### Running active scans in Linux

```
# Run a scan on localhost for top 1000 ports
./discovr local active

# Run a scan on a target ip with specified ports
./discovr local active -t 10.10.10.10 -p 80,443
```

### Windows

```
# Install Docker Desktop for Windows and WSL2

# Inside WSL2, navigate to the directory containing the repository
# Example
cd /mnt/c/Users/%USERNAME%/Documents/

# Create a new directory
mkdir -p Github && cd Github

# Update system
sudo apt update -y

# Install dependencies
sudo apt install git gh make -y

# Login to github
gh auth login

# Clone the repo
git clone https://github.com/Naman1997/discovr.git

# Go into the directory
cd discovr

# Get the nmap windows zip
make get_nmap_win_zip

# Open Windows Command Prompt(cmd) and navigate to the same directory
cd C:\Users\%USERNAME%\Documents\Github\discovr\

# Set the environment variables
for /F %A in (.env) do SET %A

# Build (from the same command prompt)
go build -ldflags="-X 'github.com/Naman1997/discovr/internal.NmapVersion=%NMAP_VERSION%'"

# Run the binary
.\discovr.exe -h
```

##### Running passive scans in Windows

```
# Get your device id
getmac /fo csv /v

# Example: "\Device\Tcpip_{ED16A895-687F-4D8C-B13B-930295C92D21}"

# Replace "Tcpip" with "NPF"
# Example: "\Device\NPF_{ED16A895-687F-4D8C-B13B-930295C92D21}"

# Run a passive scan on an interface (the default option does not work atm)
.\discovr.exe local passive -i \Device\NPF_{ED16A895-687F-4D8C-B13B-930295C92D21}
```

##### Running active scans in Windows

```
# Run a scan on localhost for top 1000 ports
.\discovr.exe local active

# Run a scan on a target ip with specified ports
.\discovr.exe local active -t 10.10.10.10 -p 80,443
```
