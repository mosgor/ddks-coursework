# Evently

A lightweight event discovery platform with vector-based personalized recommendations.

## Project Structure

| Directory / File | Description |
| --- | --- |
| `backend/` | Go REST API. Contains CLI tools, config, Swagger docs, HTTP handlers, business logic, and PostgreSQL/Redis repositories. Includes an ONNX runtime integration for ML inference. |
| `frontend/` | React 19 + TypeScript SPA (Vite). Features client-side routing, Context-based state management, Fuse.js fuzzy search, and CSS modules. |
| `embedder/` | Python pipeline. `generate_embeddings.py` creates event vectors, while `train_two_tower.py` trains and exports a user-tower recommendation model to ONNX. |
| `models/` | Stores the pre-trained `user_tower.onnx` model consumed by the Go recommender. |
| `.github/workflows/` | GitHub Actions CI/CD pipeline for building Docker images and deploying to a VM. |
| `nginx.conf` | Nginx reverse proxy routing & load-balancing configuration. |
| `docker-compose.yml` | Orchestrates PostgreSQL (pgvector), Redis, Go backend, Nginx, React frontend, and the Python embedder. |
| `load-test.js` | k6 script for performance testing the `/events` endpoint. |

## Core Tech

**Go** (Chi, pgx) • **React** + TypeScript • **PostgreSQL** + pgvector • **Redis** • **Python** (PyTorch, SentenceTransformers, ONNX) • **Docker** & **Nginx**
