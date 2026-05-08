// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
import { Avatar, ComboboxItem } from "@probo/ui";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import type {
  CommonThirdPartyComboboxQuery$data,
  CommonThirdPartyComboboxQuery,
} from "#/__generated__/core/CommonThirdPartyComboboxQuery.graphql";

export type CommonThirdPartyRef
  = CommonThirdPartyComboboxQuery$data["commonThirdParties"][number];

export const commonThirdPartiesQuery = graphql`
  query CommonThirdPartyComboboxQuery($name: String!) {
    commonThirdParties(name: $name) {
      id
      name
      websiteUrl
      ...CreateVendorDialog_commonThirdParty
    }
  }
`;

interface CommonThirdPartyComboboxProps {
  queryRef: PreloadedQuery<CommonThirdPartyComboboxQuery>;
  onSelect: (thirdPartyRef: CommonThirdPartyRef) => void;
}

export function CommonThirdPartyCombobox({
  queryRef,
  onSelect,
}: CommonThirdPartyComboboxProps) {
  const data = usePreloadedQuery(commonThirdPartiesQuery, queryRef);

  return (
    <>
      {data.commonThirdParties.map(thirdParty => (
        <ComboboxItem
          key={thirdParty.id}
          onClick={() => onSelect(thirdParty)}
        >
          <Avatar
            name={thirdParty.name}
            src={faviconUrl(thirdParty.websiteUrl)}
          />
          {thirdParty.name}
        </ComboboxItem>
      ))}
    </>
  );
}
