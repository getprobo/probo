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

package prosemirror

import "fmt"

type taskListNode struct{}

// TaskList is a checklist of task items.
func TaskList() NodeSpec {
	return taskListNode{}
}

func (taskListNode) Name() NodeType {
	return NodeTaskList
}

func (taskListNode) Group() Group {
	return GroupBlock
}

func (taskListNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeTaskItem}, Min: 1}
}

func (taskListNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

type taskItemNode struct{}

// TaskItemAttrs contains attributes for task item nodes.
type TaskItemAttrs struct {
	Checked bool `json:"checked"`
}

// TaskItemAttrs parses and returns the task item attributes.
// A missing checked attribute means the item is unchecked.
func (n Node) TaskItemAttrs() (TaskItemAttrs, error) {
	return decodeAttrs(n.Attrs, TaskItemAttrs{}, "task item")
}

// TaskItem is one entry of a task list. It starts with a paragraph and may
// hold further block content.
func TaskItem() NodeSpec {
	return taskItemNode{}
}

func (taskItemNode) Name() NodeType {
	return NodeTaskItem
}

func (taskItemNode) Group() Group {
	return GroupNone
}

func (taskItemNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, First: NodeParagraph}
}

func (taskItemNode) Check(n *Node, _ CheckMode) error {
	if _, err := n.TaskItemAttrs(); err != nil {
		return fmt.Errorf("cannot validate task item node: %w", err)
	}

	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}
