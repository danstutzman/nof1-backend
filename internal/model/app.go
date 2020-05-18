package model

import (
	"database/sql"
)

type Model struct {
	awsAccessKeyId     string
	awsRegion          string
	awsS3Bucket        string
	awsSecretAccessKey string
	dbConn             *sql.DB
	uploadDir          string
}

type Config struct {
	AwsAccessKeyId     string
	AwsRegion          string
	AwsS3Bucket        string
	AwsSecretAccessKey string
	DbConn             *sql.DB
	UploadDir          string
}

func NewModel(config Config) *Model {
	return &Model{
		awsAccessKeyId:     config.AwsAccessKeyId,
		awsRegion:          config.AwsRegion,
		awsS3Bucket:        config.AwsS3Bucket,
		awsSecretAccessKey: config.AwsSecretAccessKey,
		dbConn:             config.DbConn,
		uploadDir:          config.UploadDir,
	}
}
