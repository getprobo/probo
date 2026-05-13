// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { faviconUrl } from "@probo/helpers";
import { Avatar, Badge, IconCrossLargeX } from "@probo/ui";

import type { ThirdPartyGraphSelectQuery } from "#/__generated__/core/ThirdPartyGraphSelectQuery.graphql";
import { GraphQLCell } from "#/components/table/GraphQLCell";
import { thirdPartiesSelectQuery } from "#/hooks/graph/ThirdPartyGraph";

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
    <GraphQLCell<ThirdPartyGraphSelectQuery, ThirdParty>
      multiple
      name={props.name}
      query={thirdPartiesSelectQuery}
      variables={{
        organizationId: props.organizationId,
      }}
      items={data =>
        data.organization?.thirdParties?.edges?.map(edge => edge.node) ?? []}
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
