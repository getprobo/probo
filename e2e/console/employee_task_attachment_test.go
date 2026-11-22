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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/journey"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestEmployeeTaskAttachment_SaveAndDownload(t *testing.T) {
	t.Parallel()

	world := journey.New(t)
	owner := world.NewActor("Owner", testutil.RoleOwner)
	employee := world.NewMemberActor("Employee", testutil.RoleEmployee, owner)
	coworker := world.NewMemberActor("Coworker", testutil.RoleEmployee, owner)
	outsider := world.NewActor("Outsider", testutil.RoleOwner)

	png, err := base64.StdEncoding.DecodeString(richTextPNG)
	require.NoError(t, err)

	var (
		taskID       string
		otherTaskID  string
		evidence     uploadedRichTextFile
		template     uploadedRichTextFile
		confidential uploadedRichTextFile
		reply        uploadedRichTextFile
	)

	owner.Step("assigns tasks", func() error {
		taskID = factory.NewTaskWithoutInternalControl(owner.Client()).
			WithName(factory.SafeName("Employee evidence")).
			Create()
		otherTaskID = factory.NewTaskWithoutInternalControl(owner.Client()).
			WithName(factory.SafeName("Coworker's task")).
			Create()
		assignTask(t, owner.Client(), taskID, employee.Client().GetProfileID().String())
		assignTask(t, owner.Client(), otherTaskID, coworker.Client().GetProfileID().String())

		return nil
	})

	owner.Step("attaches a template and keeps an unrelated file", func() error {
		var uploadErr error

		template, uploadErr = uploadRichTextFile(owner.Client(), "template.png", "image/png", png)
		if uploadErr != nil {
			return uploadErr
		}

		confidential, uploadErr = uploadRichTextFile(owner.Client(), "confidential.png", "image/png", png)
		if uploadErr != nil {
			return uploadErr
		}

		content, contentErr := richTextContent(richTextFile(template))
		if contentErr != nil {
			return contentErr
		}

		return setTaskContent(owner.Client(), taskID, content)
	})

	employee.Step("downloads the template attached to their task", func() error {
		_, downloadErr := downloadPrivateFile(employee.Client(), template.ID)

		return downloadErr
	})

	employee.Step("downloads an organization file that is not on their task", func() error {
		_, downloadErr := downloadPrivateFile(employee.Client(), confidential.ID)

		return downloadErr
	})

	employee.Step("uploads evidence", func() error {
		uploaded, uploadErr := uploadRichTextFile(
			employee.Client(),
			"evidence.png",
			"image/png",
			png,
		)
		if uploadErr != nil {
			return uploadErr
		}

		evidence = uploaded

		return expectPrivateFileURL(uploaded)
	})

	employee.Step("previews the unsaved evidence", func() error {
		_, downloadErr := downloadPrivateFile(employee.Client(), evidence.ID)

		return downloadErr
	})

	coworker.Step("downloads another employee's evidence", func() error {
		_, downloadErr := downloadPrivateFile(coworker.Client(), evidence.ID)

		return downloadErr
	})

	employee.Step("saves the evidence on the task", func() error {
		content, contentErr := richTextContent(richTextFile(template), richTextImage(evidence))
		if contentErr != nil {
			return contentErr
		}

		var result struct {
			UpdateTask struct {
				Task struct {
					Content string `json:"content"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		updateErr := employee.Client().Execute(`
			mutation($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { content }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": content,
			},
		}, &result)
		if updateErr != nil {
			return updateErr
		}

		return expectStoredFiles(result.UpdateTask.Task.Content, evidence)
	})

	employee.Step("saves the evidence on a comment", func() error {
		content, contentErr := richTextContent(richTextFile(evidence))
		if contentErr != nil {
			return contentErr
		}

		var result struct {
			CreateTaskComment struct {
				TaskCommentEdge struct {
					Node struct {
						Content string `json:"content"`
					} `json:"node"`
				} `json:"taskCommentEdge"`
			} `json:"createTaskComment"`
		}

		commentErr := employee.Client().Execute(`
			mutation($input: CreateTaskCommentInput!) {
				createTaskComment(input: $input) {
					taskCommentEdge { node { content } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": content,
			},
		}, &result)
		if commentErr != nil {
			return commentErr
		}

		return expectStoredFiles(
			result.CreateTaskComment.TaskCommentEdge.Node.Content,
			evidence,
		)
	})

	owner.Step("replies with a file in a comment", func() error {
		var uploadErr error

		reply, uploadErr = uploadRichTextFile(owner.Client(), "reply.png", "image/png", png)
		if uploadErr != nil {
			return uploadErr
		}

		content, contentErr := richTextContent(richTextFile(reply))
		if contentErr != nil {
			return contentErr
		}

		_, commentErr := owner.Client().Do(`
			mutation($input: CreateTaskCommentInput!) {
				createTaskComment(input: $input) {
					taskCommentEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": content,
			},
		})

		return commentErr
	})

	employee.Step("downloads the file from the comment", func() error {
		_, downloadErr := downloadPrivateFile(employee.Client(), reply.ID)

		return downloadErr
	})

	coworker.Step("downloads the file from the comment", func() error {
		_, downloadErr := downloadPrivateFile(coworker.Client(), reply.ID)

		return downloadErr
	})

	employee.Step("downloads the evidence", func() error {
		got, downloadErr := downloadPrivateFile(employee.Client(), evidence.ID)
		if downloadErr != nil {
			return downloadErr
		}

		if string(got) != string(png) {
			return fmt.Errorf("evidence bytes differ")
		}

		return nil
	})

	employee.Step("cannot rename the task", func() error {
		_, renameErr := employee.Client().Do(`
			mutation($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"name":   factory.SafeName("Renamed by employee"),
			},
		})
		if renameErr == nil {
			return fmt.Errorf("employee renamed the task")
		}

		testutil.RequireForbiddenError(t, renameErr, "employee should not rename a task")

		return nil
	})

	employee.Step("cannot attach evidence to a coworker's task", func() error {
		content, contentErr := richTextContent(richTextImage(evidence))
		if contentErr != nil {
			return contentErr
		}

		_, updateErr := employee.Client().Do(`
			mutation($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":  otherTaskID,
				"content": content,
			},
		})
		if updateErr == nil {
			return fmt.Errorf("employee updated a coworker's task")
		}

		testutil.RequireForbiddenError(t, updateErr, "employee should not update a coworker's task")

		return nil
	})

	employee.Step("cannot upload evidence to another organization", func() error {
		uploadErr := employee.Client().ExecuteWithFile(`
			mutation($input: UploadFileInput!) {
				uploadFile(input: $input) {
					file { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": outsider.Client().GetOrganizationID().String(),
				"file":           nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "evidence.png",
			ContentType: "image/png",
			Content:     png,
		}, &struct{}{})
		if uploadErr == nil {
			return fmt.Errorf("cross-organization upload succeeded")
		}

		testutil.RequireForbiddenError(t, uploadErr, "employee should not upload to another organization")

		return nil
	})

	outsider.Step("cannot download the evidence", func() error {
		return expectFileDenied(outsider.Client(), evidence.ID)
	})
}

func setTaskContent(client *testutil.Client, taskID, content string) error {
	_, err := client.Do(`
		mutation($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"taskId":  taskID,
			"content": content,
		},
	})

	return err
}

func expectFileDenied(client *testutil.Client, fileID string) error {
	resp := client.GetNoRedirect(privateFileURL(client, fileID))
	if resp.StatusCode == http.StatusTemporaryRedirect {
		return fmt.Errorf("redirected to file %s", fileID)
	}

	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("file %s status %d", fileID, resp.StatusCode)
	}

	return nil
}
