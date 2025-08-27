import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";

export const dataQuery = graphql`
  query DatumGraphListQuery($organizationId: ID!, $snapshotId: ID = null) {
    node(id: $organizationId) {
      ... on Organization {
        ...DataPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const datumNodeQuery = graphql`
  query DatumGraphNodeQuery($dataId: ID!) {
    node(id: $dataId) {
      ... on Datum {
        id
        name
        dataClassification
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
        organization {
          id
        }
        createdAt
        updatedAt
      }
    }
  }
`;

export const createDatumMutation = graphql`
  mutation DatumGraphCreateMutation(
    $input: CreateDatumInput!
    $connections: [ID!]!
  ) {
    createDatum(input: $input) {
      datumEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          dataClassification
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

export const updateDatumMutation = graphql`
  mutation DatumGraphUpdateMutation($input: UpdateDatumInput!) {
    updateDatum(input: $input) {
      datum {
        id
        name
        dataClassification
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

export const deleteDatumMutation = graphql`
  mutation DatumGraphDeleteMutation(
    $input: DeleteDatumInput!
    $connections: [ID!]!
  ) {
    deleteDatum(input: $input) {
      deletedDatumId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteDatum = (
  datum: { id?: string; name?: string },
  connectionId: string
) => {
  const [mutate] = useMutation(deleteDatumMutation);
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!datum.id || !datum.name) {
      return alert(__("Failed to delete data: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              datumId: datum.id!,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete "%s". This action cannot be undone.'
          ),
          datum.name
        ),
      }
    );
  };
};

export const useCreateDatum = (connectionId: string) => {
  const [mutate] = useMutation(createDatumMutation);
  const { __ } = useTranslate();

  return (input: {
    name: string;
    dataClassification: string;
    ownerId: string;
    organizationId: string;
    vendorIds?: string[];
  }) => {
    if (!input.name?.trim()) {
      return alert(__("Failed to create data: name is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create data: owner is required"));
    }
    if (!input.organizationId) {
      return alert(__("Failed to create data: organization is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateDatum = () => {
  const [mutate] = useMutation(updateDatumMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    name?: string;
    dataClassification?: string;
    ownerId?: string;
    vendorIds?: string[];
  }) => {
    if (!input.id) {
      return alert(__("Failed to update data: missing id"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
