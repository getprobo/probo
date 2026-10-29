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

import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { ConfirmDialog } from "@probo/ui/src/v2/Dialog/ConfirmDialog";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeePortalDocumentListItem_addMutation } from "#/__generated__/core/EmployeePortalDocumentListItem_addMutation.graphql";
import type { EmployeePortalDocumentListItem_document$key } from "#/__generated__/core/EmployeePortalDocumentListItem_document.graphql";
import type { EmployeePortalDocumentListItem_employeePortal$key } from "#/__generated__/core/EmployeePortalDocumentListItem_employeePortal.graphql";
import type { EmployeePortalDocumentListItem_removeMutation } from "#/__generated__/core/EmployeePortalDocumentListItem_removeMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const NS = "organizations/employee-portal";

const employeePortalFragment = graphql`
  fragment EmployeePortalDocumentListItem_employeePortal on EmployeePortal {
    id
    canUpdate: permission(action: "employee-portal:portal:update")
  }
`;

const documentFragment = graphql`
  fragment EmployeePortalDocumentListItem_document on Document {
    id
    publishedVersion: versions(
      first: 1
      orderBy: { field: CREATED_AT, direction: DESC }
      filter: { statuses: [PUBLISHED] }
    ) {
      edges {
        node {
          title
          documentType
        }
      }
    }
    employeePortalDocument {
      id
    }
  }
`;

const addDocumentMutation = graphql`
  mutation EmployeePortalDocumentListItem_addMutation(
    $input: AddEmployeePortalDocumentInput!
  ) {
    addEmployeePortalDocument(input: $input) {
      document {
        id
        employeePortalDocument {
          id
        }
      }
    }
  }
`;

const removeDocumentMutation = graphql`
  mutation EmployeePortalDocumentListItem_removeMutation(
    $input: DeleteEmployeePortalDocumentInput!
  ) {
    deleteEmployeePortalDocument(input: $input) {
      deletedEmployeePortalDocumentId @deleteRecord
      document {
        id
        employeePortalDocument {
          id
        }
      }
    }
  }
`;

export function EmployeePortalDocumentListItem(props: {
  documentKey: EmployeePortalDocumentListItem_document$key;
  employeePortalKey: EmployeePortalDocumentListItem_employeePortal$key;
}) {
  const { t } = useTranslation(NS);
  const employeePortal = useFragment(employeePortalFragment, props.employeePortalKey);
  const document = useFragment(documentFragment, props.documentKey);
  const published = document.publishedVersion.edges[0]?.node;
  const link = document.employeePortalDocument;
  const serverLinked = link != null;
  const [pendingLinked, setPendingLinked] = useState<boolean | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const isLinked = pendingLinked ?? serverLinked;
  const title = published?.title ?? "";

  const [addDocument, isAdding] = useMutation<EmployeePortalDocumentListItem_addMutation>(
    addDocumentMutation,
    {
      successMessage: t("documentsPage.messages.shared"),
      errorToast: t("documentsPage.errors.share"),
    },
  );
  const [removeDocument, isRemoving] = useMutation<EmployeePortalDocumentListItem_removeMutation>(
    removeDocumentMutation,
    {
      successMessage: t("documentsPage.messages.removed"),
      errorToast: t("documentsPage.errors.remove"),
    },
  );
  const isMutating = isAdding || isRemoving;
  const canToggle = employeePortal.canUpdate && !isMutating;

  function handleCheckedChange(checked: boolean) {
    if (!canToggle || checked === isLinked) {
      return;
    }

    if (checked) {
      setPendingLinked(true);
      void addDocument({
        variables: {
          input: {
            employeePortalId: employeePortal.id,
            documentId: document.id,
          },
        },
      }).then(
        () => setPendingLinked(null),
        () => setPendingLinked(null),
      );
      return;
    }

    setConfirmOpen(true);
  }

  function handleConfirmRemove() {
    if (link == null) {
      setConfirmOpen(false);
      return;
    }

    setPendingLinked(false);
    void removeDocument({
      variables: {
        input: { id: link.id },
      },
    }).then(
      () => {
        setPendingLinked(null);
        setConfirmOpen(false);
      },
      () => setPendingLinked(null),
    );
  }

  if (published == null) {
    return null;
  }

  return (
    <ListItem>
      <Checkbox
        checked={isLinked}
        disabled={!canToggle}
        onCheckedChange={(checked) => {
          if (typeof checked === "boolean") {
            handleCheckedChange(checked);
          }
        }}
        aria-label={t("documentsPage.actions.share", { title })}
      />
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {title}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {t(`documentTypeOptions.types.${published.documentType}`)}
        </Text>
      </ListItemContent>
      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title={t("documentsPage.confirm.title")}
        description={t("documentsPage.confirm.description")}
        cancelLabel={t("common.actions.close")}
        confirmLabel={t("documentsPage.confirm.confirm")}
        loading={isRemoving}
        onConfirm={handleConfirmRemove}
      />
    </ListItem>
  );
}
