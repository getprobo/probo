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

import { Badge, Checkbox, Td, Tr } from "@probo/ui";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  CompliancePortalThirdPartyListItem_compliancePortal$key,
} from "#/__generated__/core/CompliancePortalThirdPartyListItem_compliancePortal.graphql";
import type {
  CompliancePortalThirdPartyListItem_linkMutation,
} from "#/__generated__/core/CompliancePortalThirdPartyListItem_linkMutation.graphql";
import type {
  CompliancePortalThirdPartyListItem_removeMutation,
} from "#/__generated__/core/CompliancePortalThirdPartyListItem_removeMutation.graphql";
import type {
  CompliancePortalThirdPartyListItem_thirdParty$key,
} from "#/__generated__/core/CompliancePortalThirdPartyListItem_thirdParty.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const compliancePortalFragment = graphql`
  fragment CompliancePortalThirdPartyListItem_compliancePortal on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const thirdPartyFragment = graphql`
  fragment CompliancePortalThirdPartyListItem_thirdParty on ThirdParty
  @argumentDefinitions(compliancePortalId: { type: "ID!" }) {
    id
    name
    category
    compliancePortalThirdParty(compliancePortalId: $compliancePortalId) {
      id
    }
  }
`;

const linkThirdPartyMutation = graphql`
  mutation CompliancePortalThirdPartyListItem_linkMutation(
    $input: UpdateCompliancePortalThirdPartyPublishedInput!
  ) {
    updateCompliancePortalThirdPartyPublished(input: $input) {
      catalogThirdParty {
        id
        thirdParty {
          id
        }
      }
    }
  }
`;

const removeThirdPartyMutation = graphql`
  mutation CompliancePortalThirdPartyListItem_removeMutation(
    $input: DeleteCompliancePortalThirdPartyInput!
  ) {
    deleteCompliancePortalThirdParty(input: $input) {
      deletedCompliancePortalThirdPartyId @deleteRecord
    }
  }
`;

export function CompliancePortalThirdPartyListItem(props: {
  compliancePortalKey: CompliancePortalThirdPartyListItem_compliancePortal$key;
  thirdPartyKey: CompliancePortalThirdPartyListItem_thirdParty$key;
}) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");

  const compliancePortal = useFragment<CompliancePortalThirdPartyListItem_compliancePortal$key>(
    compliancePortalFragment,
    props.compliancePortalKey,
  );
  const thirdParty = useFragment<CompliancePortalThirdPartyListItem_thirdParty$key>(
    thirdPartyFragment,
    props.thirdPartyKey,
  );
  const catalogThirdParty = thirdParty.compliancePortalThirdParty;
  const serverLinked = catalogThirdParty !== null;
  const [pendingLinked, setPendingLinked] = useState<boolean | null>(null);
  const isLinked = pendingLinked ?? serverLinked;

  const [linkThirdParty, isLinking]
    = useMutation<CompliancePortalThirdPartyListItem_linkMutation>(
      linkThirdPartyMutation,
      {
        successMessage: t("thirdPartyListItem.messages.linked"),
        errorToast: t("thirdPartyListItem.errors.link"),
      },
    );

  const [removeThirdParty, isRemoving]
    = useMutation<CompliancePortalThirdPartyListItem_removeMutation>(
      removeThirdPartyMutation,
      {
        successMessage: t("thirdPartyListItem.messages.removed"),
        errorToast: t("thirdPartyListItem.errors.remove"),
      },
    );

  const handleLinkedChange = useCallback(
    async (checked: boolean) => {
      if (!compliancePortal.canUpdate || checked === isLinked) {
        return;
      }

      setPendingLinked(checked);

      try {
        if (checked) {
          await linkThirdParty({
            variables: {
              input: {
                compliancePortalId: compliancePortal.id,
                thirdPartyId: thirdParty.id,
                published: true,
              },
            },
            updater: (store) => {
              const payload = store.getRootField(
                "updateCompliancePortalThirdPartyPublished",
              );
              const link = payload?.getLinkedRecord("catalogThirdParty");
              const thirdPartyRecord = store.get(thirdParty.id);
              if (link && thirdPartyRecord) {
                thirdPartyRecord.setLinkedRecord(
                  link,
                  "compliancePortalThirdParty",
                  { compliancePortalId: compliancePortal.id },
                );
              }
            },
          });
          setPendingLinked(null);
          return;
        }

        if (!catalogThirdParty) {
          setPendingLinked(null);
          return;
        }

        await removeThirdParty({
          variables: {
            input: {
              id: catalogThirdParty.id,
            },
          },
          updater: (store) => {
            store.get(thirdParty.id)?.setValue(
              null,
              "compliancePortalThirdParty",
              { compliancePortalId: compliancePortal.id },
            );
          },
        });
        setPendingLinked(null);
      } catch {
        setPendingLinked(null);
      }
    },
    [
      catalogThirdParty,
      compliancePortal.canUpdate,
      compliancePortal.id,
      isLinked,
      linkThirdParty,
      removeThirdParty,
      thirdParty.id,
    ],
  );

  const isMutating = isLinking || isRemoving;

  return (
    <Tr to={`/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`}>
      <Td noLink>
        <Checkbox
          checked={isLinked}
          onChange={checked => void handleLinkedChange(checked)}
          disabled={isMutating || !compliancePortal.canUpdate}
          aria-label={t("thirdPartyListItem.actions.toggle", {
            title: thirdParty.name,
          })}
        />
      </Td>
      <Td>
        <div className="flex gap-4 items-center">{thirdParty.name}</div>
      </Td>
      <Td>
        <Badge variant="neutral">{thirdParty.category}</Badge>
      </Td>
    </Tr>
  );
}
