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

package console_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/journey"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// 1x1 PNG.
const richTextPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func TestRichTextFile_SaveAndDownload(t *testing.T) {
	t.Parallel()

	world := journey.New(t)
	owner := world.NewActor("Owner", testutil.RoleOwner)
	admin := world.NewMemberActor("Admin", testutil.RoleAdmin, owner)
	viewer := world.NewMemberActor("Viewer", testutil.RoleViewer, owner)
	outsider := world.NewActor("Outsider", testutil.RoleOwner)
	guest := world.NewUnauthenticatedActor("Guest")

	png, err := base64.StdEncoding.DecodeString(richTextPNG)
	require.NoError(t, err)

	pdf := []byte("%PDF-1.1\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")

	var (
		image  uploadedRichTextFile
		report uploadedRichTextFile
	)

	owner.Step("uploads a diagram", func() error {
		uploaded, err := uploadRichTextFile(owner.Client(), "diagram.png", "image/png", png)
		if err != nil {
			return err
		}

		image = uploaded

		return expectPrivateFileURL(uploaded)
	})

	admin.Step("uploads a report", func() error {
		uploaded, err := uploadRichTextFile(admin.Client(), "policy.pdf", "application/pdf", pdf)
		if err != nil {
			return err
		}

		report = uploaded

		return expectPrivateFileURL(uploaded)
	})

	owner.Step("rejects a video upload", func() error {
		_, err := uploadRichTextFile(owner.Client(), "clip.mp4", "video/mp4", []byte("not-a-video"))
		if err == nil {
			return fmt.Errorf("video upload succeeded")
		}

		if !strings.Contains(err.Error(), "allowed") {
			return fmt.Errorf("video upload error %q", err)
		}

		return nil
	})

	viewer.Step("cannot upload a file", func() error {
		_, err := uploadRichTextFile(viewer.Client(), "diagram.png", "image/png", png)
		if err == nil {
			return fmt.Errorf("viewer upload succeeded")
		}

		testutil.RequireForbiddenError(t, err, "viewer cannot upload a rich text file")

		return nil
	})

	var (
		taskID            string
		commentID         string
		documentVersionID string
		riskAnalysisID    string
	)

	owner.Step("saves the diagram and the report on a task", func() error {
		content, err := richTextContent(
			richTextImage(image),
			richTextFile(report),
		)
		if err != nil {
			return err
		}

		var result struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID      string `json:"id"`
						Content string `json:"content"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err = owner.Client().Execute(`
			mutation($input: CreateTaskInput!) {
				createTask(input: $input) {
					taskEdge { node { id content } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.Client().GetOrganizationID().String(),
				"name":           factory.SafeName("Rich text task"),
				"content":        content,
				"priority":       "MEDIUM",
			},
		}, &result)
		if err != nil {
			return err
		}

		taskID = result.CreateTask.TaskEdge.Node.ID
		stored := result.CreateTask.TaskEdge.Node.Content

		return expectStoredFiles(stored, image, report)
	})

	owner.Step("saves the report on a comment", func() error {
		content, err := richTextContent(richTextFile(report))
		if err != nil {
			return err
		}

		var result struct {
			CreateTaskComment struct {
				TaskCommentEdge struct {
					Node struct {
						ID      string `json:"id"`
						Content string `json:"content"`
					} `json:"node"`
				} `json:"taskCommentEdge"`
			} `json:"createTaskComment"`
		}

		err = owner.Client().Execute(`
			mutation($input: CreateTaskCommentInput!) {
				createTaskComment(input: $input) {
					taskCommentEdge { node { id content } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": content,
			},
		}, &result)
		if err != nil {
			return err
		}

		commentID = result.CreateTaskComment.TaskCommentEdge.Node.ID

		return expectStoredFiles(result.CreateTaskComment.TaskCommentEdge.Node.Content, report)
	})

	owner.Step("saves the diagram on a document", func() error {
		content, err := richTextContent(richTextImage(image))
		if err != nil {
			return err
		}

		var result struct {
			CreateDocument struct {
				DocumentVersionEdge struct {
					Node struct {
						ID      string `json:"id"`
						Content string `json:"content"`
					} `json:"node"`
				} `json:"documentVersionEdge"`
			} `json:"createDocument"`
		}

		err = owner.Client().Execute(`
			mutation($input: CreateDocumentInput!) {
				createDocument(input: $input) {
					documentVersionEdge { node { id content } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.Client().GetOrganizationID().String(),
				"title":          factory.SafeName("Rich text document"),
				"content":        content,
				"documentType":   "OTHER",
				"classification": "INTERNAL",
			},
		}, &result)
		if err != nil {
			return err
		}

		documentVersionID = result.CreateDocument.DocumentVersionEdge.Node.ID

		return expectStoredFiles(result.CreateDocument.DocumentVersionEdge.Node.Content, image)
	})

	owner.Step("saves the report on a risk analysis", func() error {
		content, err := richTextContent(richTextFile(report))
		if err != nil {
			return err
		}

		var result struct {
			CreateRiskAnalysis struct {
				RiskAnalysisEdge struct {
					Node struct {
						ID          string `json:"id"`
						Description string `json:"description"`
					} `json:"node"`
				} `json:"riskAnalysisEdge"`
			} `json:"createRiskAnalysis"`
		}

		err = owner.Client().Execute(`
			mutation($input: CreateRiskAnalysisInput!) {
				createRiskAnalysis(input: $input) {
					riskAnalysisEdge { node { id description } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.Client().GetOrganizationID().String(),
				"name":           factory.SafeName("Rich text analysis"),
				"description":    content,
				"matrixSize":     map[string]any{"rows": 5, "cols": 5},
			},
		}, &result)
		if err != nil {
			return err
		}

		riskAnalysisID = result.CreateRiskAnalysis.RiskAnalysisEdge.Node.ID

		return expectStoredFiles(result.CreateRiskAnalysis.RiskAnalysisEdge.Node.Description, report)
	})

	viewer.Step("downloads files through the parent that contains them", func() error {
		gotImage, err := downloadAttachment(viewer.Client(), taskID, image.ID)
		if err != nil {
			return err
		}

		if string(gotImage) != string(png) {
			return fmt.Errorf("task diagram bytes differ")
		}

		gotReport, err := downloadAttachment(viewer.Client(), commentID, report.ID)
		if err != nil {
			return err
		}

		if string(gotReport) != string(pdf) {
			return fmt.Errorf("comment report bytes differ")
		}

		gotImage, err = downloadAttachment(viewer.Client(), documentVersionID, image.ID)
		if err != nil {
			return err
		}

		if string(gotImage) != string(png) {
			return fmt.Errorf("document diagram bytes differ")
		}

		gotReport, err = downloadAttachment(viewer.Client(), riskAnalysisID, report.ID)
		if err != nil {
			return err
		}

		if string(gotReport) != string(pdf) {
			return fmt.Errorf("risk analysis report bytes differ")
		}

		resp := viewer.Client().GetNoRedirect(attachmentURL(viewer.Client(), commentID, image.ID))
		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("comment diagram status %d", resp.StatusCode)
		}

		return nil
	})

	viewer.Step("downloads the diagram and the report", func() error {
		gotImage, err := downloadPrivateFile(viewer.Client(), image.ID)
		if err != nil {
			return err
		}

		if string(gotImage) != string(png) {
			return fmt.Errorf("diagram bytes differ")
		}

		gotReport, err := downloadPrivateFile(viewer.Client(), report.ID)
		if err != nil {
			return err
		}

		if string(gotReport) != string(pdf) {
			return fmt.Errorf("report bytes differ")
		}

		return nil
	})

	outsider.Step("cannot download the diagram through the task", func() error {
		resp := outsider.Client().GetNoRedirect(attachmentURL(outsider.Client(), taskID, image.ID))
		if resp.StatusCode == http.StatusTemporaryRedirect {
			return fmt.Errorf("outsider was redirected to the attachment")
		}

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("outsider attachment status %d", resp.StatusCode)
		}

		return nil
	})

	outsider.Step("cannot download the diagram", func() error {
		resp := outsider.Client().GetNoRedirect(privateFileURL(outsider.Client(), image.ID))
		if resp.StatusCode == http.StatusTemporaryRedirect {
			return fmt.Errorf("outsider was redirected to the file")
		}

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("outsider download status %d", resp.StatusCode)
		}

		return nil
	})

	guest.Step("cannot download the diagram through the task", func() error {
		resp := guest.Client().GetNoRedirect(attachmentURL(guest.Client(), taskID, image.ID))
		if resp.StatusCode != http.StatusUnauthorized {
			return fmt.Errorf("guest attachment status %d", resp.StatusCode)
		}

		return nil
	})

	guest.Step("cannot download the diagram", func() error {
		resp := guest.Client().GetNoRedirect(privateFileURL(guest.Client(), image.ID))
		if resp.StatusCode != http.StatusUnauthorized {
			return fmt.Errorf("guest download status %d", resp.StatusCode)
		}

		return nil
	})
}

type uploadedRichTextFile struct {
	ID          string `json:"id"`
	FileName    string `json:"fileName"`
	MimeType    string `json:"mimeType"`
	Size        int    `json:"size"`
	DownloadURL string `json:"downloadUrl"`
}

func uploadRichTextFile(client *testutil.Client, name, mimeType string, body []byte) (uploadedRichTextFile, error) {
	var result struct {
		UploadFile struct {
			File uploadedRichTextFile `json:"file"`
		} `json:"uploadFile"`
	}

	err := client.ExecuteWithFile(`
		mutation($input: UploadFileInput!) {
			uploadFile(input: $input) {
				file { id fileName mimeType size downloadUrl }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": client.GetOrganizationID().String(),
			"file":           nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    name,
		ContentType: mimeType,
		Content:     body,
	}, &result)

	return result.UploadFile.File, err
}

func expectPrivateFileURL(file uploadedRichTextFile) error {
	if file.ID == "" {
		return fmt.Errorf("uploaded file has no id")
	}

	want := "/api/files/v1/" + file.ID
	if !strings.Contains(file.DownloadURL, want) || strings.Contains(file.DownloadURL, "/attachments/") {
		return fmt.Errorf("download url %q", file.DownloadURL)
	}

	return nil
}

func richTextImage(file uploadedRichTextFile) map[string]any {
	return map[string]any{
		"type": "image",
		"attrs": map[string]any{
			"fileId": file.ID,
			"alt":    "diagram",
		},
	}
}

func richTextFile(file uploadedRichTextFile) map[string]any {
	return map[string]any{
		"type": "file",
		"attrs": map[string]any{
			"fileId":   file.ID,
			"fileName": file.FileName,
			"mimeType": file.MimeType,
			"size":     file.Size,
		},
	}
}

func richTextContent(nodes ...map[string]any) (string, error) {
	raw, err := json.Marshal(map[string]any{
		"type":    "doc",
		"content": nodes,
	})
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func expectStoredFiles(content string, files ...uploadedRichTextFile) error {
	for _, file := range files {
		if !strings.Contains(content, file.ID) {
			return fmt.Errorf("stored content does not contain file %s", file.ID)
		}

		if file.DownloadURL != "" && strings.Contains(content, file.DownloadURL) {
			return fmt.Errorf("stored content contains download url for file %s", file.ID)
		}
	}

	return nil
}

func privateFileURL(client *testutil.Client, fileID string) string {
	return client.BaseURL() + "/api/files/v1/" + fileID
}

func attachmentURL(client *testutil.Client, parentID, fileID string) string {
	return client.BaseURL() + "/api/files/v1/attachments/" + parentID + "/" + fileID
}

func downloadAttachment(client *testutil.Client, parentID, fileID string) ([]byte, error) {
	return downloadRedirect(client, attachmentURL(client, parentID, fileID))
}

func downloadPrivateFile(client *testutil.Client, fileID string) ([]byte, error) {
	return downloadRedirect(client, privateFileURL(client, fileID))
}

func downloadRedirect(client *testutil.Client, rawURL string) ([]byte, error) {
	resp := client.GetNoRedirect(rawURL)
	if resp.StatusCode != http.StatusTemporaryRedirect {
		return nil, fmt.Errorf("file status %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("file redirect has no location")
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	got, err := httpClient.Get(location)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch file: %w", err)
	}

	defer func() { _ = got.Body.Close() }()

	body, err := io.ReadAll(got.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	if got.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stored file status %d", got.StatusCode)
	}

	return body, nil
}
