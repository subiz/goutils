package file

import (
	"cloud.google.com/go/storage"
	"context"
	"git.subiz.net/errors"
	"google.golang.org/api/option"
	"io"
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

	if credFile == "" {
		panic(errors.New(400, errors.E_invalid_filestore_credential,
			"credFile cannot be empty"))
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		panic(errors.Wrap(err, 400, errors.E_invalid_filestore_credential))
	}

	return &GS{client: client, bucketName: bucketName}
}

func (s *GS) Add(id string, r io.Reader, contentType string) error {
	ctx := context.Background()
	obj := s.client.Bucket(s.bucketName).Object(id)
	w := obj.NewWriter(ctx)

	if contentType != "" {
		w.ContentType = contentType
	}

	if _, err := io.Copy(w, r); err != nil {
		return errors.Wrap(err, 500, errors.E_filestore_write_error)
	}

	if err := w.Close(); err != nil {
		return errors.Wrap(err, 500, errors.E_filestore_write_error)
	}

	return nil
}

func (s *GS) Get(id string) (io.Reader, error) {
	ctx := context.Background()
	obj := s.client.Bucket(s.bucketName).Object(id)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_filestore_read_error)
	}

	return r, nil
}

func (s *GS) MakePublic(id string) error {
	ctx := context.Background()
	acl := s.client.Bucket(s.bucketName).Object(id).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return errors.Wrap(err, 500, errors.E_filestore_acl_error)
	}

	return nil
}

func (s *GS) MakePrivate(id string) error {
	ctx := context.Background()
	acl := s.client.Bucket(s.bucketName).Object(id).ACL()
	if err := acl.Delete(ctx, storage.AllUsers); err != nil {
		return errors.Wrap(err, 500, errors.E_filestore_acl_error)
	}

	return nil
}

func (s *GS) GetURL(id string) string {
	return "https://storage.googleapis.com/" + s.bucketName + "/" + id
}
