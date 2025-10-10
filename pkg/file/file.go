// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package file

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GetFileBase64(
	ctx context.Context,
	s3Client *s3.Client,
	bucketName string,
	objectKey string,
) (base64Data string, mimeType string, err error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return "", "", fmt.Errorf("cannot get file from S3: %w", err)
	}
	defer result.Body.Close()

	fileData, err := io.ReadAll(result.Body)
	if err != nil {
		return "", "", fmt.Errorf("cannot read file data: %w", err)
	}

	if result.ContentType == nil || *result.ContentType == "" {
		return "", "", fmt.Errorf("no MIME type available for file %s", objectKey)
	}

	base64Data = base64.StdEncoding.EncodeToString(fileData)
	mimeType = *result.ContentType

	return base64Data, mimeType, nil
}
