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

import { DotsThreeVerticalIcon, PencilSimpleIcon, PlusIcon } from "@phosphor-icons/react";
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { AddableConnectorListItem_connector$key } from "#/__generated__/core/AddableConnectorListItem_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";
import { connectorDetailsPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";

import {
  createAccessReviewSourcesMutation,
  prependCreatedSourceEdges,
} from "../../dialogs/accessReviewSourceMutations";
import { listedConnectorAccounts } from "../_lib/listedConnectorAccounts";

const fragment = graphql`
  fragment AddableConnectorListItem_connector on Connector @relay(plural: true) {
    id
    displayName
    provider
    connectionStatus
    accounts(first: 50) {
      totalCount
      edges {
        node {
          id
          name
          externalAccountId
        }
      }
    }
  }
`;

interface AddableConnectorListItemProps {
  connectorKeys: AddableConnectorListItem_connector$key;
  organizationId: string;
  connectionId: string;
  normalizedSearch: string;
}

function accountNeedsOrganization(externalAccountId: string, connectorId: string) {
  return externalAccountId === connectorId;
}

function connectorTone(status: string): "green" | "amber" | "red" {
  if (status === "CONNECTED") {
    return "green";
  }
  if (status === "RECONNECT_REQUIRED" || status === "NOT_AUTHORIZED") {
    return "amber";
  }
  return "red";
}

function vendorTone(statuses: readonly string[]): "green" | "amber" | "red" {
  const [first] = statuses;
  if (first != null && statuses.every(status => status === first)) {
    return connectorTone(first);
  }

  const down = statuses.filter(status => status !== "CONNECTED").length;
  if (down === 0) {
    return "green";
  }
  if (down === statuses.length) {
    return "red";
  }
  return "amber";
}

export function AddableConnectorListItem({
  connectorKeys,
  organizationId,
  connectionId,
  normalizedSearch,
}: AddableConnectorListItemProps) {
  const { t } = useTranslation();
  const connectors = useFragment(fragment, connectorKeys);
  const [isAdding, setIsAdding] = useState(false);
  const [createAccessReviewSources, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourcesMutation,
    );
  const busy = isAdding || isCreating;
  const [face] = connectors;
  const statuses = connectors.map(connector => connector.connectionStatus);
  const sharedStatus = statuses.every(status => status === statuses[0])
    ? statuses[0]
    : null;
  const tone = vendorTone(statuses);
  const accountCount = connectors.length === 1
    ? connectors[0].accounts.totalCount
    : new Set(
      connectors.flatMap(connector =>
        connector.accounts.edges.map(({ node }) => node.externalAccountId),
      ),
    ).size;
  const listed = connectors.flatMap(connector =>
    (listedConnectorAccounts(
      connector.accounts.edges.map(({ node }) => node),
      connector,
      normalizedSearch,
    ) ?? []).map(account => ({
      id: account.id,
      name: account.name,
      needsOrganization: accountNeedsOrganization(account.externalAccountId, connector.id),
    })),
  );
  const addable = listed.filter(account => !account.needsOrganization);

  async function addSources(accounts: { id: string; name: string }[]) {
    if (busy || accounts.length === 0) {
      return;
    }

    setIsAdding(true);
    try {
      await createAccessReviewSources({
        variables: {
          input: {
            organizationId,
            sources: accounts.map(account => ({
              connectorAccountId: account.id,
              name: account.name,
              connectorId: null,
              csvData: null,
            })),
          },
        },
        updater: (store) => {
          prependCreatedSourceEdges(store, connectionId);
        },
      }, {
        successMessage: t("accessReviewSourcesPage.messages.created"),
        errorToast: t("accessReviewSourcesPage.errors.create"),
      });
    } catch {
      // The mutation hook already reported the error.
    } finally {
      setIsAdding(false);
    }
  }

  if (face == null || listed.length === 0) {
    return null;
  }

  return (
    <TonedCard
      tone={tone}
      iconSize={14}
      icon={(
        <ThirdPartyLogo
          thirdParty={face.provider}
          className="size-12"
        />
      )}
      lead={(
        <Heading level={2} size={3} weight="medium" highContrast className="min-w-0 truncate">
          {face.displayName}
        </Heading>
      )}
      control={(
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
              iconStart={<PlusIcon />}
              disabled={busy || addable.length === 0}
              onClick={() => {
                void addSources(addable);
              }}
            >
              {t("accessReviewSourcesPage.actions.addAll")}
            </DropdownItem>
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
      )}
    >
      {addable.length === 0 && (
        <Text size={2} color="faint">
          {t("accessReviewSourcesPage.needsOrganization")}
        </Text>
      )}
      <div className="mt-auto flex flex-wrap items-center gap-2">
        <Badge variant="soft" color={tone} size={1}>
          {sharedStatus != null
            ? t(`accessReviewSourcesPage.status.${sharedStatus}`)
            : t("accessReviewSourcesPage.connectedCount", {
                connected: statuses.filter(status => status === "CONNECTED").length,
                total: statuses.length,
              })}
        </Badge>
        <Badge variant="soft" color="neutral" size={1}>
          {t("accessReviewSourcesPage.accountCount", {
            count: accountCount,
          })}
        </Badge>
      </div>
    </TonedCard>
  );
}
