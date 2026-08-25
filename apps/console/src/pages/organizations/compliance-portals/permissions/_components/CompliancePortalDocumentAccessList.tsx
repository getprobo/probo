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

import type { CompliancePortalDocumentAccessInfo } from "@probo/helpers";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDocumentAccessList_access$key } from "#/__generated__/core/CompliancePortalDocumentAccessList_access.graphql";

import {
  documentAccessInfoFrom,
  documentAccessKey,
  updateAccessInput,
} from "../_lib/documentAccessInfo";
import { useUpdateCompliancePortalAccess } from "../_lib/useUpdateCompliancePortalAccess";
import { documentAccessList } from "../variants";

import { CompliancePortalDocumentAccessListItem } from "./CompliancePortalDocumentAccessListItem";
import { CompliancePortalDocumentAccessSelectionBar } from "./CompliancePortalDocumentAccessSelectionBar";

const fragment = graphql`
  fragment CompliancePortalDocumentAccessList_access on CompliancePortalAccess {
    id
    availableDocumentAccesses(
      first: 100
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          ...documentAccessInfo_documentAccess
        }
      }
    }
  }
`;

interface CompliancePortalDocumentAccessListProps {
  accessKey: CompliancePortalDocumentAccessList_access$key;
  canUpdate: boolean;
}

export function CompliancePortalDocumentAccessList({
  accessKey,
  canUpdate,
}: CompliancePortalDocumentAccessListProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, heading } = documentAccessList();
  const access = useFragment(fragment, accessKey);
  const documentAccesses = access.availableDocumentAccesses.edges.map(edge =>
    documentAccessInfoFrom(edge.node, t),
  );
  const [updateAccess, isUpdating] = useUpdateCompliancePortalAccess();
  const [selectedIds, setSelectedIds] = useState<Set<string>>(() => new Set());
  const selectedItems = documentAccesses.filter(item => selectedIds.has(documentAccessKey(item)));

  async function commit(updates: CompliancePortalDocumentAccessInfo[]) {
    await updateAccess({
      variables: { input: updateAccessInput(access.id, updates) },
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
      </div>
      <List>
        {documentAccesses.length === 0
          ? (
              <ListItem>
                <ListItemContent>
                  <Text size={2} color="faint">{t("documentAccessList.empty")}</Text>
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
      {canUpdate && (
        <CompliancePortalDocumentAccessSelectionBar
          selectedItems={selectedItems}
          allSelected={selectedItems.length === documentAccesses.length && documentAccesses.length > 0}
          loading={isUpdating}
          onClear={() => setSelectedIds(new Set())}
          onSelectAll={() => setSelectedIds(new Set(documentAccesses.map(documentAccessKey)))}
          onGrant={commitSelection}
          onRejectOrRevoke={commitSelection}
        />
      )}
    </section>
  );
}
