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

import { DotsThreeVerticalIcon, PencilSimpleIcon } from "@phosphor-icons/react";
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { CardButton } from "@probo/ui/src/v2/Card/CardButton";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, Suspense, useCallback, useLayoutEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ResolvedAddableConnectorCard_account$key } from "#/__generated__/core/ResolvedAddableConnectorCard_account.graphql";
import type {
  ResolvedAddableConnectorCard_connector$data,
  ResolvedAddableConnectorCard_connector$key,
} from "#/__generated__/core/ResolvedAddableConnectorCard_connector.graphql";
import type {
  useConnectorStatuses_connector$data,
  useConnectorStatuses_connector$key,
} from "#/__generated__/core/useConnectorStatuses_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import type { TonedCardTone } from "#/components/TonedCard/variants";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { ConnectionStatusPending } from "#/pages/organizations/_components/ConnectionStatusPending";
import {
  aggregateConnectionTone,
  connectionSignalFrom,
  type ConnectionTone,
  type ConnectorConnectionStatus,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";
import { ConnectorProviderName } from "#/pages/organizations/settings/integrations/_components/ConnectorProviderName";
import { connectorProviderAccessibleName } from "#/pages/organizations/settings/integrations/_lib/connectorProviderLabel";
import { connectorDetailsPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";
import { useConnectorStatuses } from "#/pages/organizations/settings/integrations/_lib/useConnectorStatuses";

import { listedConnectorAccounts } from "../_lib/listedConnectorAccounts";

import type { LoadedConnectorAccounts } from "./AddableConnectorAccounts";
import { addableConnectorCard } from "./variants";

const connectorFragment = graphql`
  fragment ResolvedAddableConnectorCard_connector on Connector @relay(plural: true) {
    id
    displayName
    provider
    canReconnect
    distinctAccountCount
  }
`;

const accountFragment = graphql`
  fragment ResolvedAddableConnectorCard_account on ConnectorAccount @relay(plural: true) {
    id
    name
    externalAccountId
  }
`;

export interface AddableConnectorAccount {
  id: string;
  name: string;
}

export interface AddableConnectorCard {
  provider: string;
  accounts: readonly AddableConnectorAccount[];
  selectable: boolean;
  card: ReactNode;
}

interface ResolvedAddableConnectorCardProps {
  pages: readonly LoadedConnectorAccounts[];
  connectorKeys: ResolvedAddableConnectorCard_connector$key;
  // The same connectors, read through the fragment the list spreads with @defer.
  statusKeys: useConnectorStatuses_connector$key;
  normalizedSearch: string;
  selectedProviders: ReadonlySet<string>;
  onSelectedChange: (provider: string, selected: boolean) => void;
  children: (card: AddableConnectorCard | null) => ReactNode;
}

interface LoadedAccount {
  id: string;
  name: string;
  externalAccountId: string;
  connectorId: string;
}

interface LoadedAddableConnector {
  connector: ResolvedAddableConnectorCard_connector$data[number];
  accounts: readonly LoadedAccount[];
}

interface AddableSelection {
  accounts: readonly AddableConnectorAccount[];
  selectable: boolean;
}

interface AddableConnectorStatusView extends AddableSelection {
  tone: TonedCardTone;
  pending: boolean;
  showNeedsOrganization: boolean;
  statusTone: ConnectionTone | null;
  sharedStatus: ConnectorConnectionStatus | null;
  connectedCount: number;
  presentedCount: number;
}

// The initial account is the connector itself. Until the probe says the
// provider has no organization picker, that account cannot be added.
function accountNeedsOrganization(
  externalAccountId: string,
  connectorId: string,
  statuses: useConnectorStatuses_connector$data | null,
  connectors: readonly { id: string }[],
) {
  if (externalAccountId !== connectorId) {
    return false;
  }

  const index = connectors.findIndex(connector => connector.id === connectorId);
  const organizationsStatus = statuses?.[index]?.providerOrganizations.status;
  if (organizationsStatus == null) {
    return true;
  }

  return organizationsStatus !== "NOT_APPLICABLE";
}

function addableConnectorStatusView(
  connectors: readonly { id: string; canReconnect: boolean }[],
  loaded: readonly LoadedAddableConnector[],
  listed: readonly LoadedAccount[],
  statuses: useConnectorStatuses_connector$data | null,
  hasNext: boolean,
): AddableConnectorStatusView {
  const needsOrganization = (account: LoadedAccount) => accountNeedsOrganization(
    account.externalAccountId,
    account.connectorId,
    statuses,
    connectors,
  );
  const presented = statuses == null
    ? []
    : connectors.flatMap((connector, index) => {
        const status = statuses[index];
        const signal = connectionSignalFrom({
          connectionStatus: status?.connectionStatus,
          canReconnect: connector.canReconnect,
          providerOrganizations: status?.providerOrganizations,
        });
        return signal == null ? [] : [presentConnection(signal)];
      });
  const [firstPresented] = presented;
  const sharedStatus = firstPresented != null
    && presented.every(item => item.status === firstPresented.status)
    ? firstPresented.status
    : null;
  const addable = loaded
    .flatMap(({ accounts }) => accounts)
    .filter(account => !needsOrganization(account));
  const visibleAddable = listed.filter(account => !needsOrganization(account));
  const statusTone = presented.length === 0 ? null : aggregateConnectionTone(presented);

  return {
    tone: statusTone ?? "sand",
    pending: statuses == null,
    selectable: visibleAddable.length > 0 && !hasNext,
    accounts: addable.map(account => ({ id: account.id, name: account.name })),
    showNeedsOrganization: statuses != null
      && listed.length > 0
      && visibleAddable.length === 0,
    statusTone,
    sharedStatus,
    connectedCount: presented.filter(item => item.status === "CONNECTED").length,
    presentedCount: presented.length,
  };
}

function selectionsEqual(current: AddableSelection | null, next: AddableSelection) {
  if (current == null || current.selectable !== next.selectable) {
    return false;
  }
  if (current.accounts.length !== next.accounts.length) {
    return false;
  }

  return current.accounts.every((account, index) => {
    const other = next.accounts[index];
    return other != null && account.id === other.id && account.name === other.name;
  });
}

interface AddableConnectorCardFrameProps {
  face: ResolvedAddableConnectorCard_connector$data[number];
  view: AddableConnectorStatusView;
  providerSelected: boolean;
  onSelectedChange: (provider: string, selected: boolean) => void;
}

function AddableConnectorCardFrame({
  face,
  view,
  providerSelected,
  onSelectedChange,
}: AddableConnectorCardFrameProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { t: tConnector } = useTranslation("organizations/settings/integrations");
  const selected = view.selectable && providerSelected;
  const { controls, checkbox, menu, badges, frame } = addableConnectorCard({
    selectable: view.selectable,
  });
  const selectLabel = t("newAccessReviewSourcePage.actions.select", {
    name: connectorProviderAccessibleName(face.provider, face.displayName),
  });

  return (
    <TonedCard
      aria-busy={view.pending || undefined}
      tone={view.tone}
      size={2}
      className={frame()}
      stretch={view.selectable
        ? (
            <CardButton
              type="button"
              aria-label={selectLabel}
              aria-pressed={selected}
              onClick={() => {
                onSelectedChange(face.provider, !selected);
              }}
            />
          )
        : undefined}
      icon={(
        <ThirdPartyLogo thirdParty={face.provider} />
      )}
      lead={(
        <Heading level={3} size={3} weight="medium" highContrast className="flex min-w-0">
          <ConnectorProviderName provider={face.provider} name={face.displayName} />
        </Heading>
      )}
      control={(
        <div className={controls()}>
          <Checkbox
            className={checkbox()}
            checked={selected}
            disabled={!view.selectable}
            tabIndex={view.selectable ? -1 : undefined}
            aria-hidden={view.selectable || undefined}
            aria-label={selectLabel}
            onCheckedChange={(checked) => {
              onSelectedChange(face.provider, checked);
            }}
          />
          <div className={menu()}>
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton
                    variant="ghost"
                    color="neutral"
                    size={1}
                    aria-label={t("accessReviewSourcesPage.actions.more")}
                  >
                    <DotsThreeVerticalIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup align="end">
                <DropdownItem
                  iconStart={<PencilSimpleIcon />}
                  render={(
                    <Link to={connectorDetailsPath(organizationId, face.provider)} />
                  )}
                >
                  {t("accessReviewSourcesPage.actions.edit")}
                </DropdownItem>
              </DropdownPopup>
            </Dropdown>
          </div>
        </div>
      )}
    >
      {view.showNeedsOrganization && (
        <Text size={2} color="faint">
          {t("accessReviewSourcesPage.needsOrganization")}
        </Text>
      )}
      <div className={badges()}>
        {view.pending
          ? <ConnectionStatusPending />
          : view.statusTone != null && (
            <Badge variant="soft" color={view.statusTone} size={1}>
              {view.sharedStatus != null
                ? tConnector(`detailsPage.status.${view.sharedStatus}`)
                : tConnector("listPage.connectedCount", {
                    connected: view.connectedCount,
                    total: view.presentedCount,
                  })}
            </Badge>
          )}
        <Badge variant="soft" color="neutral" size={1}>
          {t("accessReviewSourcesPage.accountCount", {
            count: face.distinctAccountCount,
          })}
        </Badge>
      </div>
    </TonedCard>
  );
}

interface AddableConnectorCardStatusesProps {
  face: ResolvedAddableConnectorCard_connector$data[number];
  statusKeys: useConnectorStatuses_connector$key;
  connectors: readonly { id: string; canReconnect: boolean }[];
  loaded: readonly LoadedAddableConnector[];
  listed: readonly LoadedAccount[];
  hasNext: boolean;
  providerSelected: boolean;
  onSelectedChange: (provider: string, selected: boolean) => void;
  onSelection: (selection: AddableSelection) => void;
}

// Reads the deferred probe. Suspends until it arrives, so it has to render
// under the card's own Suspense boundary: the page collects every card
// before painting, and a suspend here would hold that collection.
function AddableConnectorCardStatuses({
  face,
  statusKeys,
  connectors,
  loaded,
  listed,
  hasNext,
  providerSelected,
  onSelectedChange,
  onSelection,
}: AddableConnectorCardStatusesProps) {
  const statuses = useConnectorStatuses(statusKeys);
  const view = addableConnectorStatusView(
    connectors,
    loaded,
    listed,
    statuses ?? null,
    hasNext,
  );

  useLayoutEffect(() => {
    onSelection({
      accounts: view.accounts,
      selectable: view.selectable,
    });
  }, [onSelection, view.accounts, view.selectable]);

  return (
    <AddableConnectorCardFrame
      face={face}
      view={view}
      providerSelected={providerSelected}
      onSelectedChange={onSelectedChange}
    />
  );
}

export function ResolvedAddableConnectorCard({
  pages,
  connectorKeys,
  statusKeys,
  normalizedSearch,
  selectedProviders,
  onSelectedChange,
  children,
}: ResolvedAddableConnectorCardProps) {
  const connectors = useFragment(connectorFragment, connectorKeys);
  const accountKeys: ResolvedAddableConnectorCard_account$key = pages.flatMap(
    page => page.accounts,
  );
  const accounts = useFragment(accountFragment, accountKeys);
  const [selection, setSelection] = useState<AddableSelection | null>(null);
  const reportSelection = useCallback((next: AddableSelection) => {
    setSelection(current => selectionsEqual(current, next) ? current : next);
  }, []);
  const [face] = connectors;
  const loaded = pages.flatMap((page, index) => {
    const connector = connectors[index];
    const start = pages
      .slice(0, index)
      .reduce((sum, item) => sum + item.accounts.length, 0);
    const pageAccounts = accounts.slice(start, start + page.accounts.length);
    if (connector == null) {
      return [];
    }
    return [{
      connector,
      accounts: pageAccounts.map(account => ({
        id: account.id,
        name: account.name,
        externalAccountId: account.externalAccountId,
        connectorId: connector.id,
      })),
    }];
  });
  const listed = loaded.flatMap(({ connector, accounts: pageAccounts }) =>
    listedConnectorAccounts(pageAccounts, connector, normalizedSearch) ?? [],
  );
  const hasNext = pages.some(page => page.hasNext);
  if (face == null || (listed.length === 0 && !hasNext)) {
    return children(null);
  }

  const pending = addableConnectorStatusView(
    connectors,
    loaded,
    listed,
    null,
    hasNext,
  );
  const applied = selection ?? pending;
  const providerSelected = selectedProviders.has(face.provider);

  return children({
    provider: face.provider,
    accounts: applied.accounts,
    selectable: applied.selectable,
    card: (
      <Suspense
        fallback={(
          <AddableConnectorCardFrame
            face={face}
            view={pending}
            providerSelected={providerSelected}
            onSelectedChange={onSelectedChange}
          />
        )}
      >
        <AddableConnectorCardStatuses
          face={face}
          statusKeys={statusKeys}
          connectors={connectors}
          loaded={loaded}
          listed={listed}
          hasNext={hasNext}
          providerSelected={providerSelected}
          onSelectedChange={onSelectedChange}
          onSelection={reportSelection}
        />
      </Suspense>
    ),
  });
}
