# OrbStack Installation and Usage Guide

> **Target**: Comprehensive operational manual and step-by-step developer guide for installing, configuring, and using **OrbStack** as a fast, lightweight Docker and Linux virtualization engine on macOS for GraphFolio development.  
> **Key Principle**: Copy-pasteable CLI commands, clear architectural comparisons with Docker Desktop, local container orchestration for GraphFolio services, and advanced productivity tips.

---

## Table of Contents
1. [Overview & Why OrbStack?](#1-overview--why-orbstack)
2. [System Requirements & Prerequisites](#2-system-requirements--prerequisites)
3. [Step 1: Installation](#3-step-1-installation)
4. [Step 2: Migration & Coexistence with Docker Desktop](#4-step-2-migration--coexistence-with-docker-desktop)
5. [Step 3: Core Concepts & Architecture](#5-step-3-core-concepts--architecture)
6. [Step 4: Step-by-Step Guide to Using OrbStack with GraphFolio](#6-step-by-step-guide-to-using-orbstack-with-graphfolio)
   - [4.1 Running Local PostgreSQL in a Container](#41-running-local-postgresql-in-a-container)
   - [4.2 Building and Running GraphFolio Backend Container](#42-building-and-running-graphfolio-backend-container)
   - [4.3 Multi-Container Local Stack with Docker Compose](#43-multi-container-local-stack-with-docker-compose)
   - [4.4 Working with Lightweight Linux Machines (VMs)](#44-working-with-lightweight-linux-machines-vms)
   - [4.5 Enabling and Using Local Kubernetes (k8s)](#45-enabling-and-using-local-kubernetes-k8s)
7. [Step 5: OrbStack CLI Command Reference](#7-step-5-orbstack-cli-command-reference)
8. [Step 6: OrbStack GUI & Productivity Features](#8-step-6-orbstack-gui--productivity-features)
9. [Step 7: Performance Tuning & Resource Management](#9-step-7-performance-tuning--resource-management)
10. [Troubleshooting & Frequently Asked Questions](#10-troubleshooting--frequently-asked-questions)

---

## 1. Overview & Why OrbStack?

[OrbStack](https://orbstack.dev) is a native, ultra-fast, lightweight alternative to Docker Desktop built specifically for macOS (Apple Silicon and Intel). It runs Docker containers, Docker Compose, Linux machines (virtual machines), and Kubernetes clusters with minimal overhead.

### OrbStack vs. Docker Desktop Comparison

| Feature | Docker Desktop | OrbStack |
|---|---|---|
| **Startup Time** | 15 – 45 seconds | **~1 – 2 seconds** |
| **Idle Memory Consumption** | 2 GB – 4 GB RAM | **~100 MB – 200 MB RAM** (dynamic allocation) |
| **CPU Overhead** | Often 5% – 20% background usage | **< 0.1% idle CPU** (instant sleep when idle) |
| **File Sharing Speed** | VirtioFS / gRPC-FUSE with noticeable lag | **Native macOS file system speed** (shared memory / 9p) |
| **Networking & Local Domains** | Requires explicit `-p` port mappings | **Automatic `.orb.local` domain** per container + standard ports |
| **Linux Machines** | Requires custom Dockerfiles / slow VMs | **Instant Ubuntu, Debian, Alpine, Fedora VMs** with `/Users` access |
| **Battery Life Impact** | High battery drain on MacBooks | **Negligible battery drain** |
| **CLI Compatibility** | Standard `docker` / `docker compose` | **100% Drop-in replacement** for `docker` & `docker compose` |

---

## 2. System Requirements & Prerequisites

- **Operating System**: macOS 12.0 (Monterey), macOS 13.0 (Ventura), macOS 14.0 (Sonoma), or macOS 15.0+ (Sequoia).
- **Architecture**:
  - **Apple Silicon (M1, M2, M3, M4)**: Native ARM64 support.
  - **Intel Mac**: Native x86_64 support.
- **Rosetta 2** *(Recommended for Apple Silicon users running AMD64/x86 containers)*:
  If you need to run Intel-based x86_64 Docker containers on Apple Silicon, verify that Rosetta 2 is installed:
  ```bash
  softwareupdate --install-rosetta --agree-to-license
  ```
- **macOS Permissions**: Administrator privileges are required during the initial launch to install the helper tool and set up UNIX domain sockets (`/var/run/docker.sock`).

---

## 3. Step 1: Installation

### Method A: Homebrew Cask (Recommended)

Open your macOS terminal and run:

```bash
brew install --cask orbstack
```

### Method B: Direct Download (DMG)

1. Navigate to the official download page: [https://orbstack.dev/download](https://orbstack.dev/download).
2. Download the `.dmg` file matching your Mac's processor architecture (Apple Silicon `Apple chip` or `Intel chip`).
3. Open `OrbStack.dmg` and drag the **OrbStack.app** icon into `/Applications`.

### Initial Setup & First Launch

1. Launch **OrbStack** from `/Applications` or via Spotlight (`Cmd + Space` -> `OrbStack`).
2. Follow the onboarding setup wizard:
   - **Helper Tool Installation**: Enter your macOS administrator password when prompted. This configures the lightweight virtualization network bridge and Docker socket symlink.
   - **Docker Context**: Allow OrbStack to register itself as the primary Docker engine context.
3. Verify the installation via your terminal:
   ```bash
   # Check OrbStack CLI version
   orb version

   # Verify Docker CLI points to OrbStack
   docker version
   ```
   *Expected Output*:
   ```text
   Client:
    Version:           27.x.x
   Server: OrbStack
    Engine:
     Version:          27.x.x
   ```

4. Confirm the Docker info reflects OrbStack:
   ```bash
   docker info | grep "Operating System"
   ```
   *Expected Output*:
   ```text
   Operating System: OrbStack
   ```

---

## 4. Step 2: Migration & Coexistence with Docker Desktop

OrbStack seamlessly replaces Docker Desktop without requiring you to rewrite any Dockerfiles, scripts, or Makefile targets.

### Setting the Docker Context

OrbStack uses Docker's native context switching mechanism. You can switch between Docker Desktop and OrbStack at any time.

1. **List all Docker contexts**:
   ```bash
   docker context ls
   ```
   *Example Output*:
   ```text
   NAME        DESCRIPTION                               DOCKER ENDPOINT
   default     Current DOCKER_HOST                       unix:///var/run/docker.sock
   desktop-linux Docker Desktop                          unix:///Users/username/.docker/run/docker.sock
   orbstack *  OrbStack                                  unix:///Users/username/.orbstack/run/docker.sock
   ```

2. **Activate OrbStack as the active context**:
   ```bash
   docker context use orbstack
   ```

3. **Verify socket connectivity**:
   ```bash
   docker ps
   ```

### Coexisting with Docker Desktop
If you choose to keep Docker Desktop installed:
- Prevent both applications from launching simultaneously at startup to avoid port and socket collisions.
- In **Docker Desktop > Settings > General**, uncheck **"Start Docker Desktop when you log in"**.
- In **OrbStack > Settings > General**, ensure **"Start at login"** is enabled.

### Migrating Volumes and Images (Optional)
OrbStack maintains its own isolated, lightweight storage pool. If you have containers, volumes, or images in Docker Desktop that you wish to migrate:
- OrbStack includes a built-in migration prompt on first launch if it detects an existing Docker Desktop installation.
- Alternatively, push your custom images to a registry (e.g., Docker Hub or GitHub Packages) or re-build them using `make docker-build`.

---

## 5. Step 3: Core Concepts & Architecture

OrbStack introduces key architectural advantages that streamline everyday development:

```
┌────────────────────────────────────────────────────────────────────────┐
│                              macOS Host                                │
│                                                                        │
│   Terminal / IDE / Web Browser         OrbStack GUI (Menubar & App)    │
│   [ localhost:8080 ] [ *.orb.local ]   [ Resource Monitor: ~150MB ]    │
└─────────────────────────────────┬──────────────────────────────────────┘
                                  │ Direct Memory & VirtioFS
┌─────────────────────────────────▼──────────────────────────────────────┐
│                    OrbStack Micro-Virtualization                       │
│                                                                        │
│   ┌───────────────────────────┐      ┌───────────────────────────────┐ │
│   │     Docker Engine         │      │     Linux Machines (VMs)      │ │
│   │  • Containers (Postgres)  │      │  • Ubuntu 24.04 (dev-box)     │ │
│   │  • GraphFolio Backend     │      │  • Alpine / Debian / Fedora   │ │
│   │  • Automatic DNS Routing  │      │  • Instant bidirectional fs   │ │
│   └───────────────────────────┘      └───────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────┘
```

1. **Zero-Configuration Local Domain Names (`*.orb.local`)**:
   Every running container with an exposed port is automatically reachable at:
   ```text
   http://<container-name>.orb.local:<port>
   ```
   For example, a container named `graphfolio-backend` with port `8080` can be accessed via:
   - `http://localhost:8080`
   - `http://graphfolio-backend.orb.local:8080`
2. **Dynamic Memory Allocation**:
   Unlike Docker Desktop, which reserves a fixed block of RAM (e.g., 8 GB) even when idle, OrbStack allocates memory dynamically. When your containers are idle, memory is returned to macOS instantly.
3. **Transparent File Sharing**:
   Your entire macOS home directory (`/Users/<username>`) is instantly available inside OrbStack Linux machines without CPU-heavy syncing overhead.

---

## 6. Step-by-Step Guide to Using OrbStack with GraphFolio

### 4.1 Running Local PostgreSQL in a Container

If you prefer running PostgreSQL inside a container instead of a native macOS Homebrew service, launch PostgreSQL via OrbStack:

```bash
docker run -d \
  --name graphfolio-postgres \
  -p 5432:5432 \
  -e POSTGRES_USER=$(whoami) \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  -v graphfolio-pgdata:/var/lib/postgresql/data \
  postgres:16-alpine
```

Verify that PostgreSQL is healthy:
```bash
# Check container status
docker ps --filter "name=graphfolio-postgres"

# Check readiness using pg_isready
pg_isready -h localhost -p 5432
```

Now execute the GraphFolio database bootstrap and migrations:
```bash
# Bootstrap databases and schemas
make db-bootstrap

# Run migrations and seed development fixtures
make migrate-up
make db-seed
```

---

### 4.2 Building and Running GraphFolio Backend Container

GraphFolio includes a multi-stage production container image defined at [`deploy/docker/Dockerfile`](file:///Users/oscargarcia/workspace/graphfolio/deploy/docker/Dockerfile) containing the unified backend (`bff`, `portfolio-api`, `user-api`, and `portfolio-worker`).

#### Step 1: Build the Image with OrbStack
From the repository root:
```bash
make docker-build
```
*Behind the scenes*: OrbStack leverages Buildx with native ARM64 speed on Apple Silicon, finishing in seconds.

#### Step 2: Run the Unified Backend Container
```bash
docker run --rm -d \
  --name graphfolio-backend \
  -p 8080:8080 \
  --env-file .env \
  graphfolio-backend:local
```

#### Step 3: Verify Container Health & DNS Routing
OrbStack exposes the container through both standard localhost and native DNS:

1. **Localhost Endpoint**:
   ```bash
   curl -I http://localhost:8080/
   ```
2. **OrbStack Domain Resolution**:
   Open your browser to:
   ```text
   http://graphfolio-backend.orb.local:8080/
   ```
   *Expected Output*: The interactive GraphQL Playground will load instantly.
3. **Execute a GraphQL Query**:
   ```bash
   curl -X POST http://localhost:8080/query \
     -H "Content-Type: application/json" \
     -d '{"query": "{ __typename }"}'
   ```
   *Expected Output*: `{"data":{"__typename":"Query"}}`

4. **Inspect Container Logs via CLI or OrbStack GUI**:
   ```bash
   docker logs -f graphfolio-backend
   ```

---

### 4.3 Multi-Container Local Stack with Docker Compose

You can orchestrate your entire GraphFolio environment (PostgreSQL + Backend + Web Apps) using Docker Compose.

Create a `docker-compose.yml` file in the repository root or use this configuration:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: graphfolio-postgres
    environment:
      POSTGRES_USER: ${USER:-postgres}
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: postgres
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${USER:-postgres}"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: .
      dockerfile: deploy/docker/Dockerfile
    container_name: graphfolio-backend
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "8080:8080"
    environment:
      - PORTFOLIO_DB_URL=postgres://portfolio_svc:portfolio@postgres:5432/graphfolio?sslmode=disable
      - USER_DB_URL=postgres://user_svc:user@postgres:5432/graphfolio?sslmode=disable
      - BFF_PORT=8080
      - PORTFOLIO_SVC_ADDR=127.0.0.1:50051
      - USER_SVC_ADDR=127.0.0.1:50052
      - JWT_SECRET=dev-secret-graphfolio-local-test-key-32b
    env_file:
      - .env

volumes:
  pgdata:
```

#### Managing the Stack:
```bash
# Start all services in the background
docker compose up -d

# Check running status and ports
docker compose ps

# View logs
docker compose logs -f backend

# Stop the stack
docker compose down
```

---

### 4.4 Working with Lightweight Linux Machines (VMs)

One of OrbStack's most powerful features is **Linux Machines**. Unlike standard virtual machines (which take minutes to provision and gigabytes of disk), OrbStack provisions full Linux distributions in **less than 2 seconds**.

This is ideal for testing Go binaries on Linux, verifying shell scripts like [`deploy/docker/entrypoint.sh`](file:///Users/oscargarcia/workspace/graphfolio/deploy/docker/entrypoint.sh), or debugging Alpine/Ubuntu packaging issues.

#### Step 1: Create a Linux Machine
```bash
# Create an Ubuntu 24.04 LTS machine named 'dev-box'
orb create ubuntu:24.04 dev-box

# Or create an Alpine Linux machine
orb create alpine:latest alpine-box
```

#### Step 2: Enter the Linux Machine
```bash
orb enter dev-box
```
You are immediately dropped into a full Linux Bash shell.

#### Step 3: Seamless File Access
Inside the machine, your macOS filesystem is directly accessible:
```bash
# Navigate to your GraphFolio project from inside Linux
cd /Users/$(whoami)/workspace/graphfolio

# Verify files are visible
ls -la
```
Any file you edit in your IDE on macOS is immediately updated inside the Linux machine with zero lag.

#### Step 4: Run Commands from macOS Host
You don't even need to enter the shell to execute commands inside Linux:
```bash
# Run a Go build test inside the Linux environment
orb -m dev-box go version

# Check OS architecture inside the VM
orb -m dev-box uname -a
```

#### Step 5: Managing Machines
```bash
# List all machines and their resource state
orb list

# Stop a machine
orb stop dev-box

# Start a stopped machine
orb start dev-box

# Delete a machine when done
orb delete dev-box
```

---

### 4.5 Enabling and Using Local Kubernetes (k8s)

OrbStack includes an integrated, single-node Kubernetes cluster that starts in under 5 seconds with minimal memory overhead.

1. **Enable Kubernetes**:
   - Open OrbStack Preferences (`Cmd + ,`).
   - Navigate to the **Kubernetes** tab.
   - Toggle **"Enable Kubernetes"** ON.
2. **Switch your `kubectl` context**:
   ```bash
   kubectl config use-context orbstack
   ```
3. **Verify the cluster**:
   ```bash
   kubectl get nodes
   ```
   *Expected Output*:
   ```text
   NAME       STATUS   ROLES           AGE   VERSION
   orbstack   Ready    control-plane   10s   v1.30.x
   ```
4. Deploy test manifests or Helm charts directly to your local cluster.

---

## 7. Step 5: OrbStack CLI Command Reference

The `orb` command-line utility controls all aspects of the virtualization engine:

| Command | Description | Example Usage |
|---|---|---|
| `orb` | Enter the default Linux machine shell | `orb` |
| `orb enter <machine>` | Enter a specific Linux machine | `orb enter dev-box` |
| `orb create <distro> [name]` | Create a new Linux machine | `orb create ubuntu:24.04 dev-box` |
| `orb list` / `orb ls` | List all running and stopped Linux machines | `orb list` |
| `orb start <machine>` | Start a stopped machine | `orb start dev-box` |
| `orb stop <machine>` | Gracefully stop a running machine | `orb stop dev-box` |
| `orb restart <machine>` | Restart a machine | `orb restart dev-box` |
| `orb delete <machine>` | Permanently delete a machine and its disk | `orb delete dev-box` |
| `orb run -m <machine> <cmd>` | Execute a single command inside a machine | `orb run -m dev-box uname -m` |
| `orb push <src> <dest>` | Copy a file from macOS into a machine | `orb push ./config.json dev-box:/etc/` |
| `orb info` | Display detailed system & engine metrics | `orb info` |
| `orb version` | Print OrbStack version & build info | `orb version` |
| `orb logs` | Stream OrbStack system logs for debugging | `orb logs` |
| `orb clean` | Reclaim unused disk space across containers & VMs | `orb clean` |

---

## 8. Step 6: OrbStack GUI & Productivity Features

While the CLI provides complete control, OrbStack's native macOS interface offers rich productivity features:

### 1. Menubar Status & Resource Monitor
- Click the **OrbStack icon** in your macOS menu bar to view real-time CPU, RAM, and disk utilization.
- Quickly toggle engine pause/resume, restart the Docker daemon, or open the settings.

### 2. Container Inspector
- Click on any container in the **Containers** view to:
  - **View Real-Time Logs**: Searchable, colorized streaming logs with instant auto-scroll toggle.
  - **Terminal / Exec**: Open an embedded interactive terminal session directly in the GUI.
  - **File Browser**: Browse, download, or inspect files inside the container's virtual filesystem.
  - **Port Forwards**: Click on mapped ports (`8080`, `5432`) to open them immediately in your default browser.

### 3. Image & Volume Management
- **Images View**: Review disk space consumed by base images (`postgres:16-alpine`, `golang:1.24-alpine`). Prune dangling or unused images with a single click.
- **Volumes View**: Inspect named volumes (`graphfolio-pgdata`) and export or delete persistent data.

---

## 9. Step 7: Performance Tuning & Resource Management

### Memory & CPU Allocation
Unlike older container runtimes, OrbStack automatically manages system resources:
- **CPU**: Defaults to sharing all available CPU cores without throttling. Idle containers consume 0% host CPU.
- **RAM**: Memory is allocated dynamically on demand. If a container finishes a heavy build, OrbStack immediately releases the memory back to macOS.
- **Custom Limits**: To set strict upper bounds (e.g., if you want to test memory leak resilience):
  1. Open **OrbStack > Settings > CPU & Memory**.
  2. Set **Memory Limit** (e.g., `8 GB`) and **CPU Limit** (e.g., `4 cores`).

### x86_64 Emulation on Apple Silicon
OrbStack integrates Rosetta 2 for running Intel (`linux/amd64`) images on Apple Silicon Macs:
- Enable via **OrbStack > Settings > General > Intel (x86_64) emulation: Rosetta**.
- To build or run an amd64 container explicitly:
  ```bash
  docker run --platform linux/amd64 --rm alpine uname -m
  # Output: x86_64
  ```

---

## 10. Troubleshooting & Frequently Asked Questions

### Q1: `docker ps` gives "Cannot connect to the Docker daemon"
**Cause**: The Docker CLI is looking for the Docker Desktop socket or an outdated context.  
**Fix**:
1. Check your active context:
   ```bash
   docker context ls
   ```
2. Switch to OrbStack:
   ```bash
   docker context use orbstack
   ```
3. If `/var/run/docker.sock` was claimed by another tool, recreate the symlink:
   ```bash
   sudo ln -sf ~/.orbstack/run/docker.sock /var/run/docker.sock
   ```

---

### Q2: Port 5432 or 8080 is already in use
**Cause**: A native PostgreSQL instance (e.g., installed via Homebrew) or another local service is already bound to port 5432.  
**Fix**:
1. Find what process is holding the port:
   ```bash
   lsof -i :5432
   ```
2. Stop the conflicting service:
   ```bash
   brew services stop postgresql@16
   ```
3. Alternatively, map the container to an alternative port:
   ```bash
   docker run -d -p 5433:5432 --name graphfolio-postgres postgres:16-alpine
   ```

---

### Q3: `.orb.local` domains are not resolving in browser
**Cause**: Some corporate VPNs (Cisco AnyConnect, GlobalProtect, Tailscale) override macOS DNS resolution.  
**Fix**:
1. Fall back to standard `localhost:<port>` (e.g., `http://localhost:8080`).
2. Verify DNS resolution from terminal:
   ```bash
   ping graphfolio-backend.orb.local
   ```
3. If using Tailscale or custom DNS, add `orb.local` to your split-DNS search domains.

---

### Q4: How do I completely reclaim disk space?
**Fix**:
Run the built-in clean utility:
```bash
# Reclaim unused container and VM disk space
orb clean

# Prune all unused Docker data
docker system prune -a --volumes -f
```

---

### Related Documentation & Links
- [Cloudflare Deployment Step-by-Step Guide](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/cloudflare-deployment-step-by-step-guide.md)
- [GraphFolio Backend Dockerfile](file:///Users/oscargarcia/workspace/graphfolio/deploy/docker/Dockerfile)
- [GraphFolio Root Makefile](file:///Users/oscargarcia/workspace/graphfolio/Makefile)
- [Official OrbStack Documentation](https://docs.orbstack.dev)
