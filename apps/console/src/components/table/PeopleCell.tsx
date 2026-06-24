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
        filter: { contractEnded: false },
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
