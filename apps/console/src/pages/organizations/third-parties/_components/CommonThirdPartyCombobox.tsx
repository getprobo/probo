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
