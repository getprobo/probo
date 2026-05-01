// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
