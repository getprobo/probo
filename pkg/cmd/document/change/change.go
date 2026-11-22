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

package change

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdChange(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change <command>",
		Short: "Edit a document or store the same change as a suggestion",
	}

	cmd.AddCommand(newTextCmd(f))
	cmd.AddCommand(newFormatCmd(f))
	cmd.AddCommand(newTableCmd(f))
	cmd.AddCommand(newBlockCmd(f))

	return cmd
}

type changeTarget struct {
	mode         string
	documentID   string
	suggestionID string
	occurrence   int
}

func (t changeTarget) input() (map[string]any, error) {
	if err := cmdutil.ValidateEnum("mode", t.mode, []string{"APPLY", "SUGGEST"}); err != nil {
		return nil, err
	}

	input := map[string]any{"mode": t.mode}

	if t.mode == "APPLY" {
		if strings.TrimSpace(t.documentID) == "" {
			return nil, fmt.Errorf("--document is required when --mode is APPLY")
		}

		input["documentId"] = t.documentID
	}

	if t.mode == "SUGGEST" {
		if strings.TrimSpace(t.suggestionID) == "" {
			return nil, fmt.Errorf("--suggestion is required when --mode is SUGGEST")
		}

		input["suggestionId"] = t.suggestionID
	}

	if t.occurrence != 0 {
		input["quoteOccurrence"] = t.occurrence
	}

	return input, nil
}

func addTargetFlags(cmd *cobra.Command, target *changeTarget) {
	cmd.Flags().StringVar(&target.mode, "mode", "", "APPLY writes the change. SUGGEST stores it on an open suggestion.")
	cmd.Flags().StringVar(&target.documentID, "document", "", "Document ID, required when --mode is APPLY")
	cmd.Flags().StringVar(&target.suggestionID, "suggestion", "", "Open suggestion ID, required when --mode is SUGGEST")
	cmd.Flags().IntVar(&target.occurrence, "occurrence", 0, "Zero-based occurrence of the quote")
	_ = cmd.MarkFlagRequired("mode")
}

func newTextCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput      *string
		target          changeTarget
		flagQuote       string
		flagReplacement string
	)

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Replace quoted text",
		Example: `  prb document change text --mode APPLY --document <id> --quote scope --replacement range
  prb document change text --mode SUGGEST --suggestion <id> --quote scope --replacement range`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(flagQuote) == "" {
				return fmt.Errorf("quote is required")
			}

			if !cmd.Flags().Changed("replacement") {
				return fmt.Errorf("replacement is required; pass an empty value to delete the quote")
			}

			input, err := target.input()
			if err != nil {
				return err
			}

			input["quote"] = flagQuote
			input["replacement"] = flagReplacement

			return runChange(f, flagOutput, textMutation, input, "Changed text")
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)
	addTargetFlags(cmd, &target)
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Text to replace")
	cmd.Flags().StringVar(&flagReplacement, "replacement", "", "New text; empty deletes the quote")

	return cmd
}

var formatMarks = []string{"BOLD", "ITALIC", "UNDERLINE", "STRIKE", "CODE", "LINK"}

var formatActions = []string{"ADD", "REMOVE"}

func newFormatCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput *string
		target     changeTarget
		flagMark   string
		flagAction string
		flagQuote  string
		flagHref   string
	)

	cmd := &cobra.Command{
		Use:   "format",
		Short: "Add or remove bold, italic, or another style",
		Example: `  prb document change format --mode APPLY --document <id> --mark BOLD --action ADD --quote scope
  prb document change format --mode SUGGEST --suggestion <id> --mark LINK --action ADD --quote scope --href https://example.com`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("mark", flagMark, formatMarks); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum("action", flagAction, formatActions); err != nil {
				return err
			}

			if strings.TrimSpace(flagQuote) == "" {
				return fmt.Errorf("quote is required")
			}

			if flagMark == "LINK" && flagAction == "ADD" && strings.TrimSpace(flagHref) == "" {
				return fmt.Errorf("--href is required when adding a link")
			}

			if flagHref != "" && flagMark != "LINK" {
				return fmt.Errorf("--href only applies to LINK")
			}

			input, err := target.input()
			if err != nil {
				return err
			}

			input["mark"] = flagMark
			input["action"] = flagAction
			input["quote"] = flagQuote

			if flagHref != "" {
				input["href"] = flagHref
			}

			return runChange(f, flagOutput, formatMutation, input, "Changed formatting")
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)
	addTargetFlags(cmd, &target)
	cmd.Flags().StringVar(&flagMark, "mark", "", "Style: "+strings.Join(formatMarks, ", "))
	cmd.Flags().StringVar(&flagAction, "action", "", "ADD or REMOVE")
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Text to format")
	cmd.Flags().StringVar(&flagHref, "href", "", "Link address when adding a link")
	_ = cmd.MarkFlagRequired("mark")
	_ = cmd.MarkFlagRequired("action")

	return cmd
}

var tableOperations = []string{
	"INSERT_ROW",
	"DELETE_ROW",
	"INSERT_COLUMN",
	"DELETE_COLUMN",
	"INSERT_TABLE",
	"DELETE_TABLE",
}

func newTableCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput    *string
		target        changeTarget
		flagOperation string
		flagQuote     string
		flagBefore    bool
		flagCells     []string
		flagRows      []string
		flagHeaderRow bool
	)

	cmd := &cobra.Command{
		Use:   "table",
		Short: "Insert or delete a table row, column, or table",
		Example: `  prb document change table --mode APPLY --document <id> --operation INSERT_ROW --quote Read --cell Write
  prb document change table --mode SUGGEST --suggestion <id> --operation INSERT_TABLE --quote Scope --header-row --row "Right|Access"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("operation", flagOperation, tableOperations); err != nil {
				return err
			}

			if strings.TrimSpace(flagQuote) == "" {
				return fmt.Errorf("quote is required")
			}

			insertsCells := flagOperation == "INSERT_ROW" || flagOperation == "INSERT_COLUMN"
			if len(flagCells) > 0 && !insertsCells {
				return fmt.Errorf("--cell only applies to INSERT_ROW and INSERT_COLUMN")
			}

			if (len(flagRows) > 0 || flagHeaderRow) && flagOperation != "INSERT_TABLE" {
				return fmt.Errorf("--row and --header-row only apply to INSERT_TABLE")
			}

			if flagOperation == "INSERT_TABLE" && len(flagRows) == 0 {
				return fmt.Errorf("--row is required with INSERT_TABLE")
			}

			if flagBefore && !insertsCells && flagOperation != "INSERT_TABLE" {
				return fmt.Errorf("--before only applies to insert operations")
			}

			input, err := target.input()
			if err != nil {
				return err
			}

			input["operation"] = flagOperation
			input["quote"] = flagQuote

			if flagBefore {
				input["before"] = true
			}

			if len(flagCells) > 0 {
				input["cells"] = flagCells
			}

			if len(flagRows) > 0 {
				rows := make([][]string, 0, len(flagRows))
				for _, row := range flagRows {
					rows = append(rows, strings.Split(row, "|"))
				}

				input["rows"] = rows
			}

			if flagHeaderRow {
				input["headerRow"] = true
			}

			return runChange(f, flagOutput, tableMutation, input, "Changed the table")
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)
	addTargetFlags(cmd, &target)
	cmd.Flags().StringVar(&flagOperation, "operation", "", "Change: "+strings.Join(tableOperations, ", "))
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Text in a cell of the target, or the block a new table goes beside")
	cmd.Flags().BoolVar(&flagBefore, "before", false, "Insert before the quoted row, column, or block")
	cmd.Flags().StringArrayVar(&flagCells, "cell", nil, "Cell text of the new row or column (can be repeated)")
	cmd.Flags().StringArrayVar(&flagRows, "row", nil, "Row of the new table, cells separated by | (can be repeated)")
	cmd.Flags().BoolVar(&flagHeaderRow, "header-row", false, "Make the new table's first row a header row")
	_ = cmd.MarkFlagRequired("operation")

	return cmd
}

var blockOperations = []string{
	"SET_BLOCK",
	"INSERT_BLOCK",
	"DELETE_BLOCK",
	"INSERT_LIST_ITEM",
	"DELETE_LIST_ITEM",
	"NEST_LIST_ITEM",
	"UNNEST_LIST_ITEM",
	"SET_TASK",
	"INSERT_DIVIDER",
	"INSERT_IMAGE",
	"INSERT_FILE",
	"INSERT_MENTION",
	"DELETE_MENTION",
}

var blockKinds = []string{
	"PARAGRAPH",
	"HEADING",
	"BLOCKQUOTE",
	"BULLET_LIST",
	"ORDERED_LIST",
	"TASK_LIST",
	"CODE_BLOCK",
}

func newBlockCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput    *string
		target        changeTarget
		flagOperation string
		flagQuote     string
		flagBefore    bool
		flagBlock     string
		flagLevel     int
		flagLanguage  string
		flagText      string
		flagItems     []string
		flagChecked   bool
		flagFileID    string
		flagSrc       string
		flagAlt       string
		flagWidth     int
		flagProfileID string
	)

	cmd := &cobra.Command{
		Use:   "block",
		Short: "Change a heading, list, quote, code block, divider, image, file, or mention",
		Example: `  prb document change block --mode APPLY --document <id> --operation SET_BLOCK --block HEADING --level 2 --quote Scope
  prb document change block --mode SUGGEST --suggestion <id> --operation INSERT_IMAGE --file <file-id> --quote Scope`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("operation", flagOperation, blockOperations); err != nil {
				return err
			}

			if flagBlock != "" {
				if err := cmdutil.ValidateEnum("block", flagBlock, blockKinds); err != nil {
					return err
				}
			}

			input, err := target.input()
			if err != nil {
				return err
			}

			input["operation"] = flagOperation

			if flagQuote != "" {
				input["quote"] = flagQuote
			}

			if cmd.Flags().Changed("before") {
				input["before"] = flagBefore
			}

			if flagBlock != "" {
				input["block"] = flagBlock
			}

			if cmd.Flags().Changed("level") {
				input["level"] = flagLevel
			}

			if flagLanguage != "" {
				input["language"] = flagLanguage
			}

			if cmd.Flags().Changed("text") {
				input["text"] = flagText
			}

			if len(flagItems) > 0 {
				input["items"] = flagItems
			}

			if cmd.Flags().Changed("checked") {
				input["checked"] = flagChecked
			}

			if flagFileID != "" {
				input["fileId"] = flagFileID
			}

			if flagSrc != "" {
				input["src"] = flagSrc
			}

			if flagAlt != "" {
				input["alt"] = flagAlt
			}

			if cmd.Flags().Changed("width") {
				input["width"] = flagWidth
			}

			if flagProfileID != "" {
				input["profileId"] = flagProfileID
			}

			return runChange(f, flagOutput, blockMutation, input, "Changed the block")
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)
	addTargetFlags(cmd, &target)
	cmd.Flags().StringVar(&flagOperation, "operation", "", "Change: "+strings.Join(blockOperations, ", "))
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Text that locates the block. Empty inserts into an empty document.")
	cmd.Flags().BoolVar(&flagBefore, "before", false, "Insert before the quote, or delete the neighboring textless block")
	cmd.Flags().StringVar(&flagBlock, "block", "", "Block kind: "+strings.Join(blockKinds, ", "))
	cmd.Flags().IntVar(&flagLevel, "level", 0, "Heading level from 1 to 6")
	cmd.Flags().StringVar(&flagLanguage, "language", "", "Code block language. mermaid draws a diagram.")
	cmd.Flags().StringVar(&flagText, "text", "", "Text of the new block or list item")
	cmd.Flags().StringArrayVar(&flagItems, "item", nil, "List item text (can be repeated)")
	cmd.Flags().BoolVar(&flagChecked, "checked", false, "Checked state for SET_TASK")
	cmd.Flags().StringVar(&flagFileID, "file", "", "Uploaded file ID for an image or file")
	cmd.Flags().StringVar(&flagSrc, "src", "", "http or https image address")
	cmd.Flags().StringVar(&flagAlt, "alt", "", "Image alternative text")
	cmd.Flags().IntVar(&flagWidth, "width", 0, "Image width as a percentage from 1 to 100")
	cmd.Flags().StringVar(&flagProfileID, "profile", "", "Membership profile to mention")
	_ = cmd.MarkFlagRequired("operation")

	return cmd
}

func runChange(
	f *cmdutil.Factory,
	flagOutput *string,
	query string,
	input map[string]any,
	past string,
) error {
	if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
		return err
	}

	client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
	if err != nil {
		return err
	}

	data, err := client.Do(query, map[string]any{"input": input})
	if err != nil {
		return err
	}

	if *flagOutput == cmdutil.OutputJSON {
		return cmdutil.PrintJSON(f.IOStreams.Out, json.RawMessage(data))
	}

	var resp map[string]struct {
		Document struct {
			ID string `json:"id"`
		} `json:"document"`
		DocumentSuggestion *struct {
			ID string `json:"id"`
		} `json:"documentSuggestion"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("cannot parse response: %w", err)
	}

	for _, payload := range resp {
		if payload.DocumentSuggestion != nil {
			_, _ = fmt.Fprintf(f.IOStreams.Out, "%s on suggestion %s\n", past, payload.DocumentSuggestion.ID)

			return nil
		}

		_, _ = fmt.Fprintf(f.IOStreams.Out, "%s in document %s\n", past, payload.Document.ID)

		return nil
	}

	return fmt.Errorf("cannot parse response")
}

const selection = `
    document { id }
    documentSuggestion { id status }
`

const textMutation = `
mutation ChangeDocumentText($input: ChangeDocumentTextInput!) {
  changeDocumentText(input: $input) {` + selection + `
  }
}
`

const formatMutation = `
mutation ChangeDocumentFormat($input: ChangeDocumentFormatInput!) {
  changeDocumentFormat(input: $input) {` + selection + `
  }
}
`

const tableMutation = `
mutation ChangeDocumentTable($input: ChangeDocumentTableInput!) {
  changeDocumentTable(input: $input) {` + selection + `
  }
}
`

const blockMutation = `
mutation ChangeDocumentBlock($input: ChangeDocumentBlockInput!) {
  changeDocumentBlock(input: $input) {` + selection + `
  }
}
`
