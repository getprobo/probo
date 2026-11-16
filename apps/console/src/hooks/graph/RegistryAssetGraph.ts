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
  RegistryAssetGraphCreateMutation,
  RegistryAssetType,
} from "#/__generated__/core/RegistryAssetGraphCreateMutation.graphql";
import type { RegistryAssetGraphDeleteMutation } from "#/__generated__/core/RegistryAssetGraphDeleteMutation.graphql";
import type { RegistryAssetGraphUpdateMutation } from "#/__generated__/core/RegistryAssetGraphUpdateMutation.graphql";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const registryAssetsQuery = graphql`
  query RegistryAssetGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateRegistryAsset: permission(action: "core:registry-asset:create")
        canPublishAssets: permission(action: "core:registry-asset:publish")
        registryAssetListDocument {
          id
          defaultApprovers {
            id
          }
        }
        ...RegistryAssetsPageFragment
      }
    }
  }
`;

export const registryAssetNodeQuery = graphql`
  query RegistryAssetGraphNodeQuery($registryAssetId: ID!) {
    node(id: $registryAssetId) {
      ... on RegistryAsset {
        id
        name
        amount
        registryAssetType
        dataTypesStored
        owner {
          id
          fullName
          emailAddress
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
        canUpdate: permission(action: "core:registry-asset:update")
        canDelete: permission(action: "core:registry-asset:delete")
      }
    }
  }
`;

export const createRegistryAssetMutation = graphql`
  mutation RegistryAssetGraphCreateMutation(
    $input: CreateRegistryAssetInput!
    $connections: [ID!]!
  ) {
    createRegistryAsset(input: $input) {
      registryAssetEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          amount
          registryAssetType
          dataTypesStored
          owner {
            id
            fullName
            emailAddress
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
          canUpdate: permission(action: "core:registry-asset:update")
          canDelete: permission(action: "core:registry-asset:delete")
        }
      }
    }
  }
`;

export const updateRegistryAssetMutation = graphql`
  mutation RegistryAssetGraphUpdateMutation($input: UpdateRegistryAssetInput!) {
    updateRegistryAsset(input: $input) {
      registryAsset {
        id
        name
        amount
        registryAssetType
        dataTypesStored
        owner {
          id
          fullName
          emailAddress
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

export const deleteRegistryAssetMutation = graphql`
  mutation RegistryAssetGraphDeleteMutation(
    $input: DeleteRegistryAssetInput!
    $connections: [ID!]!
  ) {
    deleteRegistryAsset(input: $input) {
      deletedRegistryAssetId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteRegistryAsset = (
  asset: { id?: string; name?: string },
  connectionId: string,
) => {
  const [mutate] = useMutation<RegistryAssetGraphDeleteMutation>(deleteRegistryAssetMutation);
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
              registryAssetId: asset.id!,
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

export const useCreateRegistryAsset = (connectionId: string) => {
  const [mutate, isMutating] = useMutation<RegistryAssetGraphCreateMutation>(createRegistryAssetMutation);
  const { t } = useTranslation();

  return [
    (input: {
      name: string;
      amount: number;
      registryAssetType: RegistryAssetType;
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
            registryAssetType: input.registryAssetType,
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

export const useUpdateRegistryAsset = () => {
  const { t } = useTranslation();
  const [mutate] = useMutation<RegistryAssetGraphUpdateMutation>(updateRegistryAssetMutation);

  return (input: {
    id: string;
    name?: string;
    amount?: number;
    registryAssetType?: RegistryAssetType;
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
