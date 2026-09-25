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

import { CaretLeftIcon, PlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { ThirdPartyLogo } from "@probo/ui";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useNavigate } from "react-router";

import type { ConnectorDetailsPageQuery } from "#/__generated__/core/ConnectorDetailsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { ConnectorAccountsDrawer } from "./_components/ConnectorAccountsDrawer";
import { ConnectorDocumentationLink } from "./_components/ConnectorDocumentationLink";
import { ConnectorListItem } from "./_components/ConnectorListItem";
import {
  createdConnectorLocationState,
  type DiscoveredAccount,
} from "./_lib/discoveredAccounts";
import { connectVendorPath, integrationListPath } from "./_lib/integrationPath";
import { connectorDetailsPage } from "./variants";

export const connectorDetailsPageQuery = graphql`
  query ConnectorDetailsPageQuery(
    $organizationId: ID!
    $provider: ConnectorProvider!
  ) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
        connectors(filter: { providers: [$provider] }) {
          id
          provider
          displayName
          documentationUrl
          ...ConnectorListItem_connector @arguments(
            includeAccountCount: true
            includeOrganizationSelect: true
          )
        }
      }
    }
  }
`;

interface ConnectorDetailsPageProps {
  queryRef: PreloadedQuery<ConnectorDetailsPageQuery>;
}

export function ConnectorDetailsPage({ queryRef }: ConnectorDetailsPageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const location = useLocation();
  const created = createdConnectorLocationState(location.state);
  const [openId, setOpenId] = useState<string | null>(() => created?.connectorId ?? null);
  const [preset, setPreset] = useState<DiscoveredAccount[] | null>(() => (
    created?.discoveredAccounts ?? null
  ));
  const [presetConnectorId, setPresetConnectorId] = useState<string | null>(() => (
    created?.connectorId ?? null
  ));
  const [fetchKey, setFetchKey] = useState(0);
  const { organization } = usePreloadedQuery<ConnectorDetailsPageQuery>(
    connectorDetailsPageQuery,
    queryRef,
  );
  const connectors = organization.__typename === "Organization"
    ? organization.connectors
    : [];
  const vendor = connectors[0];
  // Keep the last id so the close animation still has the account list.
  const [displayedId, setDisplayedId] = useState<string | null>(openId);
  if (openId != null && displayedId !== openId) {
    setDisplayedId(openId);
  }
  const activeId = openId ?? displayedId;
  const displayed = connectors.find(connector => connector.id === activeId) ?? null;

  usePageTitle(vendor?.displayName ?? "");

  if (organization.__typename !== "Organization" || vendor == null) {
    throw new NotFoundError(t("detailsPage.notFound"));
  }

  const { root, back, header, intro, title, grid } = connectorDetailsPage();

  function clearCreatedState() {
    if (createdConnectorLocationState(location.state) == null) {
      return;
    }

    void navigate(
      { pathname: location.pathname, search: location.search },
      { replace: true, state: null },
    );
  }

  function openAccounts(connectorId: string) {
    clearCreatedState();
    setPreset(null);
    setPresetConnectorId(null);
    setFetchKey(key => key + 1);
    setOpenId(connectorId);
  }

  return (
    <div className={root()}>
      <Link
        to={integrationListPath(organizationId)}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("detailsPage.actions.back")}
      </Link>
      <div className={header()}>
        <div className={intro()}>
          <div className={title()}>
            <ThirdPartyLogo
              thirdParty={vendor.provider}
              className="size-8 shrink-0"
            />
            <Heading level={1} size={6} weight="medium" highContrast>
              {vendor.displayName}
            </Heading>
          </div>
          <ConnectorDocumentationLink url={vendor.documentationUrl} />
        </div>
        {organization.canCreateConnector && (
          <ButtonLink
            to={connectVendorPath(organizationId, vendor.provider)}
            variant="solid"
            iconStart={<PlusIcon />}
          >
            {t("listPage.actions.add")}
          </ButtonLink>
        )}
      </div>
      <div className={grid()}>
        {connectors.map(connector => (
          <ConnectorListItem
            key={connector.id}
            connectorKey={connector}
            organizationId={organizationId}
            canConnect={false}
            showName={false}
            showConnectorType
            showProbeError
            onSelect={openAccounts}
            onDeleted={() => {
              if (connector.id === openId) {
                setOpenId(null);
                clearCreatedState();
              }
              if (connectors.length <= 1) {
                void navigate(integrationListPath(organizationId));
              }
            }}
          />
        ))}
      </div>
      <ConnectorAccountsDrawer
        connectorId={activeId}
        provider={displayed?.provider ?? vendor.provider}
        providerName={displayed?.displayName ?? vendor.displayName}
        fetchKey={fetchKey}
        preset={activeId != null && activeId === presetConnectorId ? preset : null}
        open={openId != null}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
            setOpenId(null);
            clearCreatedState();
          }
        }}
      />
    </div>
  );
}
