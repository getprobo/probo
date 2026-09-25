// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { usePageTitle } from "@probo/hooks";
import { Breadcrumb, Dialog, useDialogRef } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { SAMLSSOPageQuery } from "#/__generated__/iam/SAMLSSOPageQuery.graphql";

import { NewSAMLConfigurationForm } from "./_components/NewSAMLConfigurationForm";
import { SAMLConfigurationList } from "./_components/SAMLConfigurationList";
import { samlSsoPage } from "./variants";

export const samlSSOPageQuery = graphql`
  query SAMLSSOPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...SAMLConfigurationList_organization
      }
    }
  }
`;

export function SAMLSSOPage(props: {
  queryRef: PreloadedQuery<SAMLSSOPageQuery>;
}) {
  const { queryRef } = props;

  const formDialogRef = useDialogRef();

  const { t } = useTranslation();
  const { root } = samlSsoPage();
  usePageTitle(t("samlSsoPage.title"));

  const { organization } = usePreloadedQuery<SAMLSSOPageQuery>(samlSSOPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  return (
    <>
      <div className={root()}>
        <SAMLConfigurationList
          organizationKey={organization}
          onAdd={() => formDialogRef.current?.open()}
        />
      </div>

      <Dialog
        ref={formDialogRef}
        onClose={() => formDialogRef.current?.close()}
        title={<Breadcrumb items={[t("samlSsoPage.breadcrumb.settings"), t("samlSsoPage.breadcrumb.configure")]} />}
      >
        <NewSAMLConfigurationForm onCreate={() => formDialogRef.current?.close()} />
      </Dialog>
    </>
  );
}
