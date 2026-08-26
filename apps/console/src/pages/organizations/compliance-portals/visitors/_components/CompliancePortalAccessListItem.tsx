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

import { FileTextIcon, SignatureIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalAccessListItemFragment$key } from "#/__generated__/core/CompliancePortalAccessListItemFragment.graphql";

import { ndaSignatureListKey, ndaSignatureTone } from "../_lib/ndaSignature";
import { accessListItem } from "../variants";

const fragment = graphql`
  fragment CompliancePortalAccessListItemFragment on CompliancePortalAccess {
    id
    createdAt
    profile {
      fullName
      emailAddress
    }
    pendingRequestCount
    ndaSignature {
      status
    }
  }
`;

interface CompliancePortalAccessListItemProps {
  accessKey: CompliancePortalAccessListItemFragment$key;
}

export function CompliancePortalAccessListItem({
  accessKey,
}: CompliancePortalAccessListItemProps) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const access = useFragment(fragment, accessKey);
  const {
    item,
    hit,
    body,
    avatar,
    main,
    identity,
    name,
    email,
    trailing,
    request,
    nda,
    joined,
  } = accessListItem();
  const ndaStatus = access.ndaSignature?.status;

  return (
    <ListItem className={item()}>
      <Link
        to={access.id}
        className={hit()}
        aria-label={t("accessListItem.actions.open")}
      />
      <div className={body()}>
        <Avatar
          size={3}
          variant="soft"
          color="gold"
          className={avatar()}
          fallback={access.profile.fullName.charAt(0).toUpperCase() || "?"}
        />
        <div className={main()}>
          <div className={identity()}>
            <ListItemContent>
              <Text size={2} weight="medium" color="neutral" highContrast className={name()}>
                {access.profile.fullName}
              </Text>
              <Text size={1} color="gold" className={email()}>
                {access.profile.emailAddress}
              </Text>
            </ListItemContent>
          </div>
          {ndaStatus != null && (
            <Text size={2} color={ndaSignatureTone(ndaStatus)} className={nda()}>
              <SignatureIcon aria-hidden />
              {t(ndaSignatureListKey(ndaStatus))}
            </Text>
          )}
        </div>
      </div>
      <div className={trailing()}>
        {access.pendingRequestCount > 0 && (
          <ButtonLink
            to={access.id}
            size={2}
            variant="soft"
            color="amber"
            className={request()}
            iconStart={<FileTextIcon aria-hidden />}
          >
            {t("accessListItem.requested", { count: access.pendingRequestCount })}
          </ButtonLink>
        )}
        <Text size={1} color="faint" className={joined()}>
          {t("visitorPage.joinedOn", { date: dateFormat(i18n.language, access.createdAt) })}
        </Text>
      </div>
    </ListItem>
  );
}
