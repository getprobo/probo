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
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { SAMLConfigurationPageQuery } from "#/__generated__/iam/SAMLConfigurationPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { EditSAMLConfigurationForm } from "./_components/EditSAMLConfigurationForm";
import { SAMLConfigurationDnsRecord } from "./_components/SAMLConfigurationDnsRecord";
import { showsSamlLoginUrl } from "./_lib/samlConfigurationCardTone";
import { newSamlSsoPage } from "./variants";

export const samlConfigurationPageQuery = graphql`
  query SAMLConfigurationPageQuery($samlConfigurationId: ID!) {
    samlConfiguration: node(id: $samlConfigurationId) @required(action: THROW) {
      __typename
      ... on SAMLConfiguration {
        emailDomain
        enforcementPolicy
        domainVerificationToken
        domainVerifiedAt
        testLoginUrl
        canGet: permission(action: "iam:saml-configuration:get")
        ...EditSAMLConfigurationForm_samlConfiguration
      }
    }
  }
`;

interface SAMLConfigurationPageProps {
  queryRef: PreloadedQuery<SAMLConfigurationPageQuery>;
}

export function SAMLConfigurationPage({ queryRef }: SAMLConfigurationPageProps) {
  const { t } = useTranslation();
  const { samlConfiguration } = usePreloadedQuery<SAMLConfigurationPageQuery>(
    samlConfigurationPageQuery,
    queryRef,
  );
  if (
    samlConfiguration.__typename !== "SAMLConfiguration"
    || !samlConfiguration.canGet
  ) {
    throw new NotFoundError(t("samlConfigurationPage.notFound"));
  }

  const title = samlConfiguration.emailDomain !== ""
    ? samlConfiguration.emailDomain
    : t("samlConfigurationPage.title");
  usePageTitle(title);

  const { root, back, intro, form, fields } = newSamlSsoPage();
  const pending = samlConfiguration.domainVerifiedAt == null;
  const domainVerificationToken = samlConfiguration.domainVerificationToken;
  const ssoLoginUrl = showsSamlLoginUrl(
    samlConfiguration.domainVerifiedAt,
    samlConfiguration.enforcementPolicy,
  )
    ? samlConfiguration.testLoginUrl
    : null;

  return (
    <div className={root()}>
      <Link
        to=".."
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("samlConfigurationPage.back")}
      </Link>
      <div className={intro()}>
        <Heading level={1} size={6} weight="medium" highContrast>
          {title}
        </Heading>
      </div>
      <div className={form()}>
        {pending && (
          <Callout color="amber">
            <div className={fields()}>
              <Text>
                <Trans
                  i18nKey="samlConfigurationList.pending.description"
                  components={{
                    type: <Text weight="medium" highContrast color="current" />,
                  }}
                />
              </Text>
              {domainVerificationToken != null && (
                <SAMLConfigurationDnsRecord
                  domainVerificationToken={domainVerificationToken}
                />
              )}
            </div>
          </Callout>
        )}
        <EditSAMLConfigurationForm
          variant="page"
          samlConfigurationKey={samlConfiguration}
          ssoLoginUrl={ssoLoginUrl}
        />
      </div>
    </div>
  );
}
