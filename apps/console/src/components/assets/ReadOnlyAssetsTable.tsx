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

import { faviconUrl, getAssetTypeVariant } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { usePaginationFragmentHookType } from "react-relay/relay-hooks/usePaginationFragment";
import type { OperationType } from "relay-runtime";

import type {
  AssetsPageFragment$data,
  AssetsPageFragment$key,
} from "#/__generated__/core/AssetsPageFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { SortableTable } from "../SortableTable";

type AssetEntry = NodeOf<AssetsPageFragment$data["assets"]>;

type Props = {
  pagination: usePaginationFragmentHookType<
    OperationType,
    AssetsPageFragment$key,
    AssetsPageFragment$data
  >;
  assets: AssetEntry[];
};

export function ReadOnlyAssetsTable(props: Props) {
  const { pagination, assets } = props;
  const { __ } = useTranslate();

  return (
    <SortableTable {...pagination} pageSize={10}>
      <Thead>
        <Tr>
          <Th>{__("Name")}</Th>
          <Th>{__("Type")}</Th>
          <Th>{__("Amount")}</Th>
          <Th>{__("Owner")}</Th>
          <Th>{__("Third parties")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {assets.map(entry => (
          <AssetRow key={entry.id} entry={entry} />
        ))}
      </Tbody>
    </SortableTable>
  );
}

function AssetRow({ entry }: { entry: AssetEntry }) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const thirdParties = entry.thirdParties?.edges.map(edge => edge.node) ?? [];

  return (
    <Tr to={`/organizations/${organizationId}/assets/${entry.id}`}>
      <Td>{entry.name}</Td>
      <Td>
        <Badge variant={getAssetTypeVariant(entry.assetType)}>
          {entry.assetType === "PHYSICAL" ? __("Physical") : __("Virtual")}
        </Badge>
      </Td>
      <Td>{entry.amount}</Td>
      <Td>{entry.owner?.fullName ?? __("Unassigned")}</Td>
      <Td>
        {thirdParties.length > 0
          ? (
              <div className="flex flex-wrap gap-1">
                {thirdParties.slice(0, 3).map(thirdParty => (
                  <Badge
                    key={thirdParty.id}
                    variant="neutral"
                    className="flex items-center gap-1"
                  >
                    <Avatar
                      name={thirdParty.name}
                      src={faviconUrl(thirdParty.websiteUrl)}
                      size="s"
                    />
                    <span className="text-xs">{thirdParty.name}</span>
                  </Badge>
                ))}
                {thirdParties.length > 3 && (
                  <Badge variant="neutral" className="text-xs">
                    +
                    {thirdParties.length - 3}
                  </Badge>
                )}
              </div>
            )
          : (
              <span className="text-txt-secondary text-sm">{__("None")}</span>
            )}
      </Td>
    </Tr>
  );
}
