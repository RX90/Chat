# Chat Application

This is a Websocket chat application with backend (Go), frontend (Vite + HTML/CSS/JS) and database (PostgreSQL). 

---

## Prerequisites

- Go (version 1.20+)
- Node.js (version 16+)
- Docker

---

## Getting Started

### 1. Clone the repository

### 2. Add .env with POSTGRES_PASSWORD and AUTH_KEY like in .env.example

### 3. Run frontend server
```bash
cd web && npm install && npm run dev
```

### 4. Run backend server
```bash
docker compose up -d
```

### Usage

Open the Network address shown by Vite (for example `http://192.168.0.103:3000`).  
After that, you can register and start chatting.