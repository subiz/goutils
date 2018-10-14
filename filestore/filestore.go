package file

import (
	"context"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type FileStore interface {
	Add(id string, r io.Reader, contentType string) error
	Get(id string) (io.Reader, error)
}

type GS struct {
	client     *storage.Client
	bucketName string
}

func NewGS(bucketName, credFile string) *GS {
	ctx := context.Background()

	if credFile == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		credFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		panic(err)
	}

	return &GS{
		client:     client,
		bucketName: bucketName,
	}
}

func (s *GS) Add(id string, r io.Reader, contentType string) error {
	ctx := context.Background()
	obj := s.client.Bucket(s.bucketName).Object(id)
	w := obj.NewWriter(ctx)

	if contentType != "" {
		w.ContentType = contentType
	}

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func (s *GS) Get(id string) (io.Reader, error) {
	ctx := context.Background()
	obj := s.client.Bucket(s.bucketName).Object(id)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *GS) MakePublic(id string) error {
	ctx := context.Background()
	acl := s.client.Bucket(s.bucketName).Object(id).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return err
	}

	return nil
}

func (s *GS) MakePrivate(id string) error {
	ctx := context.Background()
	acl := s.client.Bucket(s.bucketName).Object(id).ACL()
	if err := acl.Delete(ctx, storage.AllUsers); err != nil {
		return err
	}

	return nil
}

func (s *GS) GetURL(id string) string {
	return "https://storage.googleapis.com/" + s.bucketName + "/" + id
}
