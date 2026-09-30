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

import { CaretDownIcon, CopyIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Collapsible } from "@probo/ui/src/v2/Collapsible/Collapsible";
import { CollapsiblePanel } from "@probo/ui/src/v2/Collapsible/CollapsiblePanel";
import { CollapsibleTrigger } from "@probo/ui/src/v2/Collapsible/CollapsibleTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { SCIMEventList_scimConfiguration$key } from "#/__generated__/iam/SCIMEventList_scimConfiguration.graphql";
import type { SCIMEventListRefetchQuery } from "#/__generated__/iam/SCIMEventListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import {
  formatJSON,
  resultColor,
  SCIM_EVENT_PAGE_SIZE,
} from "../_lib/scimEvent";
import { scimEventList } from "../variants";

const scimEventListFragment = graphql`
  fragment SCIMEventList_scimConfiguration on SCIMConfiguration
  @refetchable(queryName: "SCIMEventListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    events(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          method
          path
          statusCode
          userName
          ipAddress
          errorMessage
          requestBody
          responseBody
          createdAt
        }
      }
    }
  }
`;

function EventJsonBlock({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const {
    block,
    blockHeading,
    responseWrap,
    response: responseClass,
    responseCopy,
  } = scimEventList();

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value);
      toast({
        title: t("scimEventList.copiedToClipboard"),
        description: label,
        variant: "success",
      });
    } catch {
      toast({
        title: t("scimEventList.copyFailed"),
        description: label,
        variant: "error",
      });
    }
  }

  return (
    <div className={block()}>
      <Text size={1} color="faint" className={blockHeading()}>
        {label}
      </Text>
      <div className={responseWrap()}>
        <IconButton
          variant="surface"
          color="neutral"
          size={1}
          className={responseCopy()}
          aria-label={t("scimEventList.copy")}
          onClick={() => {
            void handleCopy();
          }}
        >
          <CopyIcon />
        </IconButton>
        <pre className={responseClass()}>
          {value}
        </pre>
      </div>
    </div>
  );
}

function EventMetaField({
  color = "neutral",
  label,
  value,
}: {
  color?: "neutral" | "red";
  label: string;
  value: string | number;
}) {
  const { metaField } = scimEventList();

  return (
    <div className={metaField()}>
      <Text size={1} color="faint">
        {label}
      </Text>
      <Text size={2} color={color} highContrast={color === "neutral"}>
        {value}
      </Text>
    </div>
  );
}

function EventRow({
  createdAt,
  errorMessage,
  ipAddress,
  method,
  path,
  requestBody,
  responseBody,
  statusCode,
  userName,
}: {
  createdAt: string;
  errorMessage: string | null | undefined;
  ipAddress: string;
  method: string;
  path: string;
  requestBody: string | null | undefined;
  responseBody: string | null | undefined;
  statusCode: number;
  userName: string;
}) {
  const { t, i18n } = useTranslation();
  const {
    row,
    trigger,
    lead,
    trail,
    path: pathClass,
    caret,
    panel,
    meta,
  } = scimEventList();
  const request = requestBody != null && requestBody !== ""
    ? formatJSON(requestBody)
    : null;
  const response = responseBody != null && responseBody !== ""
    ? formatJSON(responseBody)
    : null;

  return (
    <Collapsible className={row()}>
      <CollapsibleTrigger className={trigger()}>
        <div className={lead()}>
          <Text size={2} color="faint">
            {dateTimeFormat(i18n.language, createdAt)}
          </Text>
          <Code variant="ghost">{method}</Code>
          <Text size={2} className={pathClass()}>
            {path}
          </Text>
        </div>
        <div className={trail()}>
          <Badge size={2} variant="soft" color={resultColor(statusCode)}>
            {statusCode}
          </Badge>
          <CaretDownIcon className={caret()} aria-hidden />
        </div>
      </CollapsibleTrigger>
      <CollapsiblePanel>
        <div className={panel()}>
          <div className={meta()}>
            <EventMetaField
              label={t("scimEventList.details.user")}
              value={userName}
            />
            <EventMetaField
              label={t("scimEventList.details.ipAddress")}
              value={ipAddress}
            />
            <EventMetaField
              label={t("scimEventList.details.statusCode")}
              value={statusCode}
            />
            {errorMessage != null && (
              <EventMetaField
                color="red"
                label={t("scimEventList.details.error")}
                value={errorMessage}
              />
            )}
          </div>
          {request != null && (
            <EventJsonBlock
              label={t("scimEventList.request")}
              value={request}
            />
          )}
          {response != null && (
            <EventJsonBlock
              label={t("scimEventList.response")}
              value={response}
            />
          )}
        </div>
      </CollapsiblePanel>
    </Collapsible>
  );
}

export interface SCIMEventListProps {
  scimConfigurationKey: SCIMEventList_scimConfiguration$key;
}

export function SCIMEventList({ scimConfigurationKey }: SCIMEventListProps) {
  const { t } = useTranslation();
  const [configuration, refetch] = useRefetchableFragment<
    SCIMEventListRefetchQuery,
    SCIMEventList_scimConfiguration$key
  >(scimEventListFragment, scimConfigurationKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const events = configuration.events;
  const pageInfo = events?.pageInfo ?? {
    hasPreviousPage: false,
    hasNextPage: false,
    startCursor: null,
    endCursor: null,
  };
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    pageInfo,
    SCIM_EVENT_PAGE_SIZE,
  );
  const edges = events?.edges ?? [];
  const {
    root,
    results,
    empty,
    item,
    pager,
  } = scimEventList({ pending: isPending });

  if (edges.length === 0) {
    return (
      <Card variant="soft" size={2}>
        <div className={empty()}>
          <Text size={2} color="faint">
            {t("scimEventList.empty")}
          </Text>
        </div>
      </Card>
    );
  }

  return (
    <div className={root()}>
      <div aria-busy={isPending} className={results()}>
        <List>
          {edges.map(({ node }) => (
            <ListItem key={node.id} className={item()}>
              <EventRow
                createdAt={node.createdAt}
                errorMessage={node.errorMessage}
                ipAddress={node.ipAddress}
                method={node.method}
                path={node.path}
                requestBody={node.requestBody}
                responseBody={node.responseBody}
                statusCode={node.statusCode}
                userName={node.userName}
              />
            </ListItem>
          ))}
        </List>
      </div>
      <div className={pager()}>
        <Pagination
          hasPrevious={pageInfo.hasPreviousPage}
          hasNext={pageInfo.hasNextPage}
          previousLabel={t("scimEventList.actions.previous")}
          nextLabel={t("scimEventList.actions.next")}
          showLabels
          disabled={isPending}
          onPrevious={goPrevious}
          onNext={goNext}
        />
      </div>
    </div>
  );
}
