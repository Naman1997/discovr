````markdown
# Discovr — User Documentation
---

## Overview
Discovr is a portable asset discovery CLI tool for scanning on-premise and cloud environments.  
It supports active, passive, and Nmap network discovery as well as cloud inventory for AWS, Azure, and GCP.  
Use the TUI for a guided interactive experience.

Run:
```bash
discovr -h
````

to list available commands.

---

## 1. Getting Started

### 1.1 System requirements

* Go (for building) if you need to compile from source (optional if using prebuilt binaries)
* Linux, macOS, or Windows
* Network interface access (root/administrator for some operations)
* AWS / Azure / GCP credentials for cloud scans (if using cloud commands)

### 1.2 Installation (quick)

**Linux / macOS**

```bash
git clone https://github.com/Naman1997/discovr.git
cd discovr
make
```

You can also build manually with:

```bash
go build -o discovr ./cmd
```

Verify installation:

```bash
./discovr -h
```

**Windows**

```bash
git clone https://github.com/Naman1997/discovr.git
cd discovr
build.bat
```

Then run:

```bash
discovr.exe -h
```

---

## 2. Command Reference

> Each subcommand below maps to a file in the `cmd/` directory.

### `active` — Active network scan

**Synopsis**

```bash
discovr active [flags]
```

**Description**
Sends probes across a CIDR range to discover live hosts (IP addresses).
By default, it uses ARP on a specified interface or ICMP pings.
Note: This command does **not** find MAC addresses.

**Flags**

|            Flag | Short | Type   | Default | Description                                            |
| --------------: | ----: | ------ | ------: | ------------------------------------------------------ |
|   `--interface` |  `-i` | string |       — | **Required** network interface for ARP (e.g., `eth0`). |
|        `--cidr` |  `-r` | string |       — | Target CIDR to scan (e.g., `192.168.1.0/24`).          |
|        `--mode` |  `-m` | bool   | `false` | Use ICMP echo requests instead of ARP.                 |
| `--concurrency` |  `-p` | int    |    `50` | Number of concurrent workers (ICMP).                   |
|     `--timeout` |  `-t` | int    |     `2` | Timeout (sec) for replies (ICMP).                      |
|       `--count` |  `-c` | int    |     `1` | Number of requests per IP (ICMP).                      |
|      `--export` |  `-e` | string |       — | If set, exports results to CSV at the path.            |

**Examples**

```bash
# ARP scan, export to CSV
discovr active -i eth0 -r 192.168.1.0/24 -e ./out/arp.csv

# ICMP scan with higher concurrency and 3 pings each
discovr active -m -r 10.10.0.0/16 -p 200 -t 2 -c 3 -e ./out/icmp.csv
```

---

### `passive` — Passive network scan

**Synopsis**

```bash
discovr passive [flags]
```

**Description**
Listens to traffic on an interface (no active probes) and identifies devices seen on the network.

**Flags**

|          Flag | Short | Type   | Default | Description                            |
| ------------: | ----: | ------ | ------: | -------------------------------------- |
| `--interface` |  `-i` | string |   `any` | Interface to listen on (e.g., `eth0`). |
|  `--duration` |  `-d` | int    |    `10` | Listening duration in seconds.         |
|    `--export` |  `-e` | string |       — | Export results to CSV.                 |

**Examples**

```bash
# Listen 20 seconds on eth0
discovr passive -i eth0 -d 20 -e ./out/passive.csv

# Use default interface and duration
discovr passive -e ./out/devices.csv
```

---

### `nmap` — Nmap scan wrapper

**Synopsis**

```bash
discovr nmap [flags]
```

**Description**
Runs an Nmap scan against a target IP/CIDR, optionally enabling OS detection.

**Flags**

|          Flag | Short | Type   |     Default | Description                                 |
| ------------: | ----: | ------ | ----------: | ------------------------------------------- |
|    `--target` |  `-t` | string | `127.0.0.1` | Target IP or CIDR.                          |
|     `--ports` |  `-p` | string |  (top 1000) | Ports to scan (e.g., `80,443` or `22-100`). |
| `--detect-os` |  `-d` | bool   |     `false` | Enable OS detection (may require sudo).     |
|    `--export` |  `-e` | string |           — | Export results to CSV.                      |

**Examples**

```bash
discovr nmap -t 127.0.0.1
discovr nmap -t 10.10.10.10 -p 80,443 -d -e ./out/nmap.csv
```

---

### `aws` — AWS EC2 inventory

**Synopsis**

```bash
discovr aws [flags]
```

**Description**
Lists EC2 instances in your AWS account and exports results.

**Flags**

|           Flag | Short | Type     | Default | Description                             |
| -------------: | ----: | -------- | ------: | --------------------------------------- |
|     `--region` |  `-r` | string   |       — | Region filter (e.g., `ap-southeast-2`). |
|    `--profile` |  `-p` | string   |       — | AWS profile name.                       |
|     `--config` |  `-c` | string[] |    `[]` | Custom AWS config file(s).              |
| `--credential` |  `-x` | string[] |    `[]` | Custom AWS credential file(s).          |
|     `--export` |  `-e` | string   |       — | Export results to CSV.                  |

**Examples**

```bash
discovr aws -r us-east-1 -p default -e ./out/aws_ec2.csv
discovr aws -r ap-southeast-2 -c ~/.aws/config -x ~/.aws/credentials -e ./out/ec2.csv
```

---

### `azure` — Azure subscription scan

**Synopsis**

```bash
discovr azure [flags]
```

**Description**
Discovers assets in an Azure subscription and exports to CSV.

**Flags**

|             Flag | Short | Type   |   Default | Description                       |
| ---------------: | ----: | ------ | --------: | --------------------------------- |
| `--subscription` |  `-s` | string | `default` | Azure subscription ID or default. |
|       `--tenant` |  `-t` | string |         — | Tenant ID (optional).             |
|       `--export` |  `-e` | string |         — | Export results to CSV.            |

**Examples**

```bash
discovr azure -s default -e ./out/azure.csv
discovr azure -s 11111111-2222-3333-4444-555555555555 -t 99999999-aaaa-bbbb-cccc-dddddddddddd -e ./out/azure_assets.csv
```

---

### `gcp` — GCP VM inventory

**Synopsis**

```bash
discovr gcp [flags]
```

**Description**
Lists VM instances in specified GCP projects and exports results.

**Flags**

|        Flag | Short | Type   | Default | Description                        |
| ----------: | ----: | ------ | ------: | ---------------------------------- |
| `--project` |  `-p` | string |       — | Comma-separated project names.     |
|    `--cred` |  `-c` | string |       — | Path to service account JSON file. |
|  `--export` |  `-e` | string |       — | Export results to CSV.             |

**Examples**

```bash
discovr gcp -p my-project -c ./sa.json -e ./out/gcp_vms.csv
discovr gcp -p proj-a,proj-b -c ./keys/sa.json -e ./out/vms.csv
```

---

### `tui` — Interactive Text User Interface

**Synopsis**

```bash
discovr tui
```

**Description**
Launches a TUI that guides you through Active, Passive, Nmap, AWS, or Azure scans with interactive prompts and validation.

**Example**

```bash
discovr tui
```

**Notes**

* Good for users who prefer prompts over flags.
* Validates IPs, ports, and subscription IDs.

---

## 3. Output formats & exports

* Most commands support `--export` / `-e` which writes results to CSV.
* CLI prints tabular results to stdout by default.

```
```
