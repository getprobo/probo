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

import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { GoogleLogo } from "@probo/ui/src/v2/GoogleLogo/GoogleLogo";
import { MicrosoftLogo } from "@probo/ui/src/v2/MicrosoftLogo/MicrosoftLogo";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SCIMProviderCard_scimConfiguration$key } from "#/__generated__/iam/SCIMProviderCard_scimConfiguration.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import {
  bridgeStatus,
  bridgeStatusBadgeColor,
  bridgeStatusTone,
  scimProviderCopy,
} from "../_lib/scimProvider";
import { scimProviderCard } from "../variants";

import { DisconnectSCIMProviderDialog } from "./DisconnectSCIMProviderDialog";
import { ReactivateSCIMBridgeButton } from "./ReactivateSCIMBridgeButton";
import { SCIMProviderSettingsDrawer } from "./SCIMProviderSettingsDrawer";

const scimProviderCardFragment = graphql`
  fragment SCIMProviderCard_scimConfiguration on SCIMConfiguration {
    canDelete: permission(action: "iam:scim-configuration:delete")
    ...DisconnectSCIMProviderDialog_scimConfiguration
    bridge {
      type
      state
      syncError
      excludedUserNames
      createdAt
      canUpdate: permission(action: "iam:scim-bridge:update")
      connector {
        createdAt
      }
      ...ReactivateSCIMBridgeButtonFragment
      ...SCIMProviderSettingsDrawer_scimBridge
    }
  }
`;

export interface SCIMProviderCardProps {
  scimConfigurationKey: SCIMProviderCard_scimConfiguration$key;
}

export function SCIMProviderCard({ scimConfigurationKey }: SCIMProviderCardProps) {
  const data = useFragment(scimProviderCardFragment, scimConfigurationKey);
  const bridge = data.bridge;
  const { t, i18n } = useTranslation();
  const { actions, body, errorCopy, hint } = scimProviderCard();

  if (bridge == null) {
    return null;
  }

  const copy = scimProviderCopy(bridge.type);
  const status = bridgeStatus(bridge.state);
  const tone = bridgeStatusTone[status];
  const hasError = status === "error" || status === "disabled";
  const connectedAt = bridge.connector?.createdAt ?? bridge.createdAt;

  return (
    <TonedCard
      tone={tone}
      icon={bridge.type === "MICROSOFT_365"
        ? <MicrosoftLogo className="size-6" />
        : <GoogleLogo className="size-6" />}
      lead={(
        <Text size={3} weight="medium" color={tone === "sand" ? "neutral" : tone}>
          {t(`${copy}.name`)}
        </Text>
      )}
      control={(
        <Badge size={2} variant="soft" color={bridgeStatusBadgeColor[status]}>
          {t(`${copy}.status.${status}`)}
        </Badge>
      )}
    >
      <div className={body()}>
        <Text size={2} color="neutral">
          {t(`${copy}.connectedOn`, {
            date: dateTimeFormat(i18n.language, connectedAt),
          })}
        </Text>
        {hasError && (
          <>
            <Callout color={status === "disabled" ? "neutral" : "red"}>
              <div className={errorCopy()}>
                <Text size={2} weight="medium" highContrast>
                  {status === "disabled"
                    ? t(`${copy}.errors.bridgeDisabled`)
                    : t(`${copy}.errors.bridgeFailed`)}
                </Text>
                <Text size={2}>
                  {bridge.syncError ?? t(`${copy}.errors.bridgeSync`)}
                </Text>
              </div>
            </Callout>
            <ReactivateSCIMBridgeButton bridgeKey={bridge} />
          </>
        )}
        <div className={actions()}>
          <Text size={2} color="neutral" className={hint()}>
            {t(`${copy}.excludedCount`, { count: bridge.excludedUserNames.length })}
          </Text>
          {bridge.canUpdate && (
            <SCIMProviderSettingsDrawer
              scimBridgeKey={bridge}
              copy={copy}
            />
          )}
          {data.canDelete && (
            <DisconnectSCIMProviderDialog
              scimConfigurationKey={data}
              copy={copy}
            />
          )}
        </div>
      </div>
    </TonedCard>
  );
}
