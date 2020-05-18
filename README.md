# How to run locally

Create IAM user with username=nof1-backend, access type = programmatic access, with custom policy named nof1-backend containing:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:ListBucket",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::danstutzman-nof1-backend/*",
                "arn:aws:s3:::danstutzman-nof1-backend"
            ]
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::danstutzman-nof1-backend/*"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": [
                "transcribe:*",
                "s3:HeadBucket"
            ],
            "Resource": "*"
        }
    ]
}
```

...and save its access key ID and secret to your password manager.

Then run:


```
AWS_PROFILE=personal aws s3 mb s3://danstutzman-nof1-backend

echo '{"Rules": [{"ID":"id-1", "Filter":{}, "Status":"Enabled", "Expiration": {"Days":1 }}]}' > lifecycle.json
AWS_PROFILE=personal aws s3api put-bucket-lifecycle-configuration  \
  --bucket danstutzman-nof1-backend  \
  --lifecycle-configuration file://lifecycle.json 
rm lifecycle.json

go install -race -v ./... && 
  go vet -v ./... && 
  HTTP_PORT=8080 DB_FILE=db/db.sqlite3 STATIC_DIR=static \
    ADMIN_PASSWORD=changeme AWS_S3_BUCKET=danstutzman-nof1-backend \
    AWS_REGION=us-east-1 AWS_ACCESS_KEY_ID=changeme \
    AWS_SECRET_ACCESS_KEY=changeme \
    nof1-backend
```

# How to run automated tests

`go test -v ./...`

Browse to http://localhost:8080/

# How to deploy remotely

`./deploy`
