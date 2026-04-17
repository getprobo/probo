// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const dataQuery = graphql`
  query DatumGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateDatum: permission(action: "core:datum:create")
        canPublishData: permission(action: "core:datum:publish")
        ...DataPageFragment
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
        canUpdate: permission(action: "core:datum:update")
        canDelete: permission(action: "core:datum:delete")
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
          canUpdate: permission(action: "core:datum:update")
          canDelete: permission(action: "core:datum:delete")
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
  connectionId: string,
) => {
  // eslint-disable-next-line relay/generated-typescript-types
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
            "This will permanently delete \"%s\". This action cannot be undone.",
          ),
          datum.name,
        ),
      },
    );
  };
};

export const useCreateDatum = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
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
  // eslint-disable-next-line relay/generated-typescript-types
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
