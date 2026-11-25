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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { SuggestionChange } from "../../_lib/reviewAnchors";
import { documentSuggestionListItem } from "../../variants";

import { DocumentSuggestionChange } from "./DocumentSuggestionChange";

interface DocumentSuggestionChangeSummaryProps {
  changes: readonly SuggestionChange[];
  // Every change of the suggestion, including those left out of `changes`.
  changeCount: number;
}

// The first changes of a suggestion, then how many more there are. Open
// suggestions read them from the text, decided ones from what was recorded.
export function DocumentSuggestionChangeSummary({ changes, changeCount }: DocumentSuggestionChangeSummaryProps) {
  const { t } = useTranslation("organizations/documents");
  const { changes: changesClass } = documentSuggestionListItem();
  const hidden = changeCount - changes.length;

  if (changes.length === 0 && hidden <= 0) {
    return null;
  }

  return (
    <div className={changesClass()}>
      {changes.map((change, index) => (
        // The changes have no stable identity beyond their order.
        <DocumentSuggestionChange key={index} change={change} />
      ))}
      {hidden > 0 && (
        <Text size={1} color="faint">
          {t("details.review.changes.moreChanges", { count: hidden })}
        </Text>
      )}
    </div>
  );
}
