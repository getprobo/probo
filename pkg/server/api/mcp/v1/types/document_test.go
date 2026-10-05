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

package types_test

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func TestNewDocument_IncludesLatestVersionTitle(t *testing.T) {
	t.Parallel()

	document := types.NewDocument(&coredata.Document{
		Title:        "Access control",
		DocumentType: coredata.DocumentTypePolicy,
		Status:       coredata.DocumentStatusActive,
		WriteMode:    coredata.DocumentWriteModeAuthored,
	})

	assert.Equal(t, "Access control", document.Title)
	assert.Equal(t, coredata.DocumentTypePolicy, document.DocumentType)
}

func TestDocumentToolSchemasNameTheID(t *testing.T) {
	t.Parallel()

	assertPropertyDescription(t, types.GetDocumentToolInputSchema, "id", "Not a document version id")
	assertPropertyDescription(t, types.UpdateDocumentToolInputSchema, "id", "Not a document version id")
	assertPropertyDescription(t, types.ReadDocumentToolInputSchema, "document_id", "Not a document version id")
	assertPropertyDescription(t, types.ReadDocumentToolInputSchema, "version", "PUBLISHED")
	assertPropertyDescription(t, types.ListDocumentVersionsToolInputSchema, "document_id", "Not a document version id")
	assertPropertyDescription(t, types.PublishDocumentToolInputSchema, "document_id", "Not a document version id")
	assertPropertyDescription(t, types.GetDocumentVersionToolInputSchema, "id", "Not a document id")
	assertPropertyDescription(t, types.SignDocumentToolInputSchema, "document_version_id", "Not a document id")
	assertPropertyDescription(t, types.ApproveDocumentVersionToolInputSchema, "document_version_id", "Not a document id")
	assertPropertyDescription(t, types.CancelSignatureRequestToolInputSchema, "document_version_signature_id", "Not a document id")
	assertPropertyDescription(t, types.RequestDocumentVersionSignatureToolInputSchema, "signatory_id", "Not a document id")

	filter := types.ListDocumentsToolInputSchema.Properties["filter"]
	require.NotNil(t, filter)
	assertPropertyDescription(t, filter, "document_types", "inside filter")
	assertPropertyDescription(t, filter, "published", "published version")
}

func assertPropertyDescription(t *testing.T, schema *jsonschema.Schema, property string, want string) {
	t.Helper()

	require.NotNil(t, schema)
	field := schema.Properties[property]
	require.NotNil(t, field, property)
	assert.Contains(t, field.Description, want)
}
