# Chat Application

This is a simple chat application with backend (Go) and frontend (Vite + React) with WebSocket connection.

---

## Prerequisites

- Go (version 1.20+)
- Node.js (version 16+)
- Docker and Docker Compose
- Git

---

## Getting Started

### 1. Clone the repository

### 2. Run frontend development server
```bash
cd web && npm install & npm run dev
```

### 3. Start backend service with Docker Compose
```bash
docker-compose up -d
```

### Configuration
Backend configuration is located in config.yaml (e.g. server port, database settings, CORS).

Frontend settings are in web folder.

### Usage
Open http://localhost:3000 in your browser.

Register or log in to start chatting.

The backend API runs on port 8080 by default.