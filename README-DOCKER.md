
# 🐳 Running Paqet with Docker

We now support a fully Dockerized deployment for both Client and Server. This is the recommended way to run Paqet as it handles dependencies (`libpcap`) and firewall rules (`iptables`) automatically.

## Prerequisites

- [Docker Engine](https://docs.docker.com/engine/install/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## 🚀 Quick Start Guide

### 1. Common Setup (Do this on BOTH Server and Client)

1.  **Clone the repo** (or copy files):
    ```bash
    git clone https://github.com/your/repo.git
    cd paqet
    ```
2.  **Create `.env` file** (Secrets):
    ```bash
    cp .env.example .env  # (Create this if missing, see below)
    nano .env
    ```
    *Set `PAQET_SERVER_ADDR` and `PAQET_SECRET_KEY`.*

3.  **Create `config.yaml`**:
    ```bash
    cp example/client.yaml.example config.yaml
    nano config.yaml
    ```
    *Ensure `role: "server"` or `role: "client"` is set correctly.*

---

### 2. Run on SERVER ☁️ (VPS/Linux)

1.  **Edit `config.yaml`**:
    ```yaml
    role: "server"
    transport:
      kcp:
        key: "${PAQET_SECRET_KEY}" # Uses value from .env
    server:
      listen: "0.0.0.0:9999"      # Or use ${PAQET_SERVER_ADDR}
    ```
2.  **Start Server**:
    ```bash
    docker compose up -d server
    ```
3.  **Verify**:
    ```bash
    docker compose logs -f server
    ```
    *You should see "iptables rules applied" in the logs.*

---

### 3. Run on CLIENT 💻 (Mac/Windows/Linux)

1.  **Edit `config.yaml`**:
    ```yaml
    role: "client"
    network:
      interface: "auto"        # Auto-detects interface (Important for Docker!)
    socks5:
      - listen: "0.0.0.0:1080" # Bind to all interfaces to access from host
    transport:
      kcp:
        key: "${PAQET_SECRET_KEY}"
    server:
      addr: "${PAQET_SERVER_ADDR}"
    ```
2.  **Start Client**:
    ```bash
    docker compose up -d client
    ```
3.  **Connect**:
    *   Configure your browser/system proxy to use **SOCKS5** at `127.0.0.1:1090` (Mapped port in docker-compose).

## Troubleshooting

**"Operation not permitted"**
Ensure you are running with the required capabilities. The provided `docker-compose.yml` already includes `cap_add: [NET_ADMIN, NET_RAW]`.

**"no such network interface" (Mac Users)**
On macOS, Docker containers run inside a Linux VM. They see `eth0` instead of `en0`.
Update your `config.yaml`:
- `interface: "eth0"`
- `socks5.listen: "0.0.0.0:1080"` (Allow access from host)

**"config.yaml not found"**
Ensure `config.yaml` exists in the same directory as `docker-compose.yml` before starting.
