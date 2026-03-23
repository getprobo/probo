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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { RichEditor, useToast } from "@probo/ui";
import { useCallback } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { DocumentDescriptionPage_updateContentMutation } from "#/__generated__/core/DocumentDescriptionPage_updateContentMutation.graphql";
import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";

const autoSaveIntervalMs = 1000;

export const documentDescriptionPageQuery = graphql`
  query DocumentDescriptionPageQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId/versions/:versionId/description
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        id
        content
        status
      }
    }
    document: node(id: $documentId) {
      __typename
      ... on Document {
        # We use this on /documents/:documentId/description
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) @skip(if: $versionSpecified) {
          edges {
            node {
              id
              content
              status
            }
          }
        }
      }
    }
  }
`;

const updateContentMutation = graphql`
  mutation DocumentDescriptionPage_updateContentMutation($input: UpdateDocumentVersionContentInput!) {
    updateDocumentVersionContent(input: $input) {
      content
    }
  }
`;

export function DocumentDescriptionPage(props: { queryRef: PreloadedQuery<DocumentDescriptionPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const { document, version } = usePreloadedQuery<DocumentDescriptionPageQuery>(
    documentDescriptionPageQuery,
    queryRef,
  );
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid type for node");
  }

  const lastVersion = document.lastVersion?.edges[0].node;
  const currentVersion = lastVersion ?? version as NonNullable<typeof lastVersion | typeof version>;

  const [updateContent, _] = useMutation<DocumentDescriptionPage_updateContentMutation>(updateContentMutation);

  const handleUpdate = useDebounceCallback(
    useCallback((content: string) => {
      updateContent({
        variables: {
          input: {
            id: currentVersion.id,
            content,
          },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast({
              title: __("Error"),
              description: formatError(__("Content not saved"), errors),
              variant: "error",
            });
            return;
          }

          toast({
            title: __("Success"),
            description: __("Content saved"),
            variant: "success",
          });
        },
        onError: (error) => {
          toast({
            title: __("Error"),
            description: error.message ?? __("Content not saved"),
            variant: "error",
          });
        },
      });
    }, [currentVersion.id, updateContent, toast, __]),
    autoSaveIntervalMs,
  );

  return (
    <RichEditor
      className="flex-1"
      content={currentVersion.content}
      disabled={currentVersion.status !== "DRAFT"}
      onChangeContent={handleUpdate}
    />
  );
}
