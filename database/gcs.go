package database

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	config "snowflakeservice/config"

	"cloud.google.com/go/storage"
	"github.com/customerio/clock"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

const (
	PART_BUCKET    = "sf_reports_"
	CONTENT_TYPE   = "text/csv"
	PART_CONF_FILE = "gsc_"
)

var currentBucket string
var currentConf string

func getConfig(env string) (*config.GCSConfig, error) {
	env = strings.ToLower(env)
	gcsConfig, err := config.LoadGCSConfig(env)
	if err != nil {
		return nil, err
	}

	currentBucket = PART_BUCKET + env
	currentConf = PART_CONF_FILE + env + ".json"

	return &gcsConfig, nil

}

func newClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithCredentialsFile(currentConf))
}

func UploadPath(filepath string, nameOfObject string) error {
	ctx := context.Background()
	f, err := os.Open(filepath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	return upload(ctx, f, currentBucket, nameOfObject)
}

func upload(ctx context.Context, reader io.Reader, bucket, nameOfObject string) error {
	gcs, err := newClient(ctx)
	if err != nil {
		log.Print("new client error")
		return err
	}

	obj := gcs.Bucket(bucket).Object(nameOfObject)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Print("upload error")
		return errors.WithStack(err)
	}

	if err := writer.Close(); err != nil {
		log.Print("writer close error")
		return errors.WithStack(err)
	}

	cloudFilePath := path.Base(nameOfObject)
	_, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{
		ContentType:        CONTENT_TYPE,
		ContentDisposition: fmt.Sprintf("attachment; filename=%v", cloudFilePath),
	})
	if err != nil {
		log.Print("content type error")
		return errors.WithStack(err)
	}

	return nil
}

func SignedURL(filepath string, duration time.Duration, conf config.GCSConfig) (string, error) {

	url, err := storage.SignedURL(currentBucket, filepath, &storage.SignedURLOptions{
		GoogleAccessID: conf.Client_Email,
		PrivateKey:     []byte(conf.Private_Key),
		Method:         "GET",
		Expires:        clock.Now().Add(duration),
	})
	if err != nil {
		return "", err
	}

	return url, nil
}
