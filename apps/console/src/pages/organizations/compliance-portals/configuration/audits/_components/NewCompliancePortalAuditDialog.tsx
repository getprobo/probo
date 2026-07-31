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

import { getCompliancePortalLinkedVisibilityOptions } from "@probo/helpers";
import { Badge, Button, Dialog, DialogContent, DialogFooter, Field, Option, useDialogRef } from "@probo/ui";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { type DataID, graphql, useLazyLoadQuery } from "react-relay";

import type { NewCompliancePortalAuditDialog_addMutation } from "#/__generated__/core/NewCompliancePortalAuditDialog_addMutation.graphql";
import type { NewCompliancePortalAuditDialogQuery } from "#/__generated__/core/NewCompliancePortalAuditDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const pickerQuery = graphql`
  query NewCompliancePortalAuditDialogQuery($organizationId: ID!, $compliancePortalId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        audits(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              name
              framework {
                name
              }
            }
          }
        }
      }
    }
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        audits(first: 100) {
          edges {
            node {
              audit {
                id
              }
            }
          }
        }
      }
    }
  }
`;

const addMutation = graphql`
  mutation NewCompliancePortalAuditDialog_addMutation(
    $input: UpdateCompliancePortalAuditVisibilityInput!
    $connections: [ID!]!
  ) {
    updateCompliancePortalAuditVisibility(input: $input) {
      catalogAuditEdge @appendEdge(connections: $connections) {
        cursor
        node {
          ...CompliancePortalAuditListItem_catalogAuditFragment
        }
      }
    }
  }
`;

function AuditPicker(props: {
  compliancePortalId: string;
  connectionId: DataID;
  onClose: () => void;
}) {
  const { compliancePortalId, connectionId, onClose } = props;
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");
  const visibilityOptions = getCompliancePortalLinkedVisibilityOptions(t);

  const data = useLazyLoadQuery<NewCompliancePortalAuditDialogQuery>(pickerQuery, {
    organizationId,
    compliancePortalId,
  });

  const [auditId, setAuditId] = useState("");
  const [visibility, setVisibility] = useState<"RESTRICTED" | "PUBLIC">("RESTRICTED");

  const [addAudit, isAdding] = useMutation<NewCompliancePortalAuditDialog_addMutation>(addMutation, {
    successMessage: t("auditList.addDialog.success"),
    errorToast: t("auditList.addDialog.error"),
  });

  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid organization node");
  }

  if (data.compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid compliance portal node");
  }

  const linkedAuditIDs = new Set(
    data.compliancePortal.audits.edges.map(({ node }) => node.audit.id),
  );

  const availableAudits = data.organization.audits.edges.filter(
    ({ node }) => !linkedAuditIDs.has(node.id),
  );

  const handleSubmit = async () => {
    if (auditId === "") {
      return;
    }

    await addAudit({
      variables: {
        connections: [connectionId],
        input: {
          compliancePortalId,
          auditId,
          compliancePortalVisibility: visibility,
        },
      },
    });
    onClose();
  };

  return (
    <>
      <DialogContent className="space-y-4">
        <Field
          type="select"
          label={t("auditList.addDialog.auditLabel")}
          value={auditId}
          onValueChange={value => setAuditId(typeof value === "string" ? value : "")}
        >
          {availableAudits.map(({ node }) => (
            <Option key={node.id} value={node.id}>
              {node.framework?.name ?? node.name ?? node.id}
            </Option>
          ))}
        </Field>
        <Field
          type="select"
          label={t("auditList.addDialog.visibilityLabel")}
          value={visibility}
          onValueChange={value => setVisibility(value === "PUBLIC" ? "PUBLIC" : "RESTRICTED")}
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <Badge variant={option.variant}>{option.label}</Badge>
            </Option>
          ))}
        </Field>
      </DialogContent>
      <DialogFooter>
        <Button variant="secondary" onClick={onClose}>{t("auditList.addDialog.cancel")}</Button>
        <Button disabled={isAdding || auditId === ""} onClick={() => void handleSubmit()}>
          {t("auditList.addDialog.submit")}
        </Button>
      </DialogFooter>
    </>
  );
}

export function NewCompliancePortalAuditDialog(props: {
  compliancePortalId: string;
  connectionId: DataID;
  disabled?: boolean;
}) {
  const { compliancePortalId, connectionId, disabled } = props;
  const { t } = useTranslation("organizations/compliance-portals");
  const dialogRef = useDialogRef();
  const [open, setOpen] = useState(false);

  return (
    <Dialog
      ref={dialogRef}
      trigger={(
        <Button variant="secondary" disabled={disabled}>
          {t("auditList.addAudit")}
        </Button>
      )}
      onOpenChange={setOpen}
      title={t("auditList.addDialog.title")}
    >
      {open && (
        <Suspense fallback={<p className="text-sm text-txt-secondary">{t("auditList.addDialog.loading")}</p>}>
          <AuditPicker
            compliancePortalId={compliancePortalId}
            connectionId={connectionId}
            onClose={() => dialogRef.current?.close()}
          />
        </Suspense>
      )}
    </Dialog>
  );
}
