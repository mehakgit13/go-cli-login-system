$ErrorActionPreference = "Stop"
docker compose build
docker compose run --rm cli
