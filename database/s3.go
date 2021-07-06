package database

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	config "snowflakeservice/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func newSession() (*session.Session, error) {

	sess, err := session.NewSession(&aws.Config{Region: aws.String(config.AWS_REGION)})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func UploadFile(awsSession *session.Session, file *os.File) (string,error) {

          fileInfo, _ := file.Stat()
          size := fileInfo.Size()
          buffer := make([]byte, size)
          file.Read(buffer)
          key := fmt.Sprintf("%s%s", config.AWS_S3_FILEPATH, file.Name())
        
          _, err := s3.New(awsSession).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String(config.AWS_S3_BUCKET),
				Key:                  aws.String(key),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(buffer),
				ContentLength:        aws.Int64(size),
				ContentType:          aws.String(http.DetectContentType(buffer)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})

		if(err != nil){
			return "", err
		}
         
		 return key, nil
}
