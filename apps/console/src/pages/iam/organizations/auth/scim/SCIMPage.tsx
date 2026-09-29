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
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconArrowDown,
  Input,
  PageHeader,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { useSearchParams } from "react-router";

import type { SCIMPageCreateSCIMConfigurationMutation } from "#/__generated__/iam/SCIMPageCreateSCIMConfigurationMutation.graphql";
import type { SCIMPageExportMutation } from "#/__generated__/iam/SCIMPageExportMutation.graphql";
import type { SCIMPageQuery } from "#/__generated__/iam/SCIMPageQuery.graphql";

import { ConnectorList } from "./_components/ConnectorList";
import { SCIMConfiguration } from "./_components/SCIMConfiguration";
import { SCIMEventList } from "./_components/SCIMEventList";
import { SCIMSetupCards } from "./_components/SCIMSetupCards";
import { scimPage } from "./variants";

export const scimPageQuery = graphql`
  query SCIMPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canExportSCIMEvents: permission(action: "iam:scim-event:export")

        scimConfiguration {
          id
          bridge {
            id
          }
          ...SCIMEventListFragment
        }

        ...SCIMConfigurationFragment
        ...ConnectorListFragment
        ...SCIMSetupCards_organization
      }
    }
  }
`;

const createSCIMConfigurationMutation = graphql`
  mutation SCIMPageCreateSCIMConfigurationMutation(
    $input: CreateSCIMConfigurationInput!
  ) {
    createSCIMConfiguration(input: $input) {
      scimConfiguration {
        id
      }
      scimBridge {
        id
      }
    }
  }
`;

const exportMutation = graphql`
  mutation SCIMPageExportMutation(
    $input: RequestSCIMEventExportInput!
  ) {
    requestSCIMEventExport(input: $input) {
      exportJobId
    }
  }
`;

function ExportSCIMEventsDialog({
  organizationId,
}: {
  organizationId: string;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [commitExport, isExporting] = useMutation<SCIMPageExportMutation>(exportMutation);

  const handleExport = () => {
    if (!fromDate || !toDate) return;

    commitExport({
      variables: {
        input: {
          organizationId,
          fromTime: new Date(`${fromDate}T00:00:00Z`).toISOString(),
          toTime: new Date(Date.parse(`${toDate}T00:00:00Z`) + 24 * 60 * 60 * 1000).toISOString(),
        },
      },
      onCompleted: (_response, errors) => {
        if (errors) {
          toast({
            title: t("scimPage.export.errors.title"),
            description: formatError(t("scimPage.export.errors.request"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("scimPage.export.messages.successTitle"),
          description: t("scimPage.export.messages.success"),
          variant: "success",
        });
        dialogRef.current?.close();
        setFromDate("");
        setToDate("");
      },
      onError: (error) => {
        toast({
          title: t("scimPage.export.errors.title"),
          description: formatError(t("scimPage.export.errors.request"), error),
          variant: "error",
        });
      },
    });
  };

  return (
    <>
      <Button
        variant="secondary"
        icon={IconArrowDown}
        onClick={() => dialogRef.current?.open()}
      >
        {t("scimPage.export.actions.export")}
      </Button>
      <Dialog
        className="max-w-md"
        ref={dialogRef}
        title={t("scimPage.export.title")}
      >
        <DialogContent className="space-y-4" padded>
          <p className="text-sm text-txt-secondary">
            {t("scimPage.export.description")}
          </p>
          <Field label={t("scimPage.export.fields.from")}>
            <Input
              type="date"
              value={fromDate}
              onChange={e => setFromDate(e.target.value)}
              required
            />
          </Field>
          <Field label={t("scimPage.export.fields.to")}>
            <Input
              type="date"
              value={toDate}
              onChange={e => setToDate(e.target.value)}
              required
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button
            onClick={handleExport}
            disabled={isExporting || !fromDate || !toDate || fromDate > toDate}
          >
            {isExporting
              ? (
                  <>
                    <Spinner size={16} />
                    {t("scimPage.export.actions.exporting")}
                  </>
                )
              : t("scimPage.export.actions.export")}
          </Button>
        </DialogFooter>
      </Dialog>
    </>
  );
}

export function SCIMPage(props: {
  queryRef: PreloadedQuery<SCIMPageQuery>;
}) {
  const { queryRef } = props;
  const { t } = useTranslation();
  usePageTitle(t("nav.scim"));
  const [searchParams, setSearchParams] = useSearchParams();
  const connectorId = searchParams.get("connector_id");
  const mutationTriggeredRef = useRef(false);

  const { organization } = usePreloadedQuery<SCIMPageQuery>(scimPageQuery, queryRef);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const { root, header, intro } = scimPage();
  const [createSCIMConfiguration]
    = useMutation<SCIMPageCreateSCIMConfigurationMutation>(
      createSCIMConfigurationMutation,
    );

  // Auto-create SCIM configuration and bridge when connector_id is in URL
  useEffect(() => {
    if (organization.__typename !== "Organization") {
      return;
    }
    if (!connectorId || mutationTriggeredRef.current) return;

    // Don't create if SCIM config already exists
    if (organization.scimConfiguration?.id) {
      setSearchParams((params: URLSearchParams) => {
        params.delete("connector_id");
        return params;
      });
      return;
    }

    mutationTriggeredRef.current = true;

    createSCIMConfiguration({
      variables: {
        input: {
          organizationId: organization.id,
          connectorId: connectorId,
        },
      },
      onCompleted: () => {
        const url = new URL(window.location.href);
        url.searchParams.delete("connector_id");
        window.location.href = url.toString();
      },
      onError: (error) => {
        console.error("Failed to create SCIM configuration:", error);
        mutationTriggeredRef.current = false;
        setSearchParams((params: URLSearchParams) => {
          params.delete("connector_id");
          return params;
        });
      },
    });
  }, [
    connectorId,
    organization,
    createSCIMConfiguration,
    setSearchParams,
  ]);

  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  // Show loader while creating SCIM configuration
  if (connectorId) {
    return (
      <div className="flex items-center justify-center py-12">
        <Spinner size={24} />
      </div>
    );
  }

  if (organization.scimConfiguration == null) {
    return (
      <div className={root()}>
        <div className={header()}>
          <div className={intro()}>
            <Heading level={1} size={6} weight="medium" highContrast>
              {t("scimPage.title")}
            </Heading>
            <Text size={2} color="faint">
              {t("scimPage.description")}
            </Text>
          </div>
        </div>
        <SCIMSetupCards
          organizationKey={organization}
          onManualCreated={setCreatedToken}
        />
      </div>
    );
  }

  const hasIdentityProvider = organization.scimConfiguration.bridge != null;

  return (
    <div className="space-y-8">
      <PageHeader title={t("nav.scim")} />
      {hasIdentityProvider && (
        <ConnectorList fKey={organization} />
      )}

      {!hasIdentityProvider && (
        <div className="space-y-4">
          <h2 className="text-base font-medium">{t("scimPage.manualScim.title")}</h2>
          <SCIMConfiguration
            fKey={organization}
            initialToken={createdToken}
          />
        </div>
      )}

      <div className="space-y-4">
        <div className="flex items-start justify-between">
          <h2 className="text-base font-medium">
            {t("scimPage.provisioningEventHistory")}
          </h2>
          {organization.canExportSCIMEvents && (
            <ExportSCIMEventsDialog organizationId={organization.id} />
          )}
        </div>
        <SCIMEventList fKey={organization.scimConfiguration} />
      </div>
    </div>
  );
}
