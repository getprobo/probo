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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalVisitorPageQuery } from "#/__generated__/core/CompliancePortalVisitorPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { CompliancePortalDocumentAccessList } from "./_components/CompliancePortalDocumentAccessList";
import { ElectronicSignatureSection } from "./_components/ElectronicSignatureSection";
import { NdaSignatureBadge } from "./_components/NdaSignatureBadge";
import { visitorPage } from "./variants";

export const compliancePortalVisitorPageQuery = graphql`
  query CompliancePortalVisitorPageQuery($accessId: ID!) {
    node(id: $accessId) {
      __typename
      ... on CompliancePortalAccess {
        createdAt
        canGet: permission(action: "compliance-portal:portal-access:get")
        canUpdate: permission(action: "compliance-portal:portal-access:update")
        profile {
          fullName
          emailAddress
          state
        }
        ndaSignature {
          status
          ...ElectronicSignatureSectionFragment
        }
        ...CompliancePortalDocumentAccessList_access
      }
    }
  }
`;

interface CompliancePortalVisitorPageProps {
  queryRef: PreloadedQuery<CompliancePortalVisitorPageQuery>;
}

export function CompliancePortalVisitorPage({ queryRef }: CompliancePortalVisitorPageProps) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const { root, back, header, identity, meta } = visitorPage();
  const { node } = usePreloadedQuery<CompliancePortalVisitorPageQuery>(
    compliancePortalVisitorPageQuery,
    queryRef,
  );
  if (node?.__typename !== "CompliancePortalAccess" || !node.canGet) {
    throw new NotFoundError("Visitor not found");
  }

  const canUpdate = node.canUpdate && node.profile.state === "ACTIVE";
  usePageTitle(node.profile.fullName);

  return (
    <div className={root()}>
      <Link to=".." size={2} color="neutral" underline={false} iconStart={<CaretLeftIcon />} className={back()}>
        {t("visitorPage.back")}
      </Link>
      <div className={header()}>
        <Avatar
          size={5}
          variant="soft"
          color="gold"
          fallback={node.profile.fullName.charAt(0).toUpperCase() || "?"}
        />
        <div className={identity()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {node.profile.fullName}
          </Heading>
          <Text size={2} color="gold">
            {node.profile.emailAddress}
          </Text>
          <div className={meta()}>
            <Text size={1} color="faint">
              {t("visitorPage.joinedOn", { date: dateFormat(i18n.language, node.createdAt) })}
            </Text>
            {node.ndaSignature != null && (
              <NdaSignatureBadge status={node.ndaSignature.status} />
            )}
          </div>
        </div>
      </div>
      {node.ndaSignature != null && (
        <ElectronicSignatureSection fragmentRef={node.ndaSignature} />
      )}
      <CompliancePortalDocumentAccessList
        accessKey={node}
        canUpdate={canUpdate}
      />
    </div>
  );
}
