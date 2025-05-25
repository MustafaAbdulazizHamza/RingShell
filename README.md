# RingShell
![RingShell](https://github.com/MustafaAbdulazizHamza/RingShell/blob/main/ringShell.png)
---

**RingShell** is a lightweight **Command and Control (C2)** framework written in **Golang**, provided for **educational purposes** in offensive security.  
It supports reverse shell payloads and can be extended with user-developed payloads.

RingShell consists of two main components:

1. **Listener** – Interacts with compromised machines, enabling shell access and file operations.
2. **Payload Generator** – Generates platform-specific reverse shell payloads using user-supplied network and system details.

---

## 🛰️ RingShell Listener

A Golang-based interactive interface that enables users to:

- **Manage sessions** with compromised machines:
  - Execute arbitrary commands
  - Upload/download files
  - Take screenshots

- **Set up TCP listeners** bound to specific ports

- **Use scripting** to automate interactions using pre-written RingShell command files

---

## 🛠️ Sauron: The Payload Generator

**Sauron** is a command-line tool that creates reverse shell payloads based on:

- Listener IP 
- Listener Port
- Target OS and architecture

The output is a compiled binary ready for deployment on the target system.

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
```
cd RingShell/RingShell
go build -o RingShell
cd ..
cd Sauron
go build -o sauron
```
---
## ⚠️ Disclaimer
This project is intended for educational and research purposes only.

The developers are not responsible for any misuse or damage caused by this tool.
