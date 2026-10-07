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

import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { IntegrationsPageQuery } from "#/__generated__/core/IntegrationsPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { IntegrationsConnectors } from "./_components/IntegrationsConnectors";

export const integrationsPageQuery = graphql`
  query IntegrationsPageQuery($organizationId: ID!, $filter: ConnectorFilter) {
    ...IntegrationsConnectors_query
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...IntegrationsConnectors_organization @arguments(filter: $filter)
      }
    }
  }
`;

interface IntegrationsPageProps {
  queryRef: PreloadedQuery<IntegrationsPageQuery>;
}

export function IntegrationsPage({ queryRef }: IntegrationsPageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const data = usePreloadedQuery<IntegrationsPageQuery>(integrationsPageQuery, queryRef);
  const { organization } = data;

  usePageTitle(t("listPage.title"));

  const callbackConnectorId = searchParams.get("connector_id");
  const callbackError = searchParams.get("error");

  useEffect(() => {
    if (callbackConnectorId) {
      if (callbackError) {
        toast({
          title: t("listPage.messages.error"),
          description: callbackError,
          variant: "error",
        });
      }

      setSearchParams((params) => {
        params.delete("connector_id");
        params.delete("provider");
        params.delete("error");
        return params;
      }, { replace: true });
      return;
    }

    if (callbackError) {
      toast({
        title: t("listPage.messages.error"),
        description: callbackError,
        variant: "error",
      });
      setSearchParams((params) => {
        params.delete("error");
        return params;
      }, { replace: true });
    }
  }, [
    callbackConnectorId,
    callbackError,
    setSearchParams,
    t,
    toast,
  ]);

  if (organization.__typename !== "Organization") {
    throw new NotFoundError(t("listPage.notFound"));
  }

  return (
    <IntegrationsConnectors
      queryKey={data}
      organizationKey={organization}
    />
  );
}
