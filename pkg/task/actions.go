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

package task

const (
	ActionTaskGet      = "core:task:get"
	ActionTaskList     = "core:task:list"
	ActionTaskCreate   = "core:task:create"
	ActionTaskUpdate   = "core:task:update"
	ActionTaskDelete   = "core:task:delete"
	ActionTaskAssign   = "core:task:assign"
	ActionTaskUnassign = "core:task:unassign"

	ActionTaskCommentGet    = "core:task-comment:get"
	ActionTaskCommentList   = "core:task-comment:list"
	ActionTaskCommentCreate = "core:task-comment:create"
	ActionTaskCommentUpdate = "core:task-comment:update"
	ActionTaskCommentDelete = "core:task-comment:delete"

	ActionTaskActivityGet  = "core:task-activity:get"
	ActionTaskActivityList = "core:task-activity:list"
)
