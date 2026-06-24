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
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const ObligationsConnectionKey = "ObligationsPage_obligations";

export const obligationsQuery = graphql`
  query ObligationGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateObligation: permission(action: "core:obligation:create")
        canPublishObligations: permission(action: "core:obligation:publish")
        obligationsDocument {
          id
          defaultApprovers {
            id
          }
        }
        ...ObligationsPageFragment
      }
    }
  }
`;

export const obligationNodeQuery = graphql`
  query ObligationGraphNodeQuery($obligationId: ID!) {
    node(id: $obligationId) {
      ... on Obligation {
        id
        area
        source
        requirement
        actionsToBeImplemented
        regulator
        type
        lastReviewDate
        dueDate
        status
        owner {
          id
          fullName
        }
        organization {
          id
          name
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "core:obligation:update")
        canDelete: permission(action: "core:obligation:delete")
      }
    }
  }
`;

export const createObligationMutation = graphql`
  mutation ObligationGraphCreateMutation(
    $input: CreateObligationInput!
    $connections: [ID!]!
  ) {
    createObligation(input: $input) {
      obligationEdge @prependEdge(connections: $connections) {
        node {
          id
          area
          source
          requirement
          actionsToBeImplemented
          regulator
          type
          lastReviewDate
          dueDate
          status
          owner {
            id
            fullName
          }
          createdAt
          canUpdate: permission(action: "core:obligation:update")
          canDelete: permission(action: "core:obligation:delete")
        }
      }
    }
  }
`;

export const updateObligationMutation = graphql`
  mutation ObligationGraphUpdateMutation($input: UpdateObligationInput!) {
    updateObligation(input: $input) {
      obligation {
        id
        area
        source
        requirement
        actionsToBeImplemented
        regulator
        type
        lastReviewDate
        dueDate
        status
        owner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

export const deleteObligationMutation = graphql`
  mutation ObligationGraphDeleteMutation(
    $input: DeleteObligationInput!
    $connections: [ID!]!
  ) {
    deleteObligation(input: $input) {
      deletedObligationId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteObligation = (
  obligation: { id: string },
  connectionId: string,
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteObligationMutation, {
    successMessage: __("Obligation deleted successfully"),
    errorMessage: __("Failed to delete obligation"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              obligationId: obligation.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: __(
          "This will permanently delete this obligation. This action cannot be undone.",
        ),
      },
    );
  };
};

export const useCreateObligation = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createObligationMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    area?: string;
    source?: string;
    requirement?: string;
    actionsToBeImplemented?: string;
    regulator?: string;
    type: string;
    ownerId: string;
    lastReviewDate?: string;
    dueDate?: string;
    status: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create obligation: organization is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create obligation: owner is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          area: input.area,
          source: input.source,
          requirement: input.requirement,
          actionsToBeImplemented: input.actionsToBeImplemented,
          regulator: input.regulator,
          type: input.type,
          ownerId: input.ownerId,
          lastReviewDate: input.lastReviewDate,
          dueDate: input.dueDate,
          status: input.status || "NON_COMPLIANT",
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateObligation = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateObligationMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    area?: string;
    source?: string;
    requirement?: string;
    actionsToBeImplemented?: string;
    regulator?: string;
    type?: string;
    ownerId?: string;
    lastReviewDate?: string | null;
    dueDate?: string | null;
    status?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update obligation: ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
