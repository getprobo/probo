import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";

export const assetsQuery = graphql`
  query AssetGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...AssetsPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const assetNodeQuery = graphql`
  query AssetGraphNodeQuery($assetId: ID!) {
    node(id: $assetId) {
      ... on Asset {
        id
        snapshotId
        name
        amount
        criticity
        assetType
        dataTypesStored
        owner {
          id
          fullName
        }
        vendors(first: 50) {
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
      assetEdge @prependEdge(connections: $connections) {
        node {
          id
          snapshotId
          name
          amount
          criticity
          assetType
          dataTypesStored
          owner {
            id
            fullName
          }
          vendors(first: 50) {
            edges {
              node {
                id
                name
                websiteUrl
              }
            }
          }
          createdAt
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
        snapshotId
        name
        amount
        criticity
        assetType
        dataTypesStored
        owner {
          id
          fullName
        }
        vendors(first: 50) {
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
  connectionId: string
) => {
  const [mutate] = useMutation(deleteAssetMutation);
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!asset.id || !asset.name) {
      return alert(__("Failed to delete asset: missing id or name"));
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
        message: sprintf(
          __(
            'This will permanently delete "%s". This action cannot be undone.'
          ),
          asset.name
        ),
      }
    );
  };
};

export const useCreateAsset = (connectionId: string) => {
  const [mutate] = useMutation(createAssetMutation);
  const { __ } = useTranslate();

  return (input: {
    name: string;
    amount: number;
    criticity: string;
    assetType: string;
    ownerId: string;
    organizationId: string;
    vendorIds?: string[];
    dataTypesStored: string;
  }) => {
    if (!input.name?.trim()) {
      return alert(__("Failed to create asset: name is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create asset: owner is required"));
    }
    if (!input.organizationId) {
      return alert(__("Failed to create asset: organization is required"));
    }
    if (!input.dataTypesStored) {
      return alert(__("Failed to create asset: data types stored is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          name: input.name,
          amount: input.amount,
          criticity: input.criticity,
          assetType: input.assetType,
          dataTypesStored: input.dataTypesStored || "",
          ownerId: input.ownerId,
          organizationId: input.organizationId,
          vendorIds: input.vendorIds || [],
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateAsset = () => {
  const [mutate] = useMutation(updateAssetMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    name?: string;
    amount?: number;
    criticity?: string;
    assetType?: string;
    dataTypesStored?: string;
    ownerId?: string;
    vendorIds?: string[];
  }) => {
    if (!input.id) {
      return alert(__("Failed to update asset: asset ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
