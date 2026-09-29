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
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useSearchParams } from "react-router";

import type { SCIMPageCreateSCIMConfigurationMutation } from "#/__generated__/iam/SCIMPageCreateSCIMConfigurationMutation.graphql";
import type { SCIMPageQuery } from "#/__generated__/iam/SCIMPageQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { ExportSCIMEventsDialog } from "./_components/ExportSCIMEventsDialog";
import { SCIMConfiguration } from "./_components/SCIMConfiguration";
import { SCIMEventList } from "./_components/SCIMEventList";
import { SCIMProviderCard } from "./_components/SCIMProviderCard";
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
          bridge {
            __typename
          }
          ...SCIMEventList_scimConfiguration
          ...SCIMProviderCard_scimConfiguration
        }

        ...SCIMConfiguration_organization
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
    }
  }
`;

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
  const { root, section, sectionHead, loader } = scimPage();
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
    if (organization.scimConfiguration != null) {
      setSearchParams((params: URLSearchParams) => {
        params.delete("connector_id");
        return params;
      });
      return;
    }

    mutationTriggeredRef.current = true;

    void createSCIMConfiguration({
      variables: {
        input: {
          organizationId: organization.id,
          connectorId: connectorId,
        },
      },
    }).then(() => {
      const url = new URL(window.location.href);
      url.searchParams.delete("connector_id");
      window.location.assign(url.toString());
    }).catch(() => {
      mutationTriggeredRef.current = false;
      setSearchParams((params: URLSearchParams) => {
        params.delete("connector_id");
        return params;
      });
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
      <div className={loader()}>
        <Spinner size={3} aria-label={t("scimPage.connecting")} />
      </div>
    );
  }

  if (organization.scimConfiguration == null) {
    return (
      <div className={root()}>
        <SCIMPageIntro />
        <SCIMSetupCards
          organizationKey={organization}
          onManualCreated={setCreatedToken}
        />
      </div>
    );
  }

  const hasIdentityProvider = organization.scimConfiguration.bridge != null;

  return (
    <div className={root()}>
      <SCIMPageIntro />
      {hasIdentityProvider
        ? (
            <SCIMProviderCard
              scimConfigurationKey={organization.scimConfiguration}
            />
          )
        : (
            <SCIMConfiguration
              organizationKey={organization}
              initialToken={createdToken}
              onDeleted={() => {
                setCreatedToken(null);
              }}
            />
          )}
      <div className={section()}>
        <div className={sectionHead()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("scimPage.eventHistory")}
          </Heading>
          {organization.canExportSCIMEvents && (
            <ExportSCIMEventsDialog organizationId={organization.id} />
          )}
        </div>
        <SCIMEventList scimConfigurationKey={organization.scimConfiguration} />
      </div>
    </div>
  );
}

function SCIMPageIntro() {
  const { t } = useTranslation();
  const { intro } = scimPage();

  return (
    <div className={intro()}>
      <Heading level={1} size={6} weight="medium" highContrast>
        {t("scimPage.title")}
      </Heading>
      <Text size={2} color="faint">
        {t("scimPage.description")}
      </Text>
    </div>
  );
}
