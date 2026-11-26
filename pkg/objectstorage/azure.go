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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

const (
	maxUserDelegationSASLifetime = 7 * 24 * time.Hour
	userDelegationSASClockSkew   = 5 * time.Minute
)

type AzureStore struct {
	accountURL string
	client     *azblob.Client
}

var _ Store = (*AzureStore)(nil)

func NewAzureStore(accountURL string) (*AzureStore, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create Azure credential: %w", err)
	}

	client, err := azblob.NewClient(accountURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create Azure Blob client: %w", err)
	}

	return &AzureStore{
		accountURL: strings.TrimRight(accountURL, "/"),
		client:     client,
	}, nil
}

func (s *AzureStore) PutObject(
	ctx context.Context,
	container string,
	key string,
	body io.Reader,
	options PutOptions,
) error {
	if _, err := s.client.UploadStream(
		ctx,
		container,
		key,
		body,
		azureUploadOptions(options),
	); err != nil {
		return fmt.Errorf("cannot upload Azure blob: %w", err)
	}

	return nil
}

func (s *AzureStore) UploadObject(
	ctx context.Context,
	container string,
	key string,
	body io.Reader,
	options PutOptions,
) error {
	return s.PutObject(ctx, container, key, body, options)
}

func (s *AzureStore) HeadObject(
	ctx context.Context,
	container string,
	key string,
) (Properties, error) {
	output, err := s.client.ServiceClient().
		NewContainerClient(container).
		NewBlobClient(key).
		GetProperties(ctx, nil)
	if err != nil {
		return Properties{}, fmt.Errorf("cannot get Azure blob properties: %w", err)
	}

	properties := Properties{}
	if output.ContentLength != nil {
		properties.ContentLength = *output.ContentLength
	}

	if output.ETag != nil {
		properties.ETag = string(*output.ETag)
	}

	if output.LastModified != nil {
		properties.LastModified = *output.LastModified
	}

	return properties, nil
}

func (s *AzureStore) GetObject(
	ctx context.Context,
	container string,
	key string,
	options GetOptions,
) (*Object, error) {
	downloadOptions := &azblob.DownloadStreamOptions{}

	if options.IfNoneMatch != "" || !options.IfModifiedSince.IsZero() {
		conditions := &blob.ModifiedAccessConditions{}
		if options.IfNoneMatch != "" {
			conditions.IfNoneMatch = new(azcore.ETag(options.IfNoneMatch))
		}

		if !options.IfModifiedSince.IsZero() {
			conditions.IfModifiedSince = &options.IfModifiedSince
		}

		downloadOptions.AccessConditions = &blob.AccessConditions{
			ModifiedAccessConditions: conditions,
		}
	}

	if options.Range != "" {
		rangeValue, err := parseRange(options.Range, options.FileSize)
		if err != nil {
			return &Object{
				RangeNotSatisfiable: true,
				ContentLength:       options.FileSize,
			}, nil
		}

		downloadOptions.Range = rangeValue
	}

	output, err := s.client.DownloadStream(ctx, container, key, downloadOptions)
	if err != nil {
		if responseErr, ok := errors.AsType[*azcore.ResponseError](err); ok && responseErr.RawResponse != nil {
			switch responseErr.RawResponse.StatusCode {
			case http.StatusNotModified:
				return &Object{NotModified: true}, nil
			case http.StatusRequestedRangeNotSatisfiable:
				return &Object{
					RangeNotSatisfiable: true,
					ContentLength:       options.FileSize,
				}, nil
			}
		}

		return nil, fmt.Errorf("cannot download Azure blob: %w", err)
	}

	object := &Object{
		Body:          output.Body,
		ContentLength: options.FileSize,
	}
	if output.ContentLength != nil {
		object.ContentLength = *output.ContentLength
	}

	if output.ETag != nil {
		object.ETag = string(*output.ETag)
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

func (s *AzureStore) DeleteObject(ctx context.Context, container string, key string) error {
	if _, err := s.client.DeleteBlob(ctx, container, key, nil); err != nil {
		return fmt.Errorf("cannot delete Azure blob: %w", err)
	}

	return nil
}

func (s *AzureStore) PresignGetObject(
	ctx context.Context,
	container string,
	key string,
	options PresignOptions,
) (string, error) {
	if options.Expires <= 0 {
		return "", fmt.Errorf("cannot sign Azure SAS: expiry must be positive")
	}

	if options.Expires > maxUserDelegationSASLifetime-userDelegationSASClockSkew {
		return "", fmt.Errorf("cannot sign Azure SAS: expiry exceeds seven days")
	}

	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(-userDelegationSASClockSkew)
	expiry := now.Add(options.Expires)

	credential, err := s.client.ServiceClient().GetUserDelegationCredential(
		ctx,
		service.KeyInfo{
			Start:  new(start.Format(sas.TimeFormat)),
			Expiry: new(expiry.Format(sas.TimeFormat)),
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot get Azure user delegation key: %w", err)
	}

	query, err := (sas.BlobSignatureValues{
		Protocol:           sas.ProtocolHTTPS,
		StartTime:          start,
		ExpiryTime:         expiry,
		Permissions:        (&sas.BlobPermissions{Read: true}).String(),
		ContainerName:      container,
		BlobName:           key,
		CacheControl:       options.ResponseCacheControl,
		ContentType:        options.ResponseContentType,
		ContentDisposition: options.ResponseContentDisposition,
	}).SignWithUserDelegation(credential)
	if err != nil {
		return "", fmt.Errorf("cannot sign Azure user delegation SAS: %w", err)
	}

	rawURL, err := url.JoinPath(s.accountURL, url.PathEscape(container), url.PathEscape(key))
	if err != nil {
		return "", fmt.Errorf("cannot build Azure blob URL: %w", err)
	}

	blobURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse Azure blob URL: %w", err)
	}

	blobURL.RawQuery = query.Encode()

	return blobURL.String(), nil
}

func azureUploadOptions(options PutOptions) *azblob.UploadStreamOptions {
	uploadOptions := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType:        optionalString(options.ContentType),
			BlobCacheControl:       optionalString(options.CacheControl),
			BlobContentDisposition: optionalString(options.ContentDisposition),
		},
		Metadata: make(map[string]*string, len(options.Metadata)),
	}
	for key, value := range options.Metadata {
		uploadOptions.Metadata[azureMetadataName(key)] = new(value)
	}

	return uploadOptions
}

// azureMetadataName adapts a metadata key to Azure Blob Storage, which only
// accepts C# identifiers as metadata names and therefore rejects the hyphens
// used by our keys.
func azureMetadataName(key string) string {
	return strings.ReplaceAll(key, "-", "_")
}

func parseRange(value string, fileSize int64) (blob.HTTPRange, error) {
	if !strings.HasPrefix(value, "bytes=") {
		return blob.HTTPRange{}, fmt.Errorf("unsupported range unit")
	}

	rangeValue := strings.TrimPrefix(value, "bytes=")
	if strings.Contains(rangeValue, ",") {
		return blob.HTTPRange{}, fmt.Errorf("multiple byte ranges are not supported")
	}

	startEnd := strings.SplitN(rangeValue, "-", 2)
	if len(startEnd) != 2 {
		return blob.HTTPRange{}, fmt.Errorf("invalid byte range")
	}

	if startEnd[0] == "" {
		suffixLength, err := strconv.ParseInt(startEnd[1], 10, 64)
		if err != nil || suffixLength <= 0 || fileSize <= 0 {
			return blob.HTTPRange{}, fmt.Errorf("invalid suffix byte range")
		}

		if suffixLength > fileSize {
			suffixLength = fileSize
		}

		return blob.HTTPRange{
			Offset: fileSize - suffixLength,
			Count:  suffixLength,
		}, nil
	}

	offset, err := strconv.ParseInt(startEnd[0], 10, 64)
	if err != nil || offset < 0 {
		return blob.HTTPRange{}, fmt.Errorf("invalid byte range start")
	}

	rangeValueOut := blob.HTTPRange{Offset: offset}

	if startEnd[1] != "" {
		end, err := strconv.ParseInt(startEnd[1], 10, 64)
		if err != nil || end < offset {
			return blob.HTTPRange{}, fmt.Errorf("invalid byte range end")
		}

		rangeValueOut.Count = end - offset + 1
	}

	return rangeValueOut, nil
}
