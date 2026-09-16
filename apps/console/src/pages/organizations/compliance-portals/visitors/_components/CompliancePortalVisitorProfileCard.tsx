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

import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalVisitorProfileCard_access$key } from "#/__generated__/core/CompliancePortalVisitorProfileCard_access.graphql";

import { visitorDisplayName, visitorVisitStatus } from "../_lib/visitorIdentity";
import { visitorPage } from "../variants";

const fragment = graphql`
  fragment CompliancePortalVisitorProfileCard_access on CompliancePortalAccess {
    state
    authenticatedAt
    identity {
      fullName
      email
    }
  }
`;

interface CompliancePortalVisitorProfileCardProps {
  accessKey: CompliancePortalVisitorProfileCard_access$key;
  children?: ReactNode;
}

export function CompliancePortalVisitorProfileCard({
  accessKey,
  children,
}: CompliancePortalVisitorProfileCardProps) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const { profile, person, identity, joined, actions } = visitorPage();
  const access = useFragment(fragment, accessKey);
  const displayName = visitorDisplayName(
    access.identity.fullName,
    access.identity.email,
  );
  const visitStatus = visitorVisitStatus(access.state, access.authenticatedAt);

  return (
    <Card variant="ghost" size={2} padding="none" className={profile()}>
      <div className={person()}>
        <Avatar
          size={5}
          variant="soft"
          color="gold"
          fallback={displayName.charAt(0).toUpperCase() || "?"}
        />
        <div className={identity()}>
          <Heading level={2} size={4} weight="medium" highContrast className="truncate">
            {displayName}
          </Heading>
          <Text size={2} color="gold" className="truncate">
            {access.identity.email}
          </Text>
        </div>
      </div>
      <Separator />
      <div className={joined()}>
        <Text size={1} color="faint">
          {visitStatus === "deactivated"
            ? t("visitorPage.deactivated")
            : visitStatus === "notVisited"
              ? t("visitorPage.notVisited")
              : t("visitorPage.visitedOn", {
                  date: dateFormat(i18n.language, access.authenticatedAt),
                })}
        </Text>
      </div>
      {children
        ? (
            <>
              <Separator />
              <div className={actions()}>
                {children}
              </div>
            </>
          )
        : null}
    </Card>
  );
}
