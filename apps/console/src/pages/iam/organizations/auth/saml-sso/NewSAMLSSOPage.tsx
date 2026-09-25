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
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { NewSAMLSSOPageQuery } from "#/__generated__/iam/NewSAMLSSOPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { NewSAMLConfigurationForm } from "./_components/NewSAMLConfigurationForm";
import { newSamlSsoPage } from "./variants";

export const newSamlSsoPageQuery = graphql`
  query NewSAMLSSOPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateSAMLConfiguration: permission(
          action: "iam:saml-configuration:create"
        )
      }
    }
  }
`;

interface NewSAMLSSOPageProps {
  queryRef: PreloadedQuery<NewSAMLSSOPageQuery>;
}

export function NewSAMLSSOPage({ queryRef }: NewSAMLSSOPageProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const title = t("newSamlSsoPage.title");
  usePageTitle(title);

  const { organization } = usePreloadedQuery<NewSAMLSSOPageQuery>(
    newSamlSsoPageQuery,
    queryRef,
  );
  if (
    organization.__typename !== "Organization"
    || !organization.canCreateSAMLConfiguration
  ) {
    throw new NotFoundError(t("newSamlSsoPage.notFound"));
  }

  const { root, back, intro } = newSamlSsoPage();

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
        {t("newSamlSsoPage.back")}
      </Link>
      <div className={intro()}>
        <Heading level={1} size={6} weight="medium" highContrast>
          {title}
        </Heading>
        <Text size={2} color="faint">
          {t("newSamlSsoPage.description")}
        </Text>
      </div>
      <NewSAMLConfigurationForm
        variant="page"
        onCreate={() => {
          void navigate("..");
        }}
      />
    </div>
  );
}
