package database

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	config "snowflakeservice/config"

	"cloud.google.com/go/storage"
	"github.com/customerio/clock"
	"github.com/pkg/errors"
)

func getConfig(env string) (*config.GCSConfig, error) {
	gcsConfig, err := config.LoadGCSConfig(env)
	if err != nil {
		return nil, err
	}
	return &gcsConfig, nil
}

func newClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx)
}

func UploadPath(ctx context.Context, filename string, bucket string, key string, contentType string) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	return Upload(ctx, f, bucket, key, contentType)
}

func Upload(ctx context.Context, reader io.Reader, bucket, key, contentType string) error {
	gcs, err := newClient(ctx)
	if err != nil {
		return err
	}

	obj := gcs.Bucket(bucket).Object(key)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := writer.Close(); err != nil {
		return errors.WithStack(err)
	}

	_, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{
		ContentType:        contentType,
		ContentDisposition: fmt.Sprintf("attachment; filename=%v", path.Base(key)),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func SignedURL(bucket, filepath string, duration time.Duration, conf config.GCSConfig) (string, error) {
	//key, err := getGCSKey()
	// if err != nil {
	// 	return "", err
	// }

	url, err := storage.SignedURL(bucket, filepath, &storage.SignedURLOptions{
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
