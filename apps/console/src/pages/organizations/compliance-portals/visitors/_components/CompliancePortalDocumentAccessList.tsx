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

import type { CompliancePortalDocumentAccessStatus } from "@probo/coredata";
import type { CompliancePortalDocumentAccessInfo } from "@probo/helpers";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDocumentAccessList_access$key } from "#/__generated__/core/CompliancePortalDocumentAccessList_access.graphql";
import type { CompliancePortalDocumentAccessListRefetchQuery } from "#/__generated__/core/CompliancePortalDocumentAccessListRefetchQuery.graphql";

import {
  documentAccessInfoFromResource,
  documentAccessKey,
  updateAccessInput,
} from "../_lib/documentAccessInfo";
import {
  documentAccessListGraphqlFilter,
  useDocumentAccessListFilters,
} from "../_lib/useDocumentAccessListFilters";
import { useUpdateCompliancePortalAccess } from "../_lib/useUpdateCompliancePortalAccess";
import { documentAccessList } from "../variants";

import { CompliancePortalDocumentAccessListFilter } from "./CompliancePortalDocumentAccessListFilter";
import { CompliancePortalDocumentAccessListItem } from "./CompliancePortalDocumentAccessListItem";
import { CompliancePortalDocumentAccessSelectionBar } from "./CompliancePortalDocumentAccessSelectionBar";

const accessFragment = graphql`
  fragment CompliancePortalDocumentAccessList_access on CompliancePortalAccess
  @argumentDefinitions(
    filter: { type: "CompliancePortalAccessResourceFilter", defaultValue: null }
  )
  @refetchable(queryName: "CompliancePortalDocumentAccessListRefetchQuery") {
    resources(first: 100, filter: $filter) {
      edges {
        node {
          kind
          resourceId
          name
          category
          status
        }
      }
    }
  }
`;

interface CompliancePortalDocumentAccessListProps {
  accessKey: CompliancePortalDocumentAccessList_access$key;
  accessId: string;
  canUpdate: boolean;
}

export function CompliancePortalDocumentAccessList({
  accessKey,
  accessId,
  canUpdate,
}: CompliancePortalDocumentAccessListProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { status, hasActiveFilters } = useDocumentAccessListFilters();
  const [isPending, startTransition] = useTransition();
  const { root, heading, results } = documentAccessList({ pending: isPending });
  const [access, refetch] = useRefetchableFragment<
    CompliancePortalDocumentAccessListRefetchQuery,
    CompliancePortalDocumentAccessList_access$key
  >(accessFragment, accessKey);
  const skipFirstRefetch = useRef(true);
  const [updateAccess, isUpdating] = useUpdateCompliancePortalAccess();
  const [statusOverlay, setStatusOverlay] = useState(
    () => new Map<string, CompliancePortalDocumentAccessStatus>(),
  );
  const [selectedIds, setSelectedIds] = useState<Set<string>>(() => new Set());
  const documentAccesses = access.resources.edges.map((edge) => {
    const item = documentAccessInfoFromResource(edge.node);
    const overlayStatus = statusOverlay.get(documentAccessKey(item));
    return overlayStatus === undefined ? item : { ...item, status: overlayStatus };
  });
  const selectedItems = documentAccesses.filter(item => selectedIds.has(documentAccessKey(item)));

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    setStatusOverlay(new Map());
    setSelectedIds(new Set());
    startTransition(() => {
      refetch(
        { filter: documentAccessListGraphqlFilter(status) },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [status, refetch]);

  async function commit(updates: CompliancePortalDocumentAccessInfo[]) {
    await updateAccess({
      variables: {
        input: updateAccessInput(accessId, updates),
      },
    });
    setStatusOverlay((current) => {
      const next = new Map(current);
      for (const update of updates) {
        if (update.status != null) {
          next.set(documentAccessKey(update), update.status);
        }
      }
      return next;
    });
  }

  async function commitSelection(updates: CompliancePortalDocumentAccessInfo[]) {
    await commit(updates);
    setSelectedIds(new Set());
  }

  function toggle(key: string) {
    setSelectedIds((current) => {
      const next = new Set(current);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }

  return (
    <section className={root()}>
      <div className={heading()}>
        <Heading level={3} size={3} weight="medium" highContrast>
          {t("documentAccessList.title")}
        </Heading>
        <CompliancePortalDocumentAccessListFilter />
      </div>
      <div
        aria-busy={isPending}
        className={results()}
      >
        <List>
          {documentAccesses.length === 0
            ? (
                <ListItem>
                  <ListItemContent>
                    <Text size={2} color="faint">
                      {hasActiveFilters
                        ? t("documentAccessList.emptyFilter")
                        : t("documentAccessList.empty")}
                    </Text>
                  </ListItemContent>
                </ListItem>
              )
            : documentAccesses.map((documentAccess) => {
                const key = documentAccessKey(documentAccess);
                return (
                  <CompliancePortalDocumentAccessListItem
                    key={key}
                    documentAccess={documentAccess}
                    canUpdate={canUpdate}
                    disabled={isUpdating}
                    selected={canUpdate ? selectedIds.has(key) : undefined}
                    onSelectedChange={canUpdate ? () => toggle(key) : undefined}
                    onGrant={item => void commit([{ ...item, status: "GRANTED" }])}
                    onRejectOrRevoke={item => void commit([item])}
                  />
                );
              })}
        </List>
      </div>
      {canUpdate && (
        <CompliancePortalDocumentAccessSelectionBar
          selectedItems={selectedItems}
          allSelected={selectedItems.length === documentAccesses.length && documentAccesses.length > 0}
          loading={isUpdating}
          onClear={() => setSelectedIds(new Set())}
          onGrant={commitSelection}
          onRejectOrRevoke={commitSelection}
          onSelectAll={() => setSelectedIds(new Set(documentAccesses.map(documentAccessKey)))}
        />
      )}
    </section>
  );
}
