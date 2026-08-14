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

import { faviconUrl } from "@probo/helpers";
import {
  ActionDropdown,
  Avatar,
  Badge,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, useFragment } from "react-relay";

import type { BusinessFunctionListItem_businessFunction$key } from "#/__generated__/core/BusinessFunctionListItem_businessFunction.graphql";
import type { BusinessFunctionListItemDeleteMutation } from "#/__generated__/core/BusinessFunctionListItemDeleteMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  businessFunctionListConnectionFilters,
  BusinessFunctionsConnectionKey,
  getClassificationLabel,
  getClassificationVariant,
} from "../_lib/businessFunctionHelpers";

const businessFunctionListItemFragment = graphql`
  fragment BusinessFunctionListItem_businessFunction on BusinessFunction {
    id
    referenceId
    name
    classification
    mtdMinutes
    rtoMinutes
    rpoMinutes
    owner {
      id
      fullName
    }
    assets(first: 50) {
      edges {
        node {
          id
          name
        }
      }
    }
    thirdParties(first: 50) {
      edges {
        node {
          id
          name
          websiteUrl
        }
      }
    }
    canDelete: permission(action: "core:business-function:delete")
  }
`;

const deleteBusinessFunctionMutation = graphql`
  mutation BusinessFunctionListItemDeleteMutation(
    $input: DeleteBusinessFunctionInput!
    $connections: [ID!]!
  ) {
    deleteBusinessFunction(input: $input) {
      deletedBusinessFunctionId @deleteEdge(connections: $connections)
    }
  }
`;

interface BusinessFunctionListItemProps {
  businessFunctionKey: BusinessFunctionListItem_businessFunction$key;
  hasAnyAction: boolean;
}

export function BusinessFunctionListItem({
  businessFunctionKey,
  hasAnyAction,
}: BusinessFunctionListItemProps) {
  const businessFunction = useFragment(
    businessFunctionListItemFragment,
    businessFunctionKey,
  );
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const confirm = useConfirm();
  const { toast } = useToast();
  const [deleteBusinessFunction] = useMutation<BusinessFunctionListItemDeleteMutation>(
    deleteBusinessFunctionMutation,
    {
      errorToast: t("businessFunctionListItem.errors.delete"),
    },
  );

  const deleteConnections = businessFunctionListConnectionFilters(businessFunction).map(filter =>
    ConnectionHandler.getConnectionID(
      organizationId,
      BusinessFunctionsConnectionKey,
      { filter },
    ),
  );

  const handleDelete = () => {
    confirm(
      async () => {
        await deleteBusinessFunction({
          variables: {
            input: {
              businessFunctionId: businessFunction.id,
            },
            connections: deleteConnections,
          },
        });
        toast({
          title: t("businessFunctionListItem.messages.success"),
          description: t("businessFunctionListItem.messages.deleted"),
          variant: "success",
        });
      },
      {
        message: t("businessFunctionListItem.deleteConfirmation", {
          referenceId: businessFunction.referenceId,
        }),
      },
    );
  };

  const detailsUrl
    = `/organizations/${organizationId}/registries/business-functions/${businessFunction.id}`;
  const assets = businessFunction.assets?.edges.map(edge => edge.node) ?? [];
  const thirdParties = businessFunction.thirdParties?.edges.map(edge => edge.node) ?? [];

  return (
    <Tr to={detailsUrl}>
      <Td>
        <span className="font-mono text-sm">{businessFunction.referenceId}</span>
      </Td>
      <Td>{businessFunction.name}</Td>
      <Td>
        <Badge variant={getClassificationVariant(businessFunction.classification)}>
          {getClassificationLabel(businessFunction.classification, t, "businessFunctionsPage")}
        </Badge>
      </Td>
      <Td>{businessFunction.mtdMinutes}</Td>
      <Td>{businessFunction.rtoMinutes}</Td>
      <Td>{businessFunction.rpoMinutes}</Td>
      <Td>{businessFunction.owner?.fullName || "-"}</Td>
      <Td>
        <EntityBadgeList
          items={assets}
          emptyLabel={t("businessFunctionsPage.none")}
        />
      </Td>
      <Td>
        <EntityBadgeList
          items={thirdParties}
          emptyLabel={t("businessFunctionsPage.none")}
          showFavicon
        />
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          {businessFunction.canDelete && (
            <ActionDropdown>
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {t("businessFunctionListItem.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </Td>
      )}
    </Tr>
  );
}

type BadgeItem = {
  id: string;
  name: string;
  websiteUrl?: string | null;
};

function EntityBadgeList({
  items,
  emptyLabel,
  showFavicon = false,
}: {
  items: BadgeItem[];
  emptyLabel: string;
  showFavicon?: boolean;
}) {
  if (items.length === 0) {
    return <span className="text-txt-secondary text-sm">{emptyLabel}</span>;
  }

  return (
    <div className="flex flex-wrap gap-1">
      {items.slice(0, 3).map(item => (
        <Badge
          key={item.id}
          variant="neutral"
          className="flex items-center gap-1"
        >
          {showFavicon && (
            <Avatar
              name={item.name}
              src={faviconUrl(item.websiteUrl)}
              size="s"
            />
          )}
          <span className="text-xs">{item.name}</span>
        </Badge>
      ))}
      {items.length > 3 && (
        <Badge variant="neutral" className="text-xs">
          +
          {items.length - 3}
        </Badge>
      )}
    </div>
  );
}
