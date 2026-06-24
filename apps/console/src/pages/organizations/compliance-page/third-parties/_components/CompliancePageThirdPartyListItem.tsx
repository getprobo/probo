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

import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconCheckmark1, IconCrossLargeX, Td, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageThirdPartyListItem_thirdPartyFragment$key } from "#/__generated__/core/CompliancePageThirdPartyListItem_thirdPartyFragment.graphql";
import type { CompliancePageThirdPartyListItemMutation } from "#/__generated__/core/CompliancePageThirdPartyListItemMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const thirdPartyFragment = graphql`
  fragment CompliancePageThirdPartyListItem_thirdPartyFragment on ThirdParty {
    id
    category
    name
    showOnTrustCenter
    canUpdate: permission(action: "core:thirdParty:update")
  }
`;

const updateThirdPartyVisibilityMutation = graphql`
  mutation CompliancePageThirdPartyListItemMutation($input: UpdateThirdPartyInput!) {
    updateThirdParty(input: $input) {
      thirdParty {
        id
        showOnTrustCenter
        ...CompliancePageThirdPartyListItem_thirdPartyFragment
      }
    }
  }
`;

export function CompliancePageThirdPartyListItem(props: {
  thirdPartyFragmentRef: CompliancePageThirdPartyListItem_thirdPartyFragment$key;
}) {
  const { thirdPartyFragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const thirdParty = useFragment<CompliancePageThirdPartyListItem_thirdPartyFragment$key>(
    thirdPartyFragment,
    thirdPartyFragmentRef,
  );
  const [updateThirdPartyVisibility, isUpadtingThirdPartyVisibility] = useMutationWithToasts<
    CompliancePageThirdPartyListItemMutation
  >(
    updateThirdPartyVisibilityMutation,
    {
      successMessage: __("Subprocessor visibility updated successfully."),
      errorMessage: __("Failed to update subprocessor visibility"),
    },
  );

  return (
    <Tr to={`/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`}>
      <Td>
        <div className="flex gap-4 items-center">{thirdParty.name}</div>
      </Td>
      <Td>
        <Badge variant="neutral">{thirdParty.category}</Badge>
      </Td>
      <Td>
        <Badge variant={thirdParty.showOnTrustCenter ? "success" : "danger"}>
          {thirdParty.showOnTrustCenter ? __("Visible") : __("None")}
        </Badge>
      </Td>
      <Td noLink width={100} className="text-end">
        {thirdParty.canUpdate && (
          <Button
            variant="secondary"
            onClick={() =>
              void updateThirdPartyVisibility({
                variables: {
                  input: {
                    id: thirdParty.id,
                    showOnTrustCenter: !thirdParty.showOnTrustCenter,
                  },
                },
              })}
            icon={thirdParty.showOnTrustCenter ? IconCrossLargeX : IconCheckmark1}
            disabled={isUpadtingThirdPartyVisibility}
          >
            {thirdParty.showOnTrustCenter ? __("Hide") : __("Show")}
          </Button>
        )}
      </Td>
    </Tr>
  );
};
