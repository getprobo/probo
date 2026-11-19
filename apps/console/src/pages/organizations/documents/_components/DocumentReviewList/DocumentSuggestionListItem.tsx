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
import { graphql, useFragment } from "react-relay";

import type { DocumentSuggestionListItem_suggestion$key } from "#/__generated__/core/DocumentSuggestionListItem_suggestion.graphql";

import type { SuggestionAnchor } from "../../_lib/reviewAnchors";
import { documentSuggestionListItem } from "../../variants";

import { DocumentReviewItem } from "./DocumentReviewItem";

const documentSuggestionListItemFragment = graphql`
  fragment DocumentSuggestionListItem_suggestion on DocumentSuggestion {
    createdAt
    author {
      ...DocumentReviewItem_author
    }
  }
`;

interface DocumentSuggestionListItemProps {
  suggestionKey: DocumentSuggestionListItem_suggestion$key;
  anchor: SuggestionAnchor;
  active: boolean;
  onActivate: () => void;
}

export function DocumentSuggestionListItem({
  suggestionKey,
  anchor,
  active,
  onActivate,
}: DocumentSuggestionListItemProps) {
  const suggestion = useFragment(documentSuggestionListItemFragment, suggestionKey);
  const { t } = useTranslation("organizations/documents");
  const { changes, text, deleted } = documentSuggestionListItem();

  return (
    <DocumentReviewItem
      authorKey={suggestion.author}
      createdAt={suggestion.createdAt}
      label={t("details.review.openSuggestion")}
      active={active}
      onActivate={onActivate}
    >
      <div className={changes()}>
        {anchor.deleted !== "" && (
          <Text size={2} color="red" className={deleted()}>
            {anchor.deleted}
          </Text>
        )}
        {anchor.inserted !== "" && (
          <Text size={2} color="green" className={text()}>
            {anchor.inserted}
          </Text>
        )}
        {anchor.formatting && (
          <Text size={1} color="faint">
            {t("details.review.formatting")}
          </Text>
        )}
        {anchor.structure && (
          <Text size={1} color="faint">
            {t("details.review.structure")}
          </Text>
        )}
      </div>
    </DocumentReviewItem>
  );
}
