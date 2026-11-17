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

import { MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { DocumentSignatoryMenuQuery } from "#/__generated__/core/DocumentSignatoryMenuQuery.graphql";
import type { DocumentSignatoryMenu_requestMutation } from "#/__generated__/core/DocumentSignatoryMenu_requestMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { useComboboxPage } from "#/pages/organizations/tasks/_lib/useComboboxPage";

const pageSize = 20;
const skeletonWidths = ["w-28", "w-36", "w-24"] as const;

const documentSignatoryMenuQuery = graphql`
  query DocumentSignatoryMenuQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        profiles(
          first: $first
          after: $after
          orderBy: { direction: ASC, field: FULL_NAME }
          filter: { states: [ACTIVE], contractEnded: false, query: $query }
        ) {
          edges {
            node {
              id
              fullName
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    }
  }
`;

const requestSignatureMutation = graphql`
  mutation DocumentSignatoryMenu_requestMutation(
    $input: RequestSignatureInput!
    $connections: [ID!]!
  ) {
    requestSignature(input: $input) {
      documentVersionSignatureEdge @prependEdge(connections: $connections) {
        node {
          id
          state
          canCancel: permission(action: "core:document-version-signature:cancel")
          signedBy {
            id
            fullName
          }
        }
      }
    }
  }
`;

interface SignatoryOption {
  id: string;
  fullName: string;
}

export interface DocumentSignatoryMenuProps {
  documentVersionId: string;
  connectionId: string;
  excludeIds: ReadonlySet<string>;
  disabled?: boolean;
}

export function DocumentSignatoryMenu({
  documentVersionId,
  connectionId,
  excludeIds,
  disabled,
}: DocumentSignatoryMenuProps) {
  const { t } = useTranslation("organizations/documents");
  const organizationId = useOrganizationId();
  const [open, setOpen] = useState(false);
  const [requestSignature, isRequesting] = useMutation<DocumentSignatoryMenu_requestMutation>(
    requestSignatureMutation,
    { errorToast: t("details.errors.update") },
  );
  const page = useComboboxPage<DocumentSignatoryMenuQuery, SignatoryOption>({
    query: documentSignatoryMenuQuery,
    enabled: open,
    variables: (after, query) => ({
      organizationId,
      first: pageSize,
      after,
      query: query.trim() || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const profiles = organization?.profiles;
      return {
        items: profiles?.edges.map(edge => ({
          id: edge.node.id,
          fullName: edge.node.fullName,
        })) ?? [],
        hasNextPage: profiles?.pageInfo.hasNextPage ?? false,
        endCursor: profiles?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: organizationId,
  });
  const items = page.items.filter(item => !excludeIds.has(item.id));
  const busy = disabled || isRequesting;

  return (
    <Dropdown
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) {
          page.onQueryChange("");
        }
      }}
    >
      <DropdownTrigger
        disabled={busy}
        render={(
          <IconButton
            size={1}
            variant="ghost"
            color="neutral"
            disabled={busy}
            aria-label={t("details.signatories.add")}
          />
        )}
      >
        <PlusIcon />
      </DropdownTrigger>
      <DropdownPopup align="end">
        <div className="flex w-64 flex-col gap-1">
          <div
            className="px-1 pb-1"
            onKeyDown={(event) => {
              event.stopPropagation();
            }}
          >
            <TextField
              size={2}
              icon={<MagnifyingGlassIcon />}
              value={page.query}
              disabled={busy}
              onValueChange={page.onQueryChange}
              placeholder={t("details.signatories.search")}
              aria-label={t("details.signatories.search")}
            />
          </div>
          <div className="flex max-h-64 flex-col gap-1 overflow-y-auto">
            {items.map(item => (
              <DropdownItem
                key={item.id}
                disabled={busy}
                onClick={() => {
                  void requestSignature({
                    variables: {
                      input: {
                        documentVersionId,
                        signatoryId: item.id,
                      },
                      connections: [connectionId],
                    },
                  }).catch(() => {
                    // Error toast is already shown by useMutation.
                  });
                }}
              >
                {item.fullName}
              </DropdownItem>
            ))}
            {page.loading && items.length === 0 && skeletonWidths.map(width => (
              <div key={width} className="flex h-8 items-center px-3" aria-hidden>
                <TextSkeleton size={2} className={width} />
              </div>
            ))}
            {!page.loading && page.failed && (
              <Text size={2} color="faint" className="px-3 py-1.5">
                {t("details.signatories.loadError")}
              </Text>
            )}
            {!page.loading && !page.failed && items.length === 0 && (
              <Text size={2} color="faint" className="px-3 py-1.5">
                {t("details.signatories.empty")}
              </Text>
            )}
            {page.hasNext && (
              <Button
                variant="ghost"
                color="neutral"
                size={2}
                className="w-full"
                loading={page.loading}
                disabled={busy}
                onClick={page.onLoadMore}
              >
                {t("details.signatories.loadMore")}
              </Button>
            )}
          </div>
        </div>
      </DropdownPopup>
    </Dropdown>
  );
}
