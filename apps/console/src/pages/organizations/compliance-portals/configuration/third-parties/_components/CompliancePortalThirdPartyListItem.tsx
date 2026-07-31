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

import { Badge, Button, IconCrossLargeX, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { CompliancePortalThirdPartyListItem_catalogThirdPartyFragment$key } from "#/__generated__/core/CompliancePortalThirdPartyListItem_catalogThirdPartyFragment.graphql";
import type { CompliancePortalThirdPartyListItem_removeMutation } from "#/__generated__/core/CompliancePortalThirdPartyListItem_removeMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const catalogThirdPartyFragment = graphql`
  fragment CompliancePortalThirdPartyListItem_catalogThirdPartyFragment on CompliancePortalThirdParty {
    id
    thirdParty {
      id
      category
      name
      canUpdate: permission(action: "core:thirdParty:update")
    }
  }
`;

const removeThirdPartyMutation = graphql`
  mutation CompliancePortalThirdPartyListItem_removeMutation(
    $input: DeleteCompliancePortalThirdPartyInput!
    $connections: [ID!]!
  ) {
    deleteCompliancePortalThirdParty(input: $input) {
      deletedCompliancePortalThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

export function CompliancePortalThirdPartyListItem(props: {
  catalogThirdPartyFragmentRef: CompliancePortalThirdPartyListItem_catalogThirdPartyFragment$key;
  thirdPartiesConnectionId: DataID;
  canUpdatePortal: boolean;
}) {
  const {
    catalogThirdPartyFragmentRef,
    thirdPartiesConnectionId,
    canUpdatePortal,
  } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");

  const catalogThirdParty = useFragment<CompliancePortalThirdPartyListItem_catalogThirdPartyFragment$key>(
    catalogThirdPartyFragment,
    catalogThirdPartyFragmentRef,
  );
  const thirdParty = catalogThirdParty.thirdParty;

  const [removeThirdParty, isRemoving] = useMutation<CompliancePortalThirdPartyListItem_removeMutation>(
    removeThirdPartyMutation,
    {
      successMessage: t("thirdPartyListItem.messages.removed"),
      errorToast: t("thirdPartyListItem.errors.remove"),
    },
  );

  const handleRemove = useCallback(async () => {
    await removeThirdParty({
      variables: {
        connections: [thirdPartiesConnectionId],
        input: {
          id: catalogThirdParty.id,
        },
      },
    });
  }, [catalogThirdParty.id, removeThirdParty, thirdPartiesConnectionId]);

  return (
    <Tr to={`/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`}>
      <Td>
        <div className="flex gap-4 items-center">{thirdParty.name}</div>
      </Td>
      <Td>
        <Badge variant="neutral">{thirdParty.category}</Badge>
      </Td>
      <Td noLink width={48}>
        {canUpdatePortal && (
          <Button
            variant="tertiary"
            icon={IconCrossLargeX}
            aria-label={t("thirdPartyListItem.actions.remove")}
            disabled={isRemoving}
            onClick={() => void handleRemove()}
          />
        )}
      </Td>
    </Tr>
  );
}
