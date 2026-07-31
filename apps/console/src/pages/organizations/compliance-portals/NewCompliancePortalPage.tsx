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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  Breadcrumb,
  Button,
  Card,
  Field,
  Input,
  PageHeader,
  useToast,
} from "@probo/ui";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { NewCompliancePortalPageMutation } from "#/__generated__/core/NewCompliancePortalPageMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createCompliancePortalMutation = graphql`
  mutation NewCompliancePortalPageMutation(
    $input: CreateCompliancePortalInput!
    $connections: [ID!]!
  ) {
    createCompliancePortal(input: $input) {
      compliancePortalEdge @prependEdge(connections: $connections) {
        node {
          id
          entityName
          active
          publicUrl
          createdAt
        }
      }
    }
  }
`;

export default function NewCompliancePortalPage() {
  const { t } = useTranslation("organizations/compliance-portals");
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();

  usePageTitle(t("newPortalPage.pageTitle"));

  const [createCompliancePortal, isCreating]
    = useMutation<NewCompliancePortalPageMutation>(createCompliancePortalMutation);

  const [entityName, setEntityName] = useState("");

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CompliancePortalsOverviewPage_compliancePortals",
  );

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    createCompliancePortal({
      variables: {
        input: {
          organizationId,
          entityName,
        },
        connections: [connectionId],
      },
      onCompleted(data) {
        toast({
          title: t("newPortalPage.messages.successTitle"),
          description: t("newPortalPage.messages.created"),
          variant: "success",
        });
        const portalId = data.createCompliancePortal.compliancePortalEdge.node.id;
        void navigate(`/organizations/${organizationId}/compliance-portals/${portalId}`);
      },
      onError(error) {
        toast({
          title: t("newPortalPage.errors.title"),
          description: formatError(t("newPortalPage.errors.create"), error),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("newPortalPage.breadcrumb.root"),
            to: `/organizations/${organizationId}/compliance-portals`,
          },
          {
            label: t("newPortalPage.breadcrumb.current"),
          },
        ]}
      />
      <PageHeader
        title={t("newPortalPage.title")}
        description={t("newPortalPage.description")}
      />
      <Card padded asChild>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Field label={t("newPortalPage.fields.entityName.label")}>
            <Input
              value={entityName}
              onChange={e => setEntityName(e.target.value)}
              placeholder={t("newPortalPage.fields.entityName.placeholder")}
              required
            />
          </Field>

          <Button type="submit" disabled={isCreating}>
            {isCreating
              ? t("newPortalPage.actions.creating")
              : t("newPortalPage.actions.create")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
