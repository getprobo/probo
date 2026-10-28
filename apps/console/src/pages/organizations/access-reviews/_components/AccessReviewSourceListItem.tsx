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

import { DotsThreeVerticalIcon, FileCsvIcon, TrashIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { ThirdPartyLogo } from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  AccessReviewSourceListItem_source$data,
  AccessReviewSourceListItem_source$key,
} from "#/__generated__/core/AccessReviewSourceListItem_source.graphql";
import type {
  AccessReviewSourceConnectionStatus,
  AccessReviewSourceListItem_status$key,
} from "#/__generated__/core/AccessReviewSourceListItem_status.graphql";
import type { AccessReviewSourceListItemConfigureMutation } from "#/__generated__/core/AccessReviewSourceListItemConfigureMutation.graphql";
import type { InlineOrgSelectQuery } from "#/__generated__/core/InlineOrgSelectQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { buildConnectorInitiateURL } from "#/pages/organizations/settings/integrations/_lib/connectorInitiate";

import { connectionIssue } from "../_lib/connectionIssue";
import { accessReviewSourceLabel } from "../_lib/sourceLabel";
import { sourceListItem } from "../sources/_components/variants";

import { AccessReviewSourceDeleteDialog } from "./AccessReviewSourceDeleteDialog";
import { InlineOrgSelect, inlineOrgSelectQuery } from "./InlineOrgSelect";
import { SourceConnectionIssue } from "./SourceConnectionIssue";

const fragment = graphql`
  fragment AccessReviewSourceListItem_source on AccessReviewSource {
    id
    name
    connectorAccount {
      name
    }
    connectorId
    connector {
      provider
      protocol
      canReconnect
      displayName
      oauth2Scopes
      ...SourceConnectionIssue_connector
    }
    selectedOrganization
    needsConfiguration
    createdAt
    canDelete: permission(action: "access-review:source:delete")
    ...AccessReviewSourceDeleteDialog_source
  }
`;

const statusFragment = graphql`
  fragment AccessReviewSourceListItem_status on AccessReviewSource {
    connectionStatus
  }
`;

const configureMutation = graphql`
  mutation AccessReviewSourceListItemConfigureMutation(
    $input: ConfigureAccessReviewSourceInput!
  ) {
    configureAccessReviewSource(input: $input) {
      accessReviewSource {
        id
        selectedOrganization
        needsConfiguration
      }
    }
  }
`;

interface AccessReviewSourceListItemProps {
  sourceKey: AccessReviewSourceListItem_source$key;
  statusKey: AccessReviewSourceListItem_status$key;
  connectionId: string;
}

export function AccessReviewSourceListItem({
  sourceKey,
  statusKey,
  connectionId,
}: AccessReviewSourceListItemProps) {
  const accessSource = useFragment(fragment, sourceKey);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const row = {
    accessSource,
    connectionId,
    deleteOpen,
    onDeleteOpenChange: setDeleteOpen,
  };

  // A CSV source has no connector, so its status is known without a probe.
  // A vendor row waits: treating a missing status as NOT_APPLICABLE would
  // flash the plain card before the probe says otherwise.
  if (accessSource.connector == null) {
    return (
      <AccessReviewSourceRow
        {...row}
        connectionStatus="NOT_APPLICABLE"
      />
    );
  }

  return (
    <Suspense
      fallback={(
        <AccessReviewSourceRow
          {...row}
          connectionStatus={null}
        />
      )}
    >
      <DeferredAccessReviewSourceRow
        {...row}
        statusKey={statusKey}
      />
    </Suspense>
  );
}

interface AccessReviewSourceRowProps {
  accessSource: AccessReviewSourceListItem_source$data;
  connectionId: string;
  deleteOpen: boolean;
  onDeleteOpenChange: (open: boolean) => void;
  connectionStatus: AccessReviewSourceConnectionStatus | null;
}

function DeferredAccessReviewSourceRow({
  statusKey,
  ...row
}: Omit<AccessReviewSourceRowProps, "connectionStatus"> & {
  statusKey: AccessReviewSourceListItem_status$key;
}) {
  const status = useFragment(statusFragment, statusKey);

  return (
    <AccessReviewSourceRow
      {...row}
      connectionStatus={status.connectionStatus}
    />
  );
}

function AccessReviewSourceRow({
  accessSource,
  connectionId,
  deleteOpen,
  onDeleteOpenChange,
  connectionStatus,
}: AccessReviewSourceRowProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();

  const [configure]
    = useMutation<AccessReviewSourceListItemConfigureMutation>(configureMutation, {
      successMessage: t("accessReviewSourceRow.messages.organizationUpdated"),
      errorToast: t("accessReviewSourceRow.errors.configure"),
    });
  const [orgsQueryRef, loadOrgsQuery]
    = useQueryLoader<InlineOrgSelectQuery>(inlineOrgSelectQuery);

  const connector = accessSource.connector;
  const showOrgSelector = connector != null
    && connectionStatus != null
    && connectionStatus !== "NOT_APPLICABLE"
    && Boolean(accessSource.needsConfiguration || accessSource.selectedOrganization);

  useEffect(() => {
    if (!showOrgSelector) {
      return;
    }

    loadOrgsQuery(
      { accessReviewSourceId: accessSource.id },
      { fetchPolicy: "store-or-network" },
    );
  }, [showOrgSelector, loadOrgsQuery, accessSource.id]);

  const readyOrgsQueryRef = orgsQueryRef != null
    && orgsQueryRef.variables.accessReviewSourceId === accessSource.id
    ? orgsQueryRef
    : null;

  const handleOrgChange = (slug: string) => {
    void configure({
      variables: {
        input: {
          accessReviewSourceId: accessSource.id,
          organizationSlug: slug,
        },
      },
    }).catch(() => undefined);
  };

  const {
    card,
    header,
    identity,
    title,
    logo,
    actions,
  } = sourceListItem();

  const deleteControl = accessSource.canDelete
    ? (
        <>
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
                color="error"
                iconStart={<TrashIcon />}
                onClick={() => onDeleteOpenChange(true)}
              >
                {t("accessReviewSourceRow.actions.delete")}
              </DropdownItem>
            </DropdownPopup>
          </Dropdown>
          <AccessReviewSourceDeleteDialog
            sourceKey={accessSource}
            connectionId={connectionId}
            open={deleteOpen}
            onOpenChange={onDeleteOpenChange}
          />
        </>
      )
    : null;

  if (connector == null || connectionStatus === "NOT_APPLICABLE") {
    return (
      <Card variant="soft" size={2} className={card()}>
        <div className={header()}>
          {connector == null && (
            <FileCsvIcon className={logo()} aria-hidden />
          )}
          <div className={identity()}>
            <Heading level={3} size={3} weight="medium" highContrast className={title()}>
              {accessSource.name}
            </Heading>
            <Text size={1} color="faint">
              <time dateTime={accessSource.createdAt}>
                {dateTimeFormat(i18n.language, accessSource.createdAt)}
              </time>
            </Text>
          </div>
        </div>
        {deleteControl != null && (
          <div className={actions()}>
            {deleteControl}
          </div>
        )}
      </Card>
    );
  }

  if (connectionStatus == null) {
    return (
      <Card variant="soft" size={2} className={card()}>
        <div className={header()}>
          <ThirdPartyLogo
            thirdParty={connector.provider}
            className={logo()}
          />
          <div className={identity()}>
            <Heading level={3} size={3} weight="medium" highContrast className={title()}>
              {accessReviewSourceLabel({
                name: accessSource.name,
                accountName: accessSource.connectorAccount?.name,
              })}
            </Heading>
            <Text size={1} color="faint">
              <time dateTime={accessSource.createdAt}>
                {dateTimeFormat(i18n.language, accessSource.createdAt)}
              </time>
            </Text>
          </div>
        </div>
        {deleteControl != null && (
          <div className={actions()}>
            {deleteControl}
          </div>
        )}
      </Card>
    );
  }

  const canReconnect = connector.canReconnect;
  const reconnectUrl = accessSource.connectorId == null
    ? null
    : buildConnectorInitiateURL(
        organizationId,
        connector.provider,
        connector.protocol,
        {
          connectorId: accessSource.connectorId,
          oauth2Scopes: connector.oauth2Scopes,
        },
      );
  // Stated as what a healthy source is, not as a list of the ways it can
  // break: a status added to the enum after this bundle shipped would match
  // no branch of such a list and silently render nothing at all for a source
  // that is in fact broken. An unrecognised one is treated as DISCONNECTED is.
  const hasConnectionIssue = connectionStatus !== "CONNECTED";
  const showStandaloneIssue = hasConnectionIssue && !showOrgSelector;
  const standaloneIssue = connectionIssue(
    connectionStatus,
    canReconnect,
    "NOT_APPLICABLE",
  );

  return (
    <Card variant="soft" size={2} className={card()}>
      <div className={header()}>
        <ThirdPartyLogo
          thirdParty={connector.provider}
          className={logo()}
        />
        <div className={identity()}>
          <Heading level={3} size={3} weight="medium" highContrast className={title()}>
            {accessReviewSourceLabel({
              name: accessSource.name,
              accountName: accessSource.connectorAccount?.name,
            })}
          </Heading>
          <Text size={1} color="faint">
            <time dateTime={accessSource.createdAt}>
              {dateTimeFormat(i18n.language, accessSource.createdAt)}
            </time>
          </Text>
        </div>
      </div>

      <div className={actions()}>
        {showOrgSelector && (
          readyOrgsQueryRef == null
            ? <OrgSelectPending />
            : (
                <Suspense fallback={<OrgSelectPending />}>
                  <InlineOrgSelect
                    queryRef={readyOrgsQueryRef}
                    connectionStatus={connectionStatus}
                    canReconnect={canReconnect}
                    reconnectUrl={canReconnect ? reconnectUrl : null}
                    connectorKey={connector}
                    providerName={connector.displayName}
                    onSelect={handleOrgChange}
                  />
                </Suspense>
              )
        )}
        {showStandaloneIssue && standaloneIssue != null && (
          <SourceConnectionIssue
            connectorKey={connector}
            issueKey={standaloneIssue}
            reconnectUrl={canReconnect ? reconnectUrl : null}
          />
        )}
        {deleteControl}
      </div>
    </Card>
  );
}

function OrgSelectPending() {
  const { t } = useTranslation();
  const { organizationSelect } = sourceListItem();

  return (
    <div className={organizationSelect()}>
      <Select disabled>
        <SelectTrigger
          size={1}
          placeholder={t("accessReviewSourceRow.loading")}
          aria-label={t("accessReviewSourceRow.loading")}
        />
      </Select>
    </div>
  );
}
