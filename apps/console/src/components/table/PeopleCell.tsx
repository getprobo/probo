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

import { Avatar } from "@probo/ui";

import type { PeopleGraphQuery } from "#/__generated__/core/PeopleGraphQuery.graphql";
import { GraphQLCell } from "#/components/table/GraphQLCell";
import { peopleQuery } from "#/hooks/graph/PeopleGraph";

type Props = {
  name: string;
  defaultValue?: { fullName: string; id: string };
  organizationId: string;
};

export function PeopleCell(props: Props) {
  return (
    <GraphQLCell<PeopleGraphQuery, { fullName: string }>
      name={props.name}
      query={peopleQuery}
      variables={{
        organizationId: props.organizationId,
        filter: { excludeContractEnded: true },
      }}
      items={data =>
        data.organization?.profiles?.edges.map(edge => edge.node) ?? []}
      itemRenderer={({ item }) => (
        <div className="flex gap-2 whitespace-nowrap items-center text-xs">
          <Avatar name={item.fullName} />
          {item.fullName}
        </div>
      )}
      defaultValue={props.defaultValue}
    />
  );
}
