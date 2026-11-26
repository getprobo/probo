// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package objectstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"go.probo.inc/probo/pkg/awsconfig"
)

type S3Store struct {
	client *s3.Client
}

var _ Store = (*S3Store)(nil)

func NewS3Store(client *s3.Client) *S3Store {
	return &S3Store{client: client}
}

func (s *S3Store) PutObject(
	ctx context.Context,
	bucket string,
	key string,
	body io.Reader,
	options PutOptions,
) error {
	input := &s3.PutObjectInput{
		Bucket:       new(bucket),
		Key:          new(key),
		Body:         body,
		ContentType:  optionalString(options.ContentType),
		CacheControl: optionalString(options.CacheControl),
		Metadata:     options.Metadata,
	}
	if options.ContentLength != nil {
		input.ContentLength = options.ContentLength
	}

	if options.ContentDisposition != "" {
		input.ContentDisposition = new(options.ContentDisposition)
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("cannot put S3 object: %w", err)
	}

	return nil
}

func (s *S3Store) UploadObject(
	ctx context.Context,
	bucket string,
	key string,
	body io.Reader,
	options PutOptions,
) error {
	uploader := transfermanager.New(
		s.client,
		func(o *transfermanager.Options) {
			o.RequestChecksumCalculation = s.client.Options().RequestChecksumCalculation
		},
	)

	input := &transfermanager.UploadObjectInput{
		Bucket:       new(bucket),
		Key:          new(key),
		Body:         body,
		ContentType:  optionalString(options.ContentType),
		CacheControl: optionalString(options.CacheControl),
		Metadata:     options.Metadata,
	}
	if options.ContentDisposition != "" {
		input.ContentDisposition = new(options.ContentDisposition)
	}

	if _, err := uploader.UploadObject(ctx, input); err != nil {
		return fmt.Errorf("cannot upload S3 object: %w", err)
	}

	return nil
}

func (s *S3Store) HeadObject(ctx context.Context, bucket string, key string) (Properties, error) {
	output, err := s.client.HeadObject(
		ctx,
		&s3.HeadObjectInput{
			Bucket: new(bucket),
			Key:    new(key),
		},
	)
	if err != nil {
		return Properties{}, fmt.Errorf("cannot head S3 object: %w", err)
	}

	properties := Properties{}
	if output.ContentLength != nil {
		properties.ContentLength = *output.ContentLength
	}

	if output.ETag != nil {
		properties.ETag = *output.ETag
	}

	if output.LastModified != nil {
		properties.LastModified = *output.LastModified
	}

	return properties, nil
}

func (s *S3Store) GetObject(
	ctx context.Context,
	bucket string,
	key string,
	options GetOptions,
) (*Object, error) {
	input := &s3.GetObjectInput{
		Bucket: new(bucket),
		Key:    new(key),
	}
	if options.IfNoneMatch != "" {
		input.IfNoneMatch = new(options.IfNoneMatch)
	}

	if !options.IfModifiedSince.IsZero() {
		input.IfModifiedSince = &options.IfModifiedSince
	}

	if options.Range != "" {
		input.Range = new(options.Range)
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		if responseErr, ok := errors.AsType[*smithyhttp.ResponseError](err); ok {
			switch responseErr.HTTPStatusCode() {
			case http.StatusNotModified:
				return &Object{NotModified: true}, nil
			case http.StatusRequestedRangeNotSatisfiable:
				return &Object{RangeNotSatisfiable: true, ContentLength: options.FileSize}, nil
			}
		}

		return nil, fmt.Errorf("cannot get S3 object: %w", err)
	}

	object := &Object{
		Body:          output.Body,
		ContentLength: options.FileSize,
	}
	if output.ContentLength != nil {
		object.ContentLength = *output.ContentLength
	}

	if output.ETag != nil {
		object.ETag = *output.ETag
	}

	if output.LastModified != nil {
		object.LastModified = *output.LastModified
	}

	if output.ContentRange != nil {
		object.ContentRange = *output.ContentRange
		object.PartialContent = true
	}

	return object, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, bucket string, key string) error {
	if _, err := s.client.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: new(bucket),
			Key:    new(key),
		},
	); err != nil {
		return fmt.Errorf("cannot delete S3 object: %w", err)
	}

	return nil
}

func (s *S3Store) PresignGetObject(
	ctx context.Context,
	bucket string,
	key string,
	options PresignOptions,
) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	output, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:                     new(bucket),
			Key:                        new(key),
			ResponseCacheControl:       optionalString(options.ResponseCacheControl),
			ResponseContentType:        optionalString(options.ResponseContentType),
			ResponseContentDisposition: optionalString(options.ResponseContentDisposition),
		},
		awsconfig.UnsignedChecksumMode,
		func(presignOptions *s3.PresignOptions) {
			presignOptions.Expires = options.Expires
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot presign S3 GetObject request: %w", err)
	}

	return output.URL, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return new(value)
}
