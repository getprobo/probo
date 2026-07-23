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

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const dataQuery = graphql`
  query DatumGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateDatum: permission(action: "core:datum:create")
        canPublishData: permission(action: "core:datum:publish")
        dataListDocument {
          id
          defaultApprovers {
            id
          }
        }
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
  const { t } = useTranslation();

  return () => {
    if (!datum.id || !datum.name) {
      return alert(t("datumGraph.errors.deleteMissingIdOrName"));
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
        message: t("datumGraph.deleteConfirmation", { name: datum.name }),
      },
    );
  };
};

export const useCreateDatum = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createDatumMutation);
  const { t } = useTranslation();

  return (input: {
    name: string;
    dataClassification: string;
    ownerId: string;
    organizationId: string;
    thirdPartyIds?: string[];
  }) => {
    if (!input.name?.trim()) {
      return alert(t("datumGraph.errors.createNameRequired"));
    }
    if (!input.ownerId) {
      return alert(t("datumGraph.errors.createOwnerRequired"));
    }
    if (!input.organizationId) {
      return alert(t("datumGraph.errors.createOrganizationRequired"));
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
  const { t } = useTranslation();

  return (input: {
    id: string;
    name?: string;
    dataClassification?: string;
    ownerId?: string;
    thirdPartyIds?: string[];
  }) => {
    if (!input.id) {
      return alert(t("datumGraph.errors.updateMissingId"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
