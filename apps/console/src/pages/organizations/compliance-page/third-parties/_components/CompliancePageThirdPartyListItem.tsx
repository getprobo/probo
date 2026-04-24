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
