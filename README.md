# personal_wiki_project

## Local development with Azurite (Azure Blob emulator)

This project can use Azurite to emulate Azure Blob Storage locally.

1. Copy environment file and set variables (or use the provided .env.development):

```bash
export $(cat .env.development | xargs)
```

2. Start services (includes Azurite):

```bash
docker-compose up -d
```

3. Test upload endpoint:

```bash
curl -X POST "http://localhost:8080/api/upload" \
	-H "X-User-ID: user123" \
	-F "file=@./test.jpg" \
	-F "title=My photo" \
	-F "description=desc"
```

Azutite stores blob files under `./azurite_data` by default in this repository.
