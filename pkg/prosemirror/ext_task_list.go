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

func (taskListNode) RenderHTML(w *HTMLWriter, n Node) error {
	w.WriteString("<ul")
	w.WriteAttr("data-type", "taskList")
	w.WriteByte('>')

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</ul>")

	return nil
}

func (taskListNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	tight, err := listTightness(n, NodeTaskItem)
	if err != nil {
		return fmt.Errorf("cannot render task list: %w", err)
	}

	return w.renderListItems(n.Content, tight, func(_ int, item Node) (string, string, error) {
		attrs, err := item.TaskItemAttrs()
		if err != nil {
			return "", "", fmt.Errorf("cannot render task item: %w", err)
		}

		if attrs.Checked {
			return "- [x] ", "  ", nil
		}

		return "- [ ] ", "  ", nil
	})
}

type taskItemNode struct{}

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

func (taskItemNode) RenderHTML(w *HTMLWriter, n Node) error {
	attrs, err := n.TaskItemAttrs()
	if err != nil {
		return fmt.Errorf("cannot render task item node: %w", err)
	}

	w.WriteString("<li")
	w.WriteAttr("data-type", "taskItem")

	if attrs.Checked {
		w.WriteAttr("data-checked", "true")
	} else {
		w.WriteAttr("data-checked", "false")
	}

	w.WriteByte('>')
	w.WriteString("<input")

	if attrs.Checked {
		w.WriteAttr("checked", "checked")
	}

	w.WriteAttr("disabled", "disabled")
	w.WriteAttr("type", "checkbox")
	w.WriteByte('>')

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</li>")

	return nil
}

func (taskItemNode) RenderMarkdown(_ *MarkdownWriter, _ Node) error {
	return fmt.Errorf("cannot render list item outside of list context")
}
