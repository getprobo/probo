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

import { Button, Dialog, DialogContent, DialogFooter, Field, Option, useDialogRef } from "@probo/ui";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { NewCompliancePortalThirdPartyDialog_addMutation } from "#/__generated__/core/NewCompliancePortalThirdPartyDialog_addMutation.graphql";
import type { NewCompliancePortalThirdPartyDialogQuery } from "#/__generated__/core/NewCompliancePortalThirdPartyDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const pickerQuery = graphql`
  query NewCompliancePortalThirdPartyDialogQuery($organizationId: ID!, $compliancePortalId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        thirdParties(first: 100, orderBy: { field: NAME, direction: ASC }) {
          edges {
            node {
              id
              name
              category
            }
          }
        }
      }
    }
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        thirdParties(first: 100) {
          edges {
            node {
              thirdParty {
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
  mutation NewCompliancePortalThirdPartyDialog_addMutation(
    $input: UpdateCompliancePortalThirdPartyPublishedInput!
    $connections: [ID!]!
  ) {
    updateCompliancePortalThirdPartyPublished(input: $input) {
      catalogThirdPartyEdge @appendEdge(connections: $connections) {
        cursor
        node {
          ...CompliancePortalThirdPartyListItem_catalogThirdPartyFragment
        }
      }
    }
  }
`;

function ThirdPartyPicker(props: {
  compliancePortalId: string;
  connectionId: DataID;
  onClose: () => void;
}) {
  const { compliancePortalId, connectionId, onClose } = props;
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");

  const data = useLazyLoadQuery<NewCompliancePortalThirdPartyDialogQuery>(pickerQuery, {
    organizationId,
    compliancePortalId,
  });

  const [thirdPartyId, setThirdPartyId] = useState("");

  const [addThirdParty, isAdding] = useMutation<NewCompliancePortalThirdPartyDialog_addMutation>(addMutation, {
    successMessage: t("thirdPartyList.addDialog.success"),
    errorToast: t("thirdPartyList.addDialog.error"),
  });

  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid organization node");
  }

  if (data.compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid compliance portal node");
  }

  const linkedThirdPartyIDs = new Set(
    data.compliancePortal.thirdParties.edges.map(({ node }) => node.thirdParty.id),
  );

  const availableThirdParties = data.organization.thirdParties.edges.filter(
    ({ node }) => !linkedThirdPartyIDs.has(node.id),
  );

  const handleSubmit = async () => {
    if (thirdPartyId === "") {
      return;
    }

    await addThirdParty({
      variables: {
        connections: [connectionId],
        input: {
          compliancePortalId,
          thirdPartyId,
          published: true,
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
          label={t("thirdPartyList.addDialog.thirdPartyLabel")}
          value={thirdPartyId}
          onValueChange={value => setThirdPartyId(typeof value === "string" ? value : "")}
        >
          {availableThirdParties.map(({ node }) => (
            <Option key={node.id} value={node.id}>
              {node.name}
            </Option>
          ))}
        </Field>
      </DialogContent>
      <DialogFooter>
        <Button variant="secondary" onClick={onClose}>{t("thirdPartyList.addDialog.cancel")}</Button>
        <Button disabled={isAdding || thirdPartyId === ""} onClick={() => void handleSubmit()}>
          {t("thirdPartyList.addDialog.submit")}
        </Button>
      </DialogFooter>
    </>
  );
}

export function NewCompliancePortalThirdPartyDialog(props: {
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
          {t("thirdPartyList.addThirdParty")}
        </Button>
      )}
      onOpenChange={setOpen}
      title={t("thirdPartyList.addDialog.title")}
    >
      {open && (
        <Suspense fallback={<p className="text-sm text-txt-secondary">{t("thirdPartyList.addDialog.loading")}</p>}>
          <ThirdPartyPicker
            compliancePortalId={compliancePortalId}
            connectionId={connectionId}
            onClose={() => dialogRef.current?.close()}
          />
        </Suspense>
      )}
    </Dialog>
  );
}
