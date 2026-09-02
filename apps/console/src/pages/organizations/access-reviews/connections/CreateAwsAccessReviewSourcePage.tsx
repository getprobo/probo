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

import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  Field,
  IconSquareBehindSquare2,
  Input,
  PageHeader,
  useToast,
} from "@probo/ui";
import { type ChangeEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { CreateAwsAccessReviewSourcePageCreateMutation } from "#/__generated__/core/CreateAwsAccessReviewSourcePageCreateMutation.graphql";
import type { CreateAwsAccessReviewSourcePageDeleteMutation } from "#/__generated__/core/CreateAwsAccessReviewSourcePageDeleteMutation.graphql";
import type { CreateAwsAccessReviewSourcePageQuery } from "#/__generated__/core/CreateAwsAccessReviewSourcePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  ActionSplitButton,
  type ActionSplitButtonAction,
} from "../_components/ActionSplitButton";
import { ConnectorDocumentationLink } from "../dialogs/_components/ConnectorDocumentationLink";
import {
  AWS_IAM_ROLE_ARN_PATTERN,
  awsAccessReviewSourceName,
  isAWSRoleARN,
} from "../dialogs/_lib/connectorSettings";
import { createAccessReviewSourceMutation, prependCreatedSourceEdge } from "../dialogs/accessReviewSourceMutations";

export const createAwsAccessReviewSourcePageQuery = graphql`
  query CreateAwsAccessReviewSourcePageQuery($organizationId: ID!) {
    awsConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
      cloudFormationQuickCreateURL
    }
    accessReviewDrivers {
      provider
      displayName
      documentationUrl
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "access-review:source:create")
      }
    }
  }
`;

const createWorkloadIdentityConnectorMutation = graphql`
  mutation CreateAwsAccessReviewSourcePageCreateMutation(
    $input: CreateWorkloadIdentityConnectorInput!
  ) {
    createWorkloadIdentityConnector(input: $input) {
      connector {
        id
        provider
        connectionStatus
      }
    }
  }
`;

const deleteConnectorMutation = graphql`
  mutation CreateAwsAccessReviewSourcePageDeleteMutation(
    $input: DeleteConnectorInput!
  ) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

interface CreateAwsAccessReviewSourcePageProps {
  queryRef: PreloadedQuery<CreateAwsAccessReviewSourcePageQuery>;
}

export function CreateAwsAccessReviewSourcePage({
  queryRef,
}: CreateAwsAccessReviewSourcePageProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const [roleArn, setRoleArn] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  usePageTitle(t("createAwsAccessReviewSourcePage.pageTitle"));

  const { organization, awsConnectorSetup, accessReviewDrivers }
    = usePreloadedQuery<CreateAwsAccessReviewSourcePageQuery>(
      createAwsAccessReviewSourcePageQuery,
      queryRef,
    );
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const awsDriver = accessReviewDrivers.find(
    driver => driver.provider === "AWS",
  );
  if (!awsDriver) {
    throw new Error("AWS access review driver not found");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewConnectionsPage_accessReviewSources",
  );

  const [createWorkloadIdentityConnector] = useMutation<
    CreateAwsAccessReviewSourcePageCreateMutation
  >(createWorkloadIdentityConnectorMutation);
  const [deleteConnector] = useMutation<
    CreateAwsAccessReviewSourcePageDeleteMutation
  >(deleteConnectorMutation);
  const [createAccessReviewSource] = useMutation<
    accessReviewSourceMutationsCreateMutation
  >(createAccessReviewSourceMutation);

  if (!organization.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {t("createAwsAccessReviewSourcePage.permissionDenied")}
        </p>
      </Card>
    );
  }

  const copyValue = (value: string, successKey: string) => {
    const onCopyFailure = () =>
      toast({
        title: t("createAwsAccessReviewSourcePage.messages.copyFailed"),
        description: t("createAwsAccessReviewSourcePage.errors.copy"),
        variant: "error",
      });

    if (!navigator.clipboard?.writeText) {
      onCopyFailure();
      return;
    }

    try {
      navigator.clipboard.writeText(value).then(
        () =>
          toast({
            title: t("createAwsAccessReviewSourcePage.messages.copied"),
            description: t(successKey),
            variant: "success",
          }),
        onCopyFailure,
      );
    } catch {
      onCopyFailure();
    }
  };

  const roleArnValid = isAWSRoleARN(roleArn);
  const roleArnInvalid = roleArn.trim() !== "" && !roleArnValid;

  const onSubmit = async () => {
    if (!roleArnValid) {
      return;
    }

    setIsCreating(true);

    try {
      const created = await createWorkloadIdentityConnector(
        {
          variables: {
            input: {
              organizationId,
              provider: "AWS",
              awsRoleArn: roleArn.trim(),
            },
          },
        },
        { errorToast: t("createAwsAccessReviewSourcePage.errors.create") },
      );
      const { id: connectorId, connectionStatus }
        = created.createWorkloadIdentityConnector.connector;

      const discardConnector = () =>
        deleteConnector(
          { variables: { input: { connectorId } } },
          { errorToast: t("createAwsAccessReviewSourcePage.errors.delete") },
        );

      if (connectionStatus !== "CONNECTED") {
        toast({
          title: t("createAwsAccessReviewSourcePage.messages.error"),
          description: t(
            "createAwsAccessReviewSourcePage.errors.disconnected",
          ),
          variant: "error",
        });
        await discardConnector();
        return;
      }

      try {
        await createAccessReviewSource(
          {
            variables: {
              input: {
                organizationId,
                connectorId,
                name: awsAccessReviewSourceName(awsDriver.displayName, roleArn),
                csvData: null,
              },
            },
            updater: (store) => {
              if (connectionId) {
                prependCreatedSourceEdge(store, connectionId);
              }
            },
          },
          { errorToast: t("createAwsAccessReviewSourcePage.errors.source") },
        );
      } catch {
        await discardConnector();
        return;
      }

      toast({
        title: t("createAwsAccessReviewSourcePage.messages.success"),
        description: t("createAwsAccessReviewSourcePage.messages.created"),
        variant: "success",
      });
      void navigate(`/organizations/${organizationId}/access-reviews/connections`);
    } catch {
      return;
    } finally {
      setIsCreating(false);
    }
  };

  const installActions: ActionSplitButtonAction[] = [];
  if (awsConnectorSetup.cloudFormationQuickCreateURL) {
    installActions.push({
      id: "cloudformation",
      label: t(
        "createAwsAccessReviewSourcePage.actions.installViaCloudFormation",
      ),
      href: awsConnectorSetup.cloudFormationQuickCreateURL,
    });
  }
  if (awsConnectorSetup.terraformSnippet) {
    installActions.push({
      id: "terraform",
      label: t("createAwsAccessReviewSourcePage.actions.installViaTerraform"),
      onSelect: () =>
        copyValue(
          awsConnectorSetup.terraformSnippet,
          "createAwsAccessReviewSourcePage.messages.copiedTerraform",
        ),
    });
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("createAwsAccessReviewSourcePage.title")}
        description={t("createAwsAccessReviewSourcePage.description")}
      >
        {installActions.length > 0 && (
          <ActionSplitButton
            actions={installActions}
            chooseAnotherMethodLabel={t(
              "createAwsAccessReviewSourcePage.actions.chooseAnotherInstallMethod",
            )}
          />
        )}
      </PageHeader>

      <Card padded>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void onSubmit();
          }}
          className="space-y-4"
        >
          <Field
            name="roleArn"
            label={t("createAwsAccessReviewSourcePage.fields.roleArn")}
            value={roleArn}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setRoleArn(e.target.value)}
            required
            pattern={AWS_IAM_ROLE_ARN_PATTERN}
            placeholder={t(
              "createAwsAccessReviewSourcePage.fields.roleArnPlaceholder",
            )}
            error={
              roleArnInvalid
                ? t("createAwsAccessReviewSourcePage.errors.roleArn")
                : undefined
            }
          />
          {(
            [
              {
                name: "issuer",
                value: awsConnectorSetup.issuer,
                successKey:
                  "createAwsAccessReviewSourcePage.messages.copiedIssuer",
              },
              {
                name: "audience",
                value: awsConnectorSetup.audience,
                successKey:
                  "createAwsAccessReviewSourcePage.messages.copiedAudience",
              },
              {
                name: "subject",
                value: awsConnectorSetup.subject,
                successKey:
                  "createAwsAccessReviewSourcePage.messages.copiedSubject",
              },
            ] as const
          ).map(row => (
            <Field
              key={row.name}
              name={row.name}
              label={t(`createAwsAccessReviewSourcePage.fields.${row.name}`)}
            >
              <div className="flex items-center gap-2">
                <div className="min-w-0 flex-1">
                  <Input
                    id={row.name}
                    name={row.name}
                    value={row.value}
                    readOnly
                    disabled
                  />
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  icon={IconSquareBehindSquare2}
                  onClick={() => copyValue(row.value, row.successKey)}
                  aria-label={t("createAwsAccessReviewSourcePage.actions.copy")}
                />
              </div>
            </Field>
          ))}

          <div className="flex items-center justify-between gap-2">
            <ConnectorDocumentationLink
              url={awsDriver.documentationUrl}
              variant="button"
            />
            <div className="flex items-center justify-end gap-2">
              <Button variant="secondary" asChild>
                <Link to={`/organizations/${organizationId}/access-reviews/connections`}>
                  {t("createAwsAccessReviewSourcePage.actions.back")}
                </Link>
              </Button>
              <Button disabled={!roleArnValid || isCreating} type="submit">
                {t("createAwsAccessReviewSourcePage.actions.connect")}
              </Button>
            </div>
          </div>
        </form>
      </Card>
    </div>
  );
}
