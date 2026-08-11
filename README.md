# AI Support Agent

A retrieval-augmented generation (RAG) demo for document-based customer support.

## Features

- Upload PDF documents
- Automatic text extraction and chunking
- Embeddings with Alibaba Cloud Bailian
- Vector search with PostgreSQL + pgvector
- Grounded answers with DeepSeek
- Source citations
- Duplicate document detection
- Document deletion
- Next.js chat interface
- Docker Compose setup

## Tech Stack

### Backend

- Go
- Gin
- PostgreSQL
- pgvector

### AI

- Alibaba Cloud Bailian `text-embedding-v4`
- DeepSeek

### Frontend

- Next.js
- TypeScript
- Tailwind CSS

## Architecture

```text
PDF Upload
    ↓
Text Extraction
    ↓
Chunking
    ↓
Embedding
    ↓
PostgreSQL + pgvector
    ↓
Semantic Search
    ↓
DeepSeek
    ↓
Answer + Sources
```

## Prerequisites

- Docker Desktop
- DeepSeek API Key
- Alibaba Cloud Bailian API Key

## Setup

Copy the environment template:

```bash
cp .env.example .env
```

On Windows PowerShell, you can use:

```powershell
Copy-Item .env.example .env
```

Fill in the required environment variables:

```env
DEEPSEEK_API_KEY=
DASHSCOPE_API_KEY=
BAILIAN_BASE_URL=
```

Start the application:

```bash
docker compose up -d --build
```

Open the web interface:

```text
http://localhost:3000
```

Backend API:

```text
http://localhost:8080
```

## Main API Endpoints

```text
GET    /health
GET    /health/db

GET    /api/documents
POST   /api/documents/upload
GET    /api/documents/:id
DELETE /api/documents/:id

POST   /api/chat
```

## Example

Upload a refund policy PDF and ask:

```text
How long do customers have to request a refund?
```

The application:

1. Embeds the user's question.
2. Searches PostgreSQL/pgvector for relevant document chunks.
3. Filters results using a relevance threshold.
4. Sends the retrieved context to DeepSeek.
5. Returns a grounded answer with source references.

If no sufficiently relevant information is found, the assistant does not answer from the model's general knowledge.

## Local Development

### Backend

```bash
cd backend
go run ./cmd/server
```

The backend runs on:

```text
http://localhost:8080
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on:

```text
http://localhost:3000
```

### PostgreSQL

The Docker Compose configuration exposes PostgreSQL on:

```text
localhost:5433
```

The backend container communicates with PostgreSQL internally using the Docker service name:

```text
postgres:5432
```

## Docker Services

The complete application consists of three services:

```text
frontend
    ↓
backend
    ↓
postgres + pgvector
```

Start everything:

```bash
docker compose up -d --build
```

Check service status:

```bash
docker compose ps
```

Stop everything:

```bash
docker compose down
```

## Project Status

This project is a portfolio/demo implementation focused on the core RAG workflow and practical document-based AI support use cases.