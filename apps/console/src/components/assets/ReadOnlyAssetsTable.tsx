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
          <Th>{__("Vendors")}</Th>
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
  const vendors = entry.vendors?.edges.map(edge => edge.node) ?? [];

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
        {vendors.length > 0
          ? (
              <div className="flex flex-wrap gap-1">
                {vendors.slice(0, 3).map(vendor => (
                  <Badge
                    key={vendor.id}
                    variant="neutral"
                    className="flex items-center gap-1"
                  >
                    <Avatar
                      name={vendor.name}
                      src={faviconUrl(vendor.websiteUrl)}
                      size="s"
                    />
                    <span className="text-xs">{vendor.name}</span>
                  </Badge>
                ))}
                {vendors.length > 3 && (
                  <Badge variant="neutral" className="text-xs">
                    +
                    {vendors.length - 3}
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
