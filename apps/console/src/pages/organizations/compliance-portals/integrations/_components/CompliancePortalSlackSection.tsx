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

import { Button, Card, Option, Select, Slack } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, useRefetchableFragment, useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalSlackSection_compliancePortal$key } from "#/__generated__/core/CompliancePortalSlackSection_compliancePortal.graphql";
import type { CompliancePortalSlackSectionChannelsQuery } from "#/__generated__/core/CompliancePortalSlackSectionChannelsQuery.graphql";
import type { CompliancePortalSlackSectionClearMutation } from "#/__generated__/core/CompliancePortalSlackSectionClearMutation.graphql";
import type { CompliancePortalSlackSectionRefetchQuery } from "#/__generated__/core/CompliancePortalSlackSectionRefetchQuery.graphql";
import type { CompliancePortalSlackSectionSetMutation } from "#/__generated__/core/CompliancePortalSlackSectionSetMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const fragment = graphql`
  fragment CompliancePortalSlackSection_compliancePortal on CompliancePortal
  @refetchable(queryName: "CompliancePortalSlackSectionRefetchQuery") {
    id
    canConfigureSlack: permission(action: "core:connector:initiate")
    slackbotNotificationChannel {
      channelId
      channelName
    }
    organization {
      id
      canInstallSlackbot: permission(action: "core:connector:initiate")
      slackbotAvailable
      slackbotInstallation {
        active
      }
      slackbotChannels {
        channels {
          id
          name
        }
        nextCursor
      }
    }
  }
`;

const loadMoreChannelsQuery = graphql`
  query CompliancePortalSlackSectionChannelsQuery(
    $organizationId: ID!
    $cursor: String
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      ... on Organization {
        slackbotChannels(cursor: $cursor) {
          channels {
            id
            name
          }
          nextCursor
        }
      }
    }
  }
`;

const setMutation = graphql`
  mutation CompliancePortalSlackSectionSetMutation(
    $input: SetSlackbotNotificationChannelInput!
  ) {
    setSlackbotNotificationChannel(input: $input) {
      channel {
        channelId
        channelName
      }
    }
  }
`;

const clearMutation = graphql`
  mutation CompliancePortalSlackSectionClearMutation(
    $input: ClearSlackbotNotificationChannelInput!
  ) {
    clearSlackbotNotificationChannel(input: $input) {
      channelId
    }
  }
`;

interface CompliancePortalSlackSectionProps {
  compliancePortalKey: CompliancePortalSlackSection_compliancePortal$key;
}

export function CompliancePortalSlackSection({
  compliancePortalKey,
}: CompliancePortalSlackSectionProps) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");
  const [compliancePortal, refetch] = useRefetchableFragment<
    CompliancePortalSlackSectionRefetchQuery,
    CompliancePortalSlackSection_compliancePortal$key
  >(fragment, compliancePortalKey);
  const [setChannel, isSettingChannel]
    = useMutation<CompliancePortalSlackSectionSetMutation>(setMutation);
  const [clearChannel, isClearingChannel]
    = useMutation<CompliancePortalSlackSectionClearMutation>(clearMutation);

  const setNotificationChannel = (channelId: string) => {
    void setChannel({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          channelId,
        },
      },
      updater(store) {
        const channel = store
          .getRootField("setSlackbotNotificationChannel")
          ?.getLinkedRecord("channel");
        const portal = store.get(compliancePortal.id);
        if (channel && portal) {
          portal.setLinkedRecord(channel, "slackbotNotificationChannel");
        }
      },
    });
  };

  const clearNotificationChannel = () => {
    void clearChannel({
      variables: {
        input: { compliancePortalId: compliancePortal.id },
      },
      updater(store) {
        store
          .get(compliancePortal.id)
          ?.setValue(null, "slackbotNotificationChannel");
      },
    });
  };

  const environment = useRelayEnvironment();
  const [extraChannels, setExtraChannels] = useState<
    { id: string; name: string }[]
  >([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const isInstalled = compliancePortal.organization.slackbotInstallation?.active;
  const firstPage = compliancePortal.organization.slackbotChannels;
  const listedChannels = extraChannels.length > 0
    ? [...firstPage.channels, ...extraChannels]
    : firstPage.channels;
  const configuredChannel = compliancePortal.slackbotNotificationChannel;
  const channels = configuredChannel
    && !listedChannels.some(channel => channel.id === configuredChannel.channelId)
    ? [
        {
          id: configuredChannel.channelId,
          name: configuredChannel.channelName,
        },
        ...listedChannels,
      ]
    : listedChannels;
  const hasChannels = channels.length > 0;
  const loadMoreCursor = extraChannels.length > 0
    ? nextCursor
    : firstPage.nextCursor;

  const refreshChannels = () => {
    setExtraChannels([]);
    setNextCursor(null);
    refetch({}, { fetchPolicy: "network-only" });
  };

  const loadMoreChannels = () => {
    if (!loadMoreCursor || isLoadingMore) {
      return;
    }

    setIsLoadingMore(true);
    fetchQuery<CompliancePortalSlackSectionChannelsQuery>(
      environment,
      loadMoreChannelsQuery,
      {
        organizationId: compliancePortal.organization.id,
        cursor: loadMoreCursor,
      },
      { fetchPolicy: "network-only" },
    ).subscribe({
      next(data) {
        const page = data.organization.slackbotChannels;
        if (!page) {
          return;
        }

        setExtraChannels(current => [...current, ...page.channels]);
        setNextCursor(page.nextCursor ?? null);
      },
      complete() {
        setIsLoadingMore(false);
      },
      error() {
        setIsLoadingMore(false);
      },
    });
  };

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{t("slackSection.title")}</h2>
      <Card padded className="space-y-4">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
            <Slack className="h-6 w-6" />
          </div>
          <div className="mr-auto">
            <h3 className="text-base font-semibold">
              {isInstalled
                ? t("slackSection.connectionTitle")
                : t("slackSection.notInstalledTitle")}
            </h3>
            <p className="text-sm text-txt-tertiary">
              {isInstalled
                ? t("slackSection.description")
                : t("slackSection.notInstalled")}
            </p>
          </div>
          {!isInstalled
            && compliancePortal.organization.canInstallSlackbot
            && compliancePortal.organization.slackbotAvailable && (
            <Button variant="secondary" asChild>
              <a href={getSlackInstallUrl(organizationId)}>
                {t("slackSection.actions.install")}
              </a>
            </Button>
          )}
        </div>

        {isInstalled && compliancePortal.canConfigureSlack && (
          <div className="space-y-1.5">
            {hasChannels
              ? (
                  <>
                    <label className="text-sm font-medium">
                      {t("slackSection.channel.label")}
                    </label>
                    <div className="flex items-center gap-2">
                      <div className="grow">
                        <Select
                          value={
                            compliancePortal.slackbotNotificationChannel
                              ?.channelId ?? ""
                          }
                          onValueChange={setNotificationChannel}
                          placeholder={t("slackSection.channel.placeholder")}
                          disabled={isSettingChannel || isClearingChannel}
                        >
                          {channels.map(channel => (
                            <Option key={channel.id} value={channel.id}>
                              #
                              {channel.name}
                            </Option>
                          ))}
                        </Select>
                      </div>
                      {compliancePortal.slackbotNotificationChannel && (
                        <Button
                          variant="secondary"
                          onClick={clearNotificationChannel}
                          disabled={isClearingChannel}
                        >
                          {t("slackSection.actions.clear")}
                        </Button>
                      )}
                    </div>
                    <p className="text-xs text-txt-tertiary">
                      {t("slackSection.channel.help")}
                    </p>
                    {listedChannels.length === 0 && (
                      <Button
                        variant="secondary"
                        onClick={refreshChannels}
                      >
                        {t("slackSection.actions.refresh")}
                      </Button>
                    )}
                  </>
                )
              : (
                  <div className="flex items-center gap-3 rounded-lg border border-border-low bg-level-1 px-4 py-3">
                    <div className="grow">
                      <p className="text-sm font-medium">
                        {t("slackSection.channel.emptyTitle")}
                      </p>
                      <p className="text-sm text-txt-tertiary">
                        {t("slackSection.channel.emptyDescription")}
                      </p>
                    </div>
                    <Button
                      variant="secondary"
                      onClick={refreshChannels}
                    >
                      {t("slackSection.actions.refresh")}
                    </Button>
                  </div>
                )}
            {loadMoreCursor && (
              <Button
                variant="secondary"
                onClick={loadMoreChannels}
                disabled={isLoadingMore}
              >
                {isLoadingMore
                  ? t("slackSection.actions.loadingMore")
                  : t("slackSection.actions.loadMore")}
              </Button>
            )}
          </div>
        )}
      </Card>
    </div>
  );
}

function getSlackInstallUrl(organizationId: string): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/slackbot/install/initiate", baseUrl);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.set("continue", window.location.pathname);
  return url.toString();
}
