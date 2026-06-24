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

package translate

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const translateMutation = `
mutation($input: UpsertCookieBannerTranslationInput!) {
  upsertCookieBannerTranslation(input: $input) {
    cookieBannerTranslation {
      id
      language
    }
  }
}
`

func NewCmdTranslate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLanguage     string
		flagTranslations string
	)

	cmd := &cobra.Command{
		Use:   "translate <id>",
		Short: "Upsert a translation for a language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagLanguage == "" {
				return fmt.Errorf("--language is required")
			}

			if flagTranslations == "" {
				return fmt.Errorf("--translations is required")
			}

			var translations json.RawMessage
			if err := json.Unmarshal([]byte(flagTranslations), &translations); err != nil {
				return fmt.Errorf("invalid JSON for --translations: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			input := map[string]any{
				"cookieBannerId": args[0],
				"language":       flagLanguage,
				"translations":   translations,
			}

			_, err = client.Do(translateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Upserted translation for language %q on cookie banner %s\n", flagLanguage, args[0])

			return nil
		},
	}

	cmd.Flags().StringVar(&flagLanguage, "language", "", "Language code (e.g. fr, de, es)")
	cmd.Flags().StringVar(&flagTranslations, "translations", "", "Translations JSON")
	_ = cmd.MarkFlagRequired("language")
	_ = cmd.MarkFlagRequired("translations")

	return cmd
}
