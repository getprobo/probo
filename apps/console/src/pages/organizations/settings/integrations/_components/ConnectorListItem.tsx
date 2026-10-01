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

import { CopyIcon, TrashIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { IconWarning, ThirdPartyLogo, useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type {
  ConnectorListItem_connector$data,
  ConnectorListItem_connector$key,
} from "#/__generated__/core/ConnectorListItem_connector.graphql";
import type { ConnectorListItemUpdateNameMutation } from "#/__generated__/core/ConnectorListItemUpdateNameMutation.graphql";
import type { ConnectorProviderListItem_provider$key } from "#/__generated__/core/ConnectorProviderListItem_provider.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";

import { connectMethodFromConnector } from "../_lib/connectMethods";
import { buildConnectorInitiateURL } from "../_lib/connectorSettings";
import { connectorDetailsPath } from "../_lib/integrationPath";
import { connectorCard } from "../variants";

import { ConnectorConnectMore } from "./ConnectorConnectMore";
import { ConnectorDeleteDialog } from "./ConnectorDeleteDialog";
import { ConnectorMethodIcon } from "./ConnectorMethodIcon";
import { ConnectorModules, ConnectorModulesSkeleton } from "./ConnectorModules";
import { ConnectorOrganizationSelect } from "./ConnectorOrganizationSelect";

// connectionStatus probes the vendor on every read, so this list pays one
// outbound call per connected connector. providerOrganizations does too, for
// providers that have an account picker.
export const connectorListItemFragment = graphql`
  fragment ConnectorListItem_connector on Connector
  @argumentDefinitions(
    includeAccountCount: { type: "Boolean!", defaultValue: false }
    includeOrganizationSelect: { type: "Boolean!", defaultValue: false }
  ) {
    id
    name
    provider
    displayName
    connectionStatus
    canReconnect
    protocol
    oauth2Scopes
    providerOrganizations {
      status
    }
    accounts(first: 1) @include(if: $includeAccountCount) {
      totalCount
    }
    createdAt
    canGet: permission(action: "core:connector:get")
    canUpdate: permission(action: "core:connector:update")
    canDelete: permission(action: "core:connector:delete")
    ...ConnectorOrganizationSelect_connector @include(if: $includeOrganizationSelect)
    ...ConnectorDeleteDialog_connector
  }
`;

const updateConnectorNameMutation = graphql`
  mutation ConnectorListItemUpdateNameMutation($input: UpdateConnectorInput!) {
    updateConnector(input: $input) {
      connector {
        id
        name
      }
    }
  }
`;

type ConnectionStatus = "CONNECTED" | "DISCONNECTED" | "NOT_AUTHORIZED" | "RECONNECT_REQUIRED";

type OrganizationsStatus = "AVAILABLE" | "EMPTY" | "NOT_APPLICABLE" | "UNAVAILABLE";

type IssueKey = "reconnect" | "notAuthorized" | "credentials";

type ConnectionTone = "green" | "amber" | "red";

interface ConnectionSignal {
  status: ConnectionStatus;
  organizationsStatus: OrganizationsStatus;
  canReconnect: boolean;
}

interface ConnectionPresentation {
  status: ConnectionStatus;
  issue: IssueKey | null;
}

// A record linked in by a narrower query (modules, or a create that returned
// only an id) has no status yet. Absent is not a failure: the page refetch
// writes the real one. A missing account list is also not a failure.
export function connectionSignalFrom(connector: {
  connectionStatus?: ConnectionStatus | null;
  canReconnect?: boolean | null;
  providerOrganizations?: { status?: OrganizationsStatus | null } | null;
}): ConnectionSignal | null {
  if (connector.connectionStatus == null) {
    return null;
  }

  return {
    status: connector.connectionStatus,
    organizationsStatus: connector.providerOrganizations?.status ?? "NOT_APPLICABLE",
    canReconnect: connector.canReconnect ?? false,
  };
}

// One outcome for the badge and the warning icon. A single card uses the same
// status for its color. An account list that cannot be read is a failure,
// including when the token-verify probe still says connected.
function presentConnection(signal: ConnectionSignal): ConnectionPresentation {
  if (signal.organizationsStatus === "UNAVAILABLE") {
    if (signal.status === "NOT_AUTHORIZED") {
      return { status: "NOT_AUTHORIZED", issue: "notAuthorized" };
    }
    if (signal.canReconnect) {
      return {
        status: signal.status === "CONNECTED" ? "RECONNECT_REQUIRED" : signal.status,
        issue: "reconnect",
      };
    }
    return {
      status: signal.status === "CONNECTED" ? "DISCONNECTED" : signal.status,
      issue: "credentials",
    };
  }

  if (signal.status === "RECONNECT_REQUIRED") {
    return { status: signal.status, issue: "reconnect" };
  }
  if (signal.status === "NOT_AUTHORIZED") {
    return { status: signal.status, issue: "notAuthorized" };
  }
  if (signal.status !== "CONNECTED") {
    return { status: signal.status, issue: "credentials" };
  }

  return { status: "CONNECTED", issue: null };
}

// Account rows have no status of their own. The probe can still say connected
// when the credential cannot be used, which is the case the card presents as
// disconnected.
export function isConnectorConnected(
  connector: Parameters<typeof connectionSignalFrom>[0],
): boolean {
  const signal = connectionSignalFrom(connector);
  if (signal == null) {
    return false;
  }

  return presentConnection(signal).status === "CONNECTED";
}

function connectionIssueKeys(presented: readonly ConnectionPresentation[]): IssueKey[] {
  const present = new Set<IssueKey>();
  for (const item of presented) {
    if (item.issue != null) {
      present.add(item.issue);
    }
  }

  return (["reconnect", "notAuthorized", "credentials"] as const).filter(key => present.has(key));
}

function aggregateConnectionTone(presented: readonly ConnectionPresentation[]): ConnectionTone {
  const down = presented.filter(item => item.status !== "CONNECTED").length;
  if (down === 0) {
    return "green";
  }
  if (down === presented.length) {
    return "red";
  }
  return "amber";
}

function connectionTone(status: ConnectionStatus): ConnectionTone {
  if (status === "CONNECTED") {
    return "green";
  }
  if (status === "RECONNECT_REQUIRED" || status === "NOT_AUTHORIZED") {
    return "amber";
  }
  return "red";
}

interface ConnectorListItemProps {
  connectorKey: ConnectorListItem_connector$key;
  providerKey?: ConnectorProviderListItem_provider$key;
  organizationId: string;
  canConnect: boolean;
  showName?: boolean;
  showConnectorType?: boolean;
  showProbeError?: boolean;
  aggregatedConnectorIds?: readonly string[];
  connectionSignals?: readonly ConnectionSignal[];
  accountCount?: number;
  onSelect?: (connectorId: string) => void;
  onDeleted?: () => void;
}

function ConnectorNameHeading({
  connectorId,
  name,
  canUpdate,
}: {
  connectorId: string;
  name: string;
  canUpdate: boolean;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(name);
  const [error, setError] = useState<string | undefined>();
  const canceling = useRef(false);
  const saving = useRef(false);
  const [updateName, isUpdating] = useMutation<ConnectorListItemUpdateNameMutation>(
    updateConnectorNameMutation,
  );

  async function save(value: string) {
    if (saving.current) {
      return;
    }
    const trimmed = value.trim();
    if (trimmed === "") {
      setError(t("connectForm.name.required"));
      return;
    }
    if (trimmed === name) {
      setError(undefined);
      setEditing(false);
      return;
    }
    saving.current = true;
    try {
      await updateName({
        variables: { input: { connectorId, name: trimmed } },
      }, { errorToast: t("connectForm.name.saveFailed") });
      setError(undefined);
      setEditing(false);
    } catch {
      setDraft(trimmed);
    } finally {
      saving.current = false;
    }
  }

  if (!canUpdate) {
    return (
      <Heading level={2} size={3} weight="medium" highContrast className="truncate">
        {name}
      </Heading>
    );
  }

  if (!editing) {
    return (
      <button
        type="button"
        className="pointer-events-auto relative z-10 min-w-0 cursor-text text-left"
        onClick={(event) => {
          event.stopPropagation();
          setDraft(name);
          setError(undefined);
          setEditing(true);
        }}
      >
        <Heading level={2} size={3} weight="medium" highContrast className="truncate">
          {name}
        </Heading>
      </button>
    );
  }

  return (
    <div
      className="pointer-events-auto relative z-10 min-w-0"
      onClick={event => event.stopPropagation()}
      onKeyDown={event => event.stopPropagation()}
    >
      <Field error={error}>
        <TextField
          value={draft}
          autoFocus
          disabled={isUpdating}
          aria-label={t("connectForm.name.label")}
          onChange={event => setDraft(event.target.value)}
          onBlur={() => {
            if (canceling.current) {
              canceling.current = false;
              return;
            }
            void save(draft);
          }}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              void save(draft);
            }
            if (event.key === "Escape") {
              event.preventDefault();
              canceling.current = true;
              setDraft(name);
              setError(undefined);
              setEditing(false);
            }
          }}
        />
      </Field>
    </div>
  );
}

export function ConnectorListItem({
  connectorKey,
  providerKey,
  organizationId,
  canConnect,
  showName = true,
  showConnectorType = false,
  showProbeError = false,
  aggregatedConnectorIds,
  connectionSignals,
  accountCount,
  onSelect,
  onDeleted,
}: ConnectorListItemProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const connector = useFragment(connectorListItemFragment, connectorKey);
  const { card, controls, identity, metaRow, name, tags, title } = connectorCard();
  const aggregated = aggregatedConnectorIds != null;
  const showDelete = !aggregated && connector.canDelete;
  const ownSignal = connectionSignalFrom(connector);
  const signals = connectionSignals ?? (ownSignal == null ? [] : [ownSignal]);
  const presented = signals.map(presentConnection);
  const connectionIssues = connectionIssueKeys(presented);
  const solo = aggregated ? null : presented[0] ?? null;
  const tone = aggregated
    ? (presented.length === 0 ? "sand" : aggregateConnectionTone(presented))
    : (solo == null ? "sand" : connectionTone(solo.status));
  const connectedCount = presented.filter(item => item.status === "CONNECTED").length;
  const isDetailsCard = onSelect != null;
  const detailAccountCount = isDetailsCard ? connector.accounts?.totalCount : null;
  const listAccountCount = isDetailsCard
    ? null
    : (aggregated ? (accountCount ?? null) : (connector.accounts?.totalCount ?? null));
  const menu = (showDelete || canConnect)
    ? (
        <div className={controls({ className: "pointer-events-auto" })}>
          {canConnect && providerKey != null && (
            <ConnectorConnectMore
              providerKey={providerKey}
              organizationId={organizationId}
            />
          )}
          {showDelete && (
            <IconButton
              variant="ghost"
              color="red"
              size={1}
              aria-label={t("detailsPage.actions.delete")}
              onClick={() => setDeleteOpen(true)}
            >
              <TrashIcon />
            </IconButton>
          )}
        </div>
      )
    : undefined;

  return (
    <div className={card()}>
      {connector.canGet && onSelect != null && (
        <button
          type="button"
          aria-label={connector.name}
          className="absolute inset-0 z-0 cursor-pointer"
          onClick={() => onSelect(connector.id)}
        />
      )}
      {connector.canGet && onSelect == null && (
        <Link
          to={connectorDetailsPath(organizationId, connector.provider)}
          underline={false}
          aria-label={connector.displayName}
          className="absolute inset-0 z-0"
        />
      )}
      <TonedCard
        tone={tone}
        iconSize={14}
        className={connector.canGet ? "pointer-events-none relative z-0 h-full" : "relative z-0 h-full"}
        icon={showConnectorType
          ? (
              <ConnectorTypeMark
                protocol={connector.protocol}
                canReconnect={connector.canReconnect}
              />
            )
          : (
              <ThirdPartyLogo
                thirdParty={connector.provider}
                className="size-12"
              />
            )}
        lead={(
          <div className={identity()}>
            <div className={name()}>
              {onSelect != null
                ? (
                    <ConnectorNameHeading
                      connectorId={connector.id}
                      name={connector.name}
                      canUpdate={connector.canUpdate}
                    />
                  )
                : showName && (
                  <div className="flex min-w-0 flex-col">
                    <Heading level={2} size={3} weight="medium" highContrast className={title()}>
                      {connector.displayName}
                    </Heading>
                    {!aggregated && (
                      <Text size={1} color="faint" className="truncate">
                        {connector.name}
                      </Text>
                    )}
                  </div>
                )}
            </div>
          </div>
        )}
        control={menu}
      >
        {isDetailsCard && (
          <div className={tags()}>
            <ConnectorConnectionBadge
              aggregated={aggregated}
              aggregatedTotal={aggregatedConnectorIds?.length ?? null}
              connectedCount={connectedCount}
              tone={tone}
              solo={solo}
            />
            {detailAccountCount != null && (
              <Badge variant="soft" color="neutral" size={1}>
                {t("detailsPage.accountCount", { count: detailAccountCount })}
              </Badge>
            )}
          </div>
        )}
        {!isDetailsCard && (listAccountCount != null || aggregated || solo != null) && (
          <div className={tags()}>
            <ConnectorConnectionBadge
              aggregated={aggregated}
              aggregatedTotal={aggregatedConnectorIds?.length ?? null}
              connectedCount={connectedCount}
              tone={tone}
              solo={solo}
            />
            {listAccountCount != null && (
              <Badge variant="soft" color="neutral" size={1}>
                {t("listPage.accountCount", { count: listAccountCount })}
              </Badge>
            )}
          </div>
        )}
        {isDetailsCard && !aggregated && (
          <div className={metaRow()}>
            <div className="pointer-events-auto ml-auto">
              <ConnectorOrganizationSelect connectorKey={connector} />
            </div>
          </div>
        )}
        {!aggregated && connector.canReconnect && (
          <div className="pointer-events-auto">
            <ButtonAnchor
              href={buildConnectorInitiateURL(
                organizationId,
                connector.provider,
                connector.protocol,
                {
                  connectorId: connector.id,
                  oauth2Scopes: connector.oauth2Scopes,
                },
              )}
              variant="solid"
              size={1}
            >
              {t("detailsPage.actions.reconnect")}
            </ButtonAnchor>
          </div>
        )}
        <div className="pointer-events-auto">
          <Suspense fallback={<ConnectorModulesSkeleton />}>
            <ConnectorModules
              connectorIds={aggregatedConnectorIds ?? [connector.id]}
              organizationId={organizationId}
            />
          </Suspense>
        </div>
        {!aggregated && (
          <Text size={1} color="faint">
            {t("listPage.created", {
              date: dateFormat(i18n.language, connector.createdAt),
            })}
          </Text>
        )}
        {!aggregated && showProbeError && connectionIssues.length > 0 && (
          <ConnectorProbeError
            issues={connectionIssues}
            provider={connector.displayName}
          />
        )}
        {showDelete && (
          <ConnectorDeleteDialog
            connectorKey={connector}
            open={deleteOpen}
            onOpenChange={setDeleteOpen}
            onDeleted={onDeleted}
          />
        )}
      </TonedCard>
      <ConnectionIssueMark issues={showProbeError ? [] : connectionIssues} />
    </div>
  );
}

function ConnectorConnectionBadge({
  aggregated,
  aggregatedTotal,
  connectedCount,
  tone,
  solo,
}: {
  aggregated: boolean;
  aggregatedTotal: number | null;
  connectedCount: number;
  tone: ConnectionTone | "sand";
  solo: ConnectionPresentation | null;
}) {
  const { t } = useTranslation("organizations/settings/integrations");

  if (aggregated && aggregatedTotal != null) {
    return (
      <Badge
        variant="soft"
        color={tone === "sand" ? "neutral" : tone}
        size={1}
      >
        {t("listPage.connectedCount", {
          connected: connectedCount,
          total: aggregatedTotal,
        })}
      </Badge>
    );
  }

  if (!aggregated && solo != null) {
    return (
      <Badge
        variant="soft"
        color={connectionTone(solo.status)}
        size={1}
      >
        {t(`detailsPage.status.${solo.status}`)}
      </Badge>
    );
  }

  return null;
}

function ConnectorTypeMark({
  protocol,
  canReconnect,
}: {
  protocol: ConnectorListItem_connector$data["protocol"];
  canReconnect: boolean;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const method = connectMethodFromConnector(protocol, canReconnect);
  const label = t(`marketplacePage.methods.${method}`);

  return (
    <Tooltip>
      <TooltipTrigger
        aria-label={label}
        className="pointer-events-auto text-inherit"
      >
        <ConnectorMethodIcon method={method} className="size-8" />
      </TooltipTrigger>
      <TooltipPopup>{label}</TooltipPopup>
    </Tooltip>
  );
}

function ConnectorProbeError({
  issues,
  provider,
}: {
  issues: readonly IssueKey[];
  provider: string;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const { probeError, probeErrorHeader, probeErrorRow, probeErrorText } = connectorCard();

  async function copyError(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      return;
    }
    toast({
      title: t("detailsPage.probeError.copied"),
      description: t("detailsPage.probeError.copied"),
      variant: "success",
    });
  }

  return (
    <div className={probeError()}>
      <div className={probeErrorHeader()}>
        <span className="text-red-11" aria-hidden>
          <IconWarning size={16} className="shrink-0" />
        </span>
      </div>
      {issues.map((issue) => {
        const message = t(`listPage.connectionIssues.${issue}`, { provider });
        return (
          <div key={issue} className={probeErrorRow()}>
            <Text size={2} color="neutral" className={probeErrorText()}>
              {message}
            </Text>
            <IconButton
              size={1}
              variant="soft"
              color="neutral"
              aria-label={t("detailsPage.probeError.copy")}
              onClick={() => void copyError(message)}
            >
              <CopyIcon />
            </IconButton>
          </div>
        );
      })}
    </div>
  );
}

function ConnectionIssueMark({
  issues,
}: {
  issues: readonly IssueKey[];
}) {
  if (issues.length === 0) {
    return null;
  }

  return (
    <span
      className="pointer-events-none absolute right-3 bottom-3 z-1 text-red-11"
      aria-hidden
    >
      <IconWarning size={16} className="shrink-0" />
    </span>
  );
}
