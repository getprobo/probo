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

import { promisifyMutation } from "@probo/helpers";
import { useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  AssetGraphCreateMutation,
  AssetType,
} from "#/__generated__/core/AssetGraphCreateMutation.graphql";
import type { AssetGraphDeleteMutation } from "#/__generated__/core/AssetGraphDeleteMutation.graphql";
import type { AssetGraphUpdateMutation } from "#/__generated__/core/AssetGraphUpdateMutation.graphql";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const assetsQuery = graphql`
  query AssetGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateAsset: permission(action: "core:asset:create")
        canPublishAssets: permission(action: "core:asset:publish")
        assetListDocument {
          id
          defaultApprovers {
            id
          }
        }
        ...AssetsPageFragment
      }
    }
  }
`;

export const assetNodeQuery = graphql`
  query AssetGraphNodeQuery($assetId: ID!) {
    node(id: $assetId) {
      ... on Asset {
        id
        name
        amount
        assetType
        dataTypesStored
        owner {
          id
          fullName
        }
        thirdParties(first: 50) {
          edges {
            node {
              id
              name
              websiteUrl
              category
            }
          }
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "core:asset:update")
        canDelete: permission(action: "core:asset:delete")
      }
    }
  }
`;

export const createAssetMutation = graphql`
  mutation AssetGraphCreateMutation(
    $input: CreateAssetInput!
    $connections: [ID!]!
  ) {
    createAsset(input: $input) {
      assetEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          amount
          assetType
          dataTypesStored
          owner {
            id
            fullName
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
          createdAt
          canUpdate: permission(action: "core:asset:update")
          canDelete: permission(action: "core:asset:delete")
        }
      }
    }
  }
`;

export const updateAssetMutation = graphql`
  mutation AssetGraphUpdateMutation($input: UpdateAssetInput!) {
    updateAsset(input: $input) {
      asset {
        id
        name
        amount
        assetType
        dataTypesStored
        owner {
          id
          fullName
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
        updatedAt
      }
    }
  }
`;

export const deleteAssetMutation = graphql`
  mutation AssetGraphDeleteMutation(
    $input: DeleteAssetInput!
    $connections: [ID!]!
  ) {
    deleteAsset(input: $input) {
      deletedAssetId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteAsset = (
  asset: { id?: string; name?: string },
  connectionId: string,
) => {
  const [mutate] = useMutation<AssetGraphDeleteMutation>(deleteAssetMutation);
  const confirm = useConfirm();
  const { t } = useTranslation();

  return () => {
    if (!asset.id || !asset.name) {
      return alert(t("assetGraph.errors.deleteMissingIdOrName"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              assetId: asset.id!,
            },
            connections: [connectionId],
          },
        }),
      {
        message: t("assetGraph.deleteConfirmation", { name: asset.name }),
      },
    );
  };
};

export const useCreateAsset = (connectionId: string) => {
  const [mutate, isMutating] = useMutation<AssetGraphCreateMutation>(createAssetMutation);
  const { t } = useTranslation();

  return [
    (input: {
      name: string;
      amount: number;
      assetType: AssetType;
      ownerId: string;
      organizationId: string;
      thirdPartyIds?: string[];
      dataTypesStored: string;
    }) => {
      if (!input.name?.trim()) {
        return alert(t("assetGraph.errors.createNameRequired"));
      }
      if (!input.ownerId) {
        return alert(t("assetGraph.errors.createOwnerRequired"));
      }
      if (!input.organizationId) {
        return alert(t("assetGraph.errors.createOrganizationRequired"));
      }
      if (!input.dataTypesStored) {
        return alert(t("assetGraph.errors.createDataTypesStoredRequired"));
      }

      return promisifyMutation(mutate)({
        variables: {
          input: {
            name: input.name,
            amount: input.amount,
            assetType: input.assetType,
            dataTypesStored: input.dataTypesStored || "",
            ownerId: input.ownerId,
            organizationId: input.organizationId,
            thirdPartyIds: input.thirdPartyIds || [],
          },
          connections: [connectionId],
        },
      });
    },
    isMutating,
  ] as const;
};

export const useUpdateAsset = () => {
  const { t } = useTranslation();
  const [mutate] = useMutation<AssetGraphUpdateMutation>(updateAssetMutation);

  return (input: {
    id: string;
    name?: string;
    amount?: number;
    assetType?: AssetType;
    dataTypesStored?: string;
    ownerId?: string;
    thirdPartyIds?: string[];
  }) => {
    if (!input.id) {
      return alert(t("assetGraph.errors.updateIdRequired"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
