# How to run locally

```
go install -race -v ./... && 
  go vet -v ./... && 
  HTTP_PORT=8080 DB_FILE=db/db.sqlite3 STATIC_DIR=static \
    ADMIN_PASSWORD=changeme nof1-backend
```

# How to run automated tests

`go test -v ./...`

Browse to http://localhost:8080/

# How to deploy remotely

`./deploy`
