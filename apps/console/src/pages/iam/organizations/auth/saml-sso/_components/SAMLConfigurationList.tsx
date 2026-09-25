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

import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SAMLConfigurationList_organization$key } from "#/__generated__/iam/SAMLConfigurationList_organization.graphql";
import type { SAMLConfigurationListRefetchQuery } from "#/__generated__/iam/SAMLConfigurationListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { samlConfigurationList, samlSsoPage } from "../variants";

import { SAMLConfigurationListItem } from "./SAMLConfigurationListItem";

export const SAML_CONFIGURATION_PAGE_SIZE = 15;

export const samlConfigurationListFragment = graphql`
  fragment SAMLConfigurationList_organization on Organization
  @refetchable(queryName: "SAMLConfigurationListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    canCreateSAMLConfiguration: permission(
      action: "iam:saml-configuration:create"
    )
    samlConfigurations(
      first: $first
      after: $after
      last: $last
      before: $before
    )
      @required(action: THROW)
      @connection(key: "SAMLConfigurationListFragment_samlConfigurations") {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges @required(action: THROW) {
        node {
          id
          ...SAMLConfigurationListItem_samlConfiguration
        }
      }
    }
  }
`;

interface SAMLConfigurationListProps {
  organizationKey: SAMLConfigurationList_organization$key;
  onAdd: () => void;
}

export function SAMLConfigurationList({
  organizationKey,
  onAdd,
}: SAMLConfigurationListProps) {
  const { t } = useTranslation();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const [organization, refetch] = useRefetchableFragment<
    SAMLConfigurationListRefetchQuery,
    SAMLConfigurationList_organization$key
  >(samlConfigurationListFragment, organizationKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: SAML_CONFIGURATION_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  });

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    pageVariablesRef.current = variables;
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    organization.samlConfigurations.pageInfo,
    SAML_CONFIGURATION_PAGE_SIZE,
  );

  const edges = organization.samlConfigurations.edges;
  const pageInfo = organization.samlConfigurations.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { header, intro } = samlSsoPage();
  const { root, results, grid, empty, pager } = samlConfigurationList({ pending: isPending });

  function handleDeleted() {
    if (edges.length === 1 && pageInfo.hasPreviousPage) {
      goPrevious();
      return;
    }
    startRefetchTransition(() => {
      refetch(pageVariablesRef.current, { fetchPolicy: "network-only" });
    });
  }

  return (
    <>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("samlSsoPage.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("samlSsoPage.description")}
          </Text>
        </div>
        {organization.canCreateSAMLConfiguration && (
          <Button variant="solid" iconStart={<PlusIcon />} onClick={onAdd}>
            {t("samlSsoPage.actions.addConfiguration")}
          </Button>
        )}
      </div>
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t("samlConfigurationList.empty.description")}
                </Text>
              </div>
            </Card>
          )
        : (
            <div className={root()}>
              <div aria-busy={isPending} className={results()}>
                <div className={grid()}>
                  {edges.map(({ node }) => (
                    <SAMLConfigurationListItem
                      key={node.id}
                      samlConfigurationKey={node}
                      onDeleted={handleDeleted}
                    />
                  ))}
                </div>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("samlSsoPage.actions.previous")}
                  nextLabel={t("samlSsoPage.actions.next")}
                  showLabels
                  variant="soft"
                  disabled={isPending}
                  onPrevious={goPrevious}
                  onNext={goNext}
                />
              </div>
            </div>
          )}
    </>
  );
}
