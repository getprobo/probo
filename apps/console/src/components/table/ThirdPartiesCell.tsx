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

import { faviconUrl } from "@probo/helpers";
import { Avatar, Badge, IconCrossLargeX } from "@probo/ui";
import { graphql } from "relay-runtime";

import type { ThirdPartiesCellQuery } from "#/__generated__/core/ThirdPartiesCellQuery.graphql";
import { GraphQLCell } from "#/components/table/GraphQLCell";

const thirdPartiesCellQuery = graphql`
  query ThirdPartiesCellQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        thirdParties(
          first: 100
          orderBy: { direction: ASC, field: NAME }
        ) {
          edges {
            node {
              id
              name
              websiteUrl
            }
          }
        }
      }
    }
  }
`;

type ThirdParty = {
  id: string;
  name: string;
  websiteUrl: string | null | undefined;
};

type Props = {
  name: string;
  defaultValue?: ThirdParty[];
  organizationId: string;
};

const empty = [] as ThirdParty[];

export function ThirdPartiesCell(props: Props) {
  return (
    <GraphQLCell<ThirdPartiesCellQuery, ThirdParty>
      multiple
      name={props.name}
      query={thirdPartiesCellQuery}
      variables={{
        organizationId: props.organizationId,
      }}
      items={data =>
        data.organization?.thirdParties?.edges?.map(edge => ({
          id: edge.node.id,
          name: edge.node.name,
          websiteUrl: edge.node.websiteUrl,
        })) ?? []}
      itemRenderer={({ item, onRemove }) => (
        <ThirdPartyBadge thirdParty={item} onRemove={onRemove} />
      )}
      defaultValue={props.defaultValue ?? empty}
    />
  );
}

function ThirdPartyBadge({
  thirdParty,
  onRemove,
}: {
  thirdParty: ThirdParty;
  onRemove?: (v: ThirdParty) => void;
}) {
  return (
    <Badge variant="neutral" className="flex items-center gap-1">
      <Avatar name={thirdParty.name} src={faviconUrl(thirdParty.websiteUrl)} size="s" />
      <span className="max-w-[100px] text-ellipsis overflow-hidden min-w-0 block">
        {thirdParty.name}
      </span>
      {onRemove && (
        <button
          onClick={() => onRemove(thirdParty)}
          className="size-4 hover:text-txt-primary cursor-pointer"
          type="button"
        >
          <IconCrossLargeX size={14} />
        </button>
      )}
    </Badge>
  );
}
