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

import { CheckIcon, XIcon } from "@phosphor-icons/react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentSuggestionDecisions_suggestion$key } from "#/__generated__/core/DocumentSuggestionDecisions_suggestion.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import { useAcceptDocumentSuggestion } from "../../_lib/useAcceptDocumentSuggestion";
import { useRejectDocumentSuggestion } from "../../_lib/useRejectDocumentSuggestion";

import { DocumentReviewAction } from "./DocumentReviewAction";

const documentSuggestionDecisionsFragment = graphql`
  fragment DocumentSuggestionDecisions_suggestion on DocumentSuggestion {
    id
    canAccept: permission(action: "core:document:suggest-accept")
    canReject: permission(action: "core:document:suggest-reject")
  }
`;

interface DocumentSuggestionDecisionsProps {
  suggestionKey: DocumentSuggestionDecisions_suggestion$key;
  threadConnections: DocumentThreadConnections;
}

export function DocumentSuggestionDecisions({ suggestionKey, threadConnections }: DocumentSuggestionDecisionsProps) {
  const suggestion = useFragment(documentSuggestionDecisionsFragment, suggestionKey);
  const { t } = useTranslation("organizations/documents");
  const [acceptDocumentSuggestion, isAccepting] = useAcceptDocumentSuggestion();
  const [rejectDocumentSuggestion, isRejecting] = useRejectDocumentSuggestion();
  const deciding = isAccepting || isRejecting;

  if (!suggestion.canAccept && !suggestion.canReject) {
    return null;
  }

  return (
    <>
      {suggestion.canAccept && (
        <DocumentReviewAction
          label={t("details.review.accept")}
          color="green"
          loading={isAccepting}
          disabled={deciding}
          onClick={() => {
            void acceptDocumentSuggestion(suggestion.id, threadConnections).catch(() => {
              // The mutation already shows an error toast.
            });
          }}
        >
          <CheckIcon />
        </DocumentReviewAction>
      )}
      {suggestion.canReject && (
        <DocumentReviewAction
          label={t("details.review.reject")}
          color="red"
          loading={isRejecting}
          disabled={deciding}
          onClick={() => {
            void rejectDocumentSuggestion(suggestion.id, threadConnections).catch(() => {
              // The mutation already shows an error toast.
            });
          }}
        >
          <XIcon />
        </DocumentReviewAction>
      )}
    </>
  );
}
