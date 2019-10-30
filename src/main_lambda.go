package main

import (
	"beanstalk"
	"encoding/base64"
	"encoding/json"
	"encryption"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"os"
	"time"
)

type BackupEvent struct {
	AssumeRoleArn string `json:"AssumeRoleArn"`
	Bucket        string `json:"Bucket"`
	KMSKeyArn     string `json:"KMSKeyArn"`
	Region        string `json:"Region"`
}

func HandleRequest(event BackupEvent) (string, error) {
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	creds := stscreds.NewCredentials(sess, event.AssumeRoleArn)
	sess, _ = session.NewSession(&aws.Config{
		Region:      aws.String(event.Region),
		Credentials: creds,
	})
	envDetails := beanstalk.FetchAllBeanstalkEnvVars(sess)
	// creating a json out of the main slice
	b, _ := json.MarshalIndent(envDetails, "", "    ")

	encryption.Init()
	encryptionMethod := encryption.AvailableEncryptionMethods["KMSEncryption"]
	encryptionMethod.SetEncryptionArgs(map[string]string{"kmsKeyArn": event.KMSKeyArn})
	encryptionOutput := encryptionMethod.EncryptData(b)

	dataFile, _ := ioutil.TempFile("/tmp", "beanstalk")
	dataFile.Write(encryptionOutput.EncryptedData)

	dataKey := base64.StdEncoding.EncodeToString(encryptionOutput.EncryptedDataKey)
	objectName := "beanstalk_env_vars_backup_" + time.Now().Format("20060102150405") + ".json"

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(event.Bucket),
		Key:      aws.String(objectName),
		Body:     dataFile,
		Metadata: map[string]*string{"dataKey": &dataKey},
	})
	if err != nil {
		// Print the error and exit.
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", objectName, event.Bucket)

	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}
