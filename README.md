# How to run locally

```
go vet -v ./... && 
  go install -v ./... && 
  HTTP_PORT=8080 DB_FILE=db/db.sqlite3 wellsaid-backend
```

Browse to http://localhost:8080/

# How to deploy remotely

`./deploy`
