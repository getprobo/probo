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

import { HashIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { SlackLogo } from "@probo/ui/src/v2/SlackLogo/SlackLogo";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useParams } from "react-router";

import type { SlackCard_organization$key } from "#/__generated__/core/SlackCard_organization.graphql";
import type { SlackCard_viewer$key } from "#/__generated__/core/SlackCard_viewer.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { dashboardCard, deviceCard } from "./variants";

const slackCardViewerFragment = graphql`
  fragment SlackCard_viewer on Viewer @throwOnFieldError {
    probotIdentityBindings {
      __typename
    }
  }
`;

const slackCardOrganizationFragment = graphql`
  fragment SlackCard_organization on Organization @throwOnFieldError {
    slackbotInstallation {
      active
    }
  }
`;

export interface SlackCardProps {
  viewerKey: SlackCard_viewer$key;
  organizationKey: SlackCard_organization$key;
}

export function SlackCard({ viewerKey, organizationKey }: SlackCardProps) {
  const { t } = useTranslation();
  const { organizationId } = useParams();
  const slots = dashboardCard();
  const status = deviceCard();
  const viewer = useFragment(slackCardViewerFragment, viewerKey);
  const organization = useFragment(slackCardOrganizationFragment, organizationKey);

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  if (organization.slackbotInstallation?.active !== true) {
    return null;
  }

  const linked = viewer.probotIdentityBindings.length > 0;

  return (
    <Card variant="soft" padding="none" size={3} className={slots.frame()}>
      <div className={slots.header()}>
        <SlackLogo className={slots.icon()} />
        <div className={slots.copy()}>
          <Heading level={2} size={2} weight="medium" highContrast>
            {t("homePage.dashboard.slack.title")}
          </Heading>
          <Text size={1} color="current" className={slots.description()}>
            {t("homePage.dashboard.slack.description")}
          </Text>
        </div>
        <Link
          to={`/${organizationId}/bindings`}
          size={2}
          color="neutral"
          underline={false}
          className={slots.view()}
        >
          {t("homePage.dashboard.view")}
        </Link>
      </div>
      <div className={slots.body()}>
        {linked
          ? (
              <div className={slots.empty()}>
                <span className={status.status()}>
                  <span className={status.ring()} aria-hidden />
                  <span className={status.pip()} />
                </span>
                <Text size={2} weight="medium" color="current" className={slots.emptyLabel()}>
                  {t("homePage.dashboard.slack.connected")}
                </Text>
              </div>
            )
          : (
              <div className={slots.empty()}>
                <HashIcon className={slots.emptyIcon()} />
                <Text size={2} weight="medium" color="current" className={slots.emptyLabel()}>
                  <Trans
                    i18nKey="homePage.dashboard.slack.hint"
                    components={{ cmd: <Code size={2} /> }}
                  />
                </Text>
              </div>
            )}
      </div>
    </Card>
  );
}
