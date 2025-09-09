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
# Clone the repo with github desktop
git clone https://github.com/Naman1997/discovr.git

# Go into the directory
cd discovr

# Build the project
./build.bat

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
