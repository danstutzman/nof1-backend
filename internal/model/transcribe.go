package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const LEAVE_TOKEN_EMPTY = ""

func (model *Model) transcribeRecording(recording db.RecordingsRow) {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(
				model.awsAccessKeyId,
				model.awsSecretAccessKey,
				LEAVE_TOKEN_EMPTY),
			Region: aws.String(model.awsRegion),
		},
	}))

	ctx := context.Background()

	uploader := s3manager.NewUploader(awsSession)

	path := fmt.Sprintf("%s/%d/%s",
		model.uploadDir, recording.UserId, recording.Filename)
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Couldn't open %s", path)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	key := fmt.Sprintf("%d_%s", recording.UserId, recording.Filename)
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(model.awsS3Bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		log.Printf("Failed to upload %s to s3://%s: %s",
			key, model.awsS3Bucket, err)
		return
	}

	transcribeService := transcribeservice.New(awsSession)

	uri := fmt.Sprintf("s3://%s/%s", model.awsS3Bucket, key)
	log.Printf("S3 uri: %s", uri)

	_, err = transcribeService.StartTranscriptionJob(
		&transcribeservice.StartTranscriptionJobInput{
			LanguageCode: aws.String("en-US"),
			Media: &transcribeservice.Media{
				MediaFileUri: aws.String(uri),
			},
			TranscriptionJobName: aws.String(key),
		})
	if err != nil {
		log.Printf("Failed to StartTranscriptionJob: %s", err)
		return
	}
	log.Printf("Started Transcription Job")

	var transcriptUri string
	sleepDelaySeconds := 1
	for {
		time.Sleep(time.Duration(sleepDelaySeconds) * time.Second)

		result, err := transcribeService.GetTranscriptionJob(
			&transcribeservice.GetTranscriptionJobInput{
				TranscriptionJobName: aws.String(key),
			})
		if err != nil {
			log.Printf("Failed to GetTranscriptionJob: %s", err)
			return
		}

		if *result.TranscriptionJob.TranscriptionJobStatus != "IN_PROGRESS" {
			transcriptUri = *result.TranscriptionJob.Transcript.TranscriptFileUri
			break
		}
		log.Printf("Job %s still IN_PROGRESS", key)

		if sleepDelaySeconds > 3600 {
			log.Printf("Waiting too long for transcription")
			return
		}
		sleepDelaySeconds *= 2
	}

	s3Service := s3.New(awsSession)

	_, err = s3Service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(model.awsS3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Couldn't delete s3://%s/%s", model.awsS3Bucket, key)
		return
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	response, err := httpClient.Get(transcriptUri)

	if err != nil {
		log.Printf("Failed to download transcription result: %s", err)
		return
	}
	defer response.Body.Close()

	buffer := new(strings.Builder)
	_, err = io.Copy(buffer, response.Body)
	if err != nil {
		log.Printf("Failed to copy transcription result: %s", err)
		return
	}
	log.Printf("AwsTranscribeJson: %s", buffer.String())

	db.UpdateAwsTranscribeJsonOnRecording(
		model.dbConn, buffer.String(), recording.Id)

	if recording.Transcript == "" {
		db.UpdateTranscriptOnRecording(
			model.dbConn, extractTranscriptFromJson(buffer.String()), recording.Id)
	}
}

/* Example JSON:
{"jobName":"2_1589833267","accountId":"553826207523","results":{"transcripts":[{"transcript":"This is the test for my"}],"items":[{"start_time":"1.74","end_time":"1.97","alternatives":[{"confidence":"1.0","content":"This"}],"type":"pronunciation"},{"start_time":"1.97","end_time":"2.11","alternatives":[{"confidence":"1.0","content":"is"}],"type":"pronunciation"},{"start_time":"2.11","end_time":"2.2","alternatives":[{"confidence":"0.9274","content":"the"}],"type":"pronunciation"},{"start_time":"2.2","end_time":"2.66","alternatives":[{"confidence":"1.0","content":"test"}],"type":"pronunciation"},{"start_time":"2.66","end_time":"2.76","alternatives":[{"confidence":"0.7689","content":"for"}],"type":"pronunciation"},{"start_time":"2.76","end_time":"3.02","alternatives":[{"confidence":"0.9595","content":"my"}],"type":"pronunciation"}]},"status":"COMPLETED"}
*/

func extractTranscriptFromJson(s string) string {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		log.Printf("Can't unmarshal JSON: %s", s)
		return ""
	}

	results, ok := m["results"].(map[string]interface{})
	if !ok {
		log.Printf("Can't get .results")
		return ""
	}

	transcripts, ok := results["transcripts"].([]interface{})
	if !ok {
		log.Printf("Can't get .results.transcripts")
		return ""
	}

	transcripts0 := transcripts[0].(map[string]interface{})
	if !ok {
		log.Printf("Can't get .results.transcripts[0]")
		return ""
	}

	transcript, ok := transcripts0["transcript"].(string)
	if !ok {
		log.Printf("Can't get .results.transcripts[0].transcript")
		return ""
	}

	return transcript
}
