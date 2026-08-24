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

import { SlackLogoIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
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

import { hostingCard } from "../../hosting/variants";
import { slackSection } from "../variants";

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
  const channelNames = Object.fromEntries(
    channels.map(channel => [channel.id, `#${channel.name}`]),
  );
  const showInstall = !isInstalled
    && compliancePortal.organization.canInstallSlackbot
    && compliancePortal.organization.slackbotAvailable;
  const showChannelConfig = isInstalled === true && compliancePortal.canConfigureSlack;

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

  const { root, intro, lead, copy, channel, channelRow, channelField, empty, emptyCopy, emptyCallout }
    = slackSection();
  const { frame, header, wash, fade, icon: iconSlot, control, body } = hostingCard({
    tone: "sand",
    wide: true,
  });

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("slackSection.title")}
        </Heading>
      </div>
      <Card variant="ghost" size={2} padding="none" className={frame()}>
        <div className={header()}>
          <div className={wash()} />
          <div className={fade()} />
          <div className={lead()}>
            <div className={iconSlot()}>
              <SlackLogoIcon size={24} weight="duotone" />
            </div>
            <div className={copy()}>
              <Text size={3} weight="medium" highContrast>
                {isInstalled
                  ? t("slackSection.connectionTitle")
                  : t("slackSection.notInstalledTitle")}
              </Text>
              <Text size={2} color="neutral">
                {isInstalled
                  ? t("slackSection.description")
                  : t("slackSection.notInstalled")}
              </Text>
            </div>
          </div>
          {showInstall && (
            <div className={control()}>
              <ButtonAnchor
                href={getSlackInstallUrl(organizationId)}
                variant="solid"
                color="neutral"
                highContrast
              >
                {t("slackSection.actions.install")}
              </ButtonAnchor>
            </div>
          )}
        </div>
        {showChannelConfig && (
          <div className={body()}>
            {hasChannels
              ? (
                  <Select
                    value={configuredChannel?.channelId ?? null}
                    onValueChange={(channelId) => {
                      if (channelId != null) {
                        setNotificationChannel(channelId);
                      }
                    }}
                    disabled={isSettingChannel || isClearingChannel}
                  >
                    <div className={channel()}>
                      <div className={channelRow()}>
                        <Field
                          label={t("slackSection.channel.label")}
                          className={channelField()}
                        >
                          <SelectTrigger placeholder={t("slackSection.channel.placeholder")}>
                            {(value: string | null) => (value ? channelNames[value] : null)}
                          </SelectTrigger>
                        </Field>
                        {configuredChannel && (
                          <Button
                            variant="surface"
                            color="neutral"
                            loading={isClearingChannel}
                            onClick={clearNotificationChannel}
                          >
                            {t("slackSection.actions.clear")}
                          </Button>
                        )}
                      </div>
                      <Text size={1} color="faint">
                        {t("slackSection.channel.help")}
                      </Text>
                      {listedChannels.length === 0 && (
                        <Button
                          variant="surface"
                          color="neutral"
                          onClick={refreshChannels}
                        >
                          {t("slackSection.actions.refresh")}
                        </Button>
                      )}
                    </div>
                    <SelectPopup>
                      {channels.map(listedChannel => (
                        <SelectItem key={listedChannel.id} value={listedChannel.id}>
                          {`#${listedChannel.name}`}
                        </SelectItem>
                      ))}
                    </SelectPopup>
                  </Select>
                )
              : (
                  <div className={empty()}>
                    <Callout variant="surface" color="neutral" className={emptyCallout()}>
                      <div className={emptyCopy()}>
                        <Text size={2} weight="medium" highContrast>
                          {t("slackSection.channel.emptyTitle")}
                        </Text>
                        <Text size={2}>
                          {t("slackSection.channel.emptyDescription")}
                        </Text>
                      </div>
                    </Callout>
                    <Button
                      variant="surface"
                      color="neutral"
                      onClick={refreshChannels}
                    >
                      {t("slackSection.actions.refresh")}
                    </Button>
                  </div>
                )}
            {loadMoreCursor && (
              <Button
                variant="surface"
                color="neutral"
                loading={isLoadingMore}
                onClick={loadMoreChannels}
              >
                {t("slackSection.actions.loadMore")}
              </Button>
            )}
          </div>
        )}
      </Card>
    </section>
  );
}

function getSlackInstallUrl(organizationId: string): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/slackbot/install/initiate", baseUrl);
  url.searchParams.set("organization_id", organizationId);
  url.searchParams.set("continue", window.location.pathname);
  return url.toString();
}
