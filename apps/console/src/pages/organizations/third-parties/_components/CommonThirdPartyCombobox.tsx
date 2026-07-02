// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Avatar, ComboboxItem } from "@probo/ui";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";
import { readInlineData } from "relay-runtime";

import type {
  CommonThirdPartyCombobox_commonThirdParty$data,
  CommonThirdPartyCombobox_commonThirdParty$key,
} from "#/__generated__/core/CommonThirdPartyCombobox_commonThirdParty.graphql";
import type { CommonThirdPartyComboboxQuery } from "#/__generated__/core/CommonThirdPartyComboboxQuery.graphql";
import type { CreateThirdPartyInput } from "#/__generated__/core/CreateThirdPartyDialogCreateMutation.graphql";

export type CommonThirdPartyRef
  = CommonThirdPartyCombobox_commonThirdParty$data;

export const commonThirdPartyFragment = graphql`
  fragment CommonThirdPartyCombobox_commonThirdParty on CommonThirdParty @inline {
    name
    logo {
      downloadUrl
    }
    category
    websiteUrl
    headquarterAddress
    legalName
    privacyPolicyUrl
    serviceLevelAgreementUrl
    dataProcessingAgreementUrl
    certifications
    securityPageUrl
    trustPageUrl
    statusPageUrl
    termsOfServiceUrl
  }
`;

export const commonThirdPartiesQuery = graphql`
  query CommonThirdPartyComboboxQuery($name: String!) {
    commonThirdParties(name: $name) {
      id
      name
      logo {
      downloadUrl
    }
      ...CommonThirdPartyCombobox_commonThirdParty
    }
  }
`;

function toCreateInput(tp: CommonThirdPartyRef): Omit<CreateThirdPartyInput, "organizationId"> {
  return {
    name: tp.name,
    headquarterAddress: tp.headquarterAddress,
    legalName: tp.legalName,
    websiteUrl: tp.websiteUrl,
    category: tp.category,
    privacyPolicyUrl: tp.privacyPolicyUrl,
    serviceLevelAgreementUrl: tp.serviceLevelAgreementUrl,
    dataProcessingAgreementUrl: tp.dataProcessingAgreementUrl,
    certifications: tp.certifications,
    securityPageUrl: tp.securityPageUrl,
    trustPageUrl: tp.trustPageUrl,
    statusPageUrl: tp.statusPageUrl,
    termsOfServiceUrl: tp.termsOfServiceUrl,
  };
}

interface CommonThirdPartyComboboxProps {
  queryRef: PreloadedQuery<CommonThirdPartyComboboxQuery>;
  onSelect: (thirdParty: Omit<CreateThirdPartyInput, "organizationId">) => void;
  excludeNames?: Set<string>;
}

export function CommonThirdPartyCombobox({
  queryRef,
  onSelect,
  excludeNames,
}: CommonThirdPartyComboboxProps) {
  const data = usePreloadedQuery<CommonThirdPartyComboboxQuery>(commonThirdPartiesQuery, queryRef);

  const items = excludeNames
    ? data.commonThirdParties.filter(tp => !excludeNames.has(tp.name.toLowerCase()))
    : data.commonThirdParties;

  return (
    <>
      {items.map(thirdParty => (
        <ComboboxItem
          key={thirdParty.id}
          onClick={() => {
            const tp = readInlineData<CommonThirdPartyCombobox_commonThirdParty$key>(
              commonThirdPartyFragment,
              thirdParty,
            );
            onSelect(toCreateInput(tp));
          }}
        >
          <Avatar
            name={thirdParty.name}
            src={thirdParty.logo?.downloadUrl}
          />
          {thirdParty.name}
        </ComboboxItem>
      ))}
    </>
  );
}
