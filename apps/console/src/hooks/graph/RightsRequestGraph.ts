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

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const RightsRequestsConnectionKey = "RightsRequestsPage_rightsRequests";

export const rightsRequestsQuery = graphql`
  query RightsRequestGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateRightsRequest: permission(action: "core:rights-request:create")
        ...RightsRequestsPageFragment
      }
    }
  }
`;

export const rightsRequestNodeQuery = graphql`
  query RightsRequestGraphNodeQuery($rightsRequestId: ID!) {
    node(id: $rightsRequestId) {
      ... on RightsRequest {
        id
        requestType
        requestState
        dataSubject
        contact
        details
        deadline
        actionTaken
        canUpdate: permission(action: "core:rights-request:update")
        canDelete: permission(action: "core:rights-request:delete")
        organization {
          id
          name
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "core:rights-request:update")
        canDelete: permission(action: "core:rights-request:delete")
      }
    }
  }
`;

export const createRightsRequestMutation = graphql`
  mutation RightsRequestGraphCreateMutation(
    $input: CreateRightsRequestInput!
    $connections: [ID!]!
  ) {
    createRightsRequest(input: $input) {
      rightsRequestEdge @prependEdge(connections: $connections) {
        node {
          id
          canDelete: permission(action: "core:rights-request:delete")
          canUpdate: permission(action: "core:rights-request:update")
          requestType
          requestState
          dataSubject
          contact
          details
          deadline
          actionTaken
          createdAt
        }
      }
    }
  }
`;

export const updateRightsRequestMutation = graphql`
  mutation RightsRequestGraphUpdateMutation($input: UpdateRightsRequestInput!) {
    updateRightsRequest(input: $input) {
      rightsRequest {
        id
        requestType
        requestState
        dataSubject
        contact
        details
        deadline
        actionTaken
        updatedAt
      }
    }
  }
`;

export const deleteRightsRequestMutation = graphql`
  mutation RightsRequestGraphDeleteMutation(
    $input: DeleteRightsRequestInput!
    $connections: [ID!]!
  ) {
    deleteRightsRequest(input: $input) {
      deletedRightsRequestId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteRightsRequest = (
  request: { id: string },
  connectionId: string,
) => {
  const { t } = useTranslation();
  const [mutate] = useMutationWithToasts(deleteRightsRequestMutation, {
    successMessage: t("rightsRequestGraph.messages.deleted"),
    errorMessage: t("rightsRequestGraph.errors.delete"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              rightsRequestId: request.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: t("rightsRequestGraph.deleteConfirmation"),
      },
    );
  };
};

export const useCreateRightsRequest = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createRightsRequestMutation);
  const { t } = useTranslation();

  return (input: {
    organizationId: string;
    requestType: string;
    requestState: string;
    dataSubject?: string;
    contact?: string;
    details?: string;
    deadline?: string;
    actionTaken?: string;
  }) => {
    if (!input.organizationId) {
      return alert(
        t("rightsRequestGraph.errors.createOrganizationRequired"),
      );
    }
    if (!input.requestType) {
      return alert(
        t("rightsRequestGraph.errors.createRequestTypeRequired"),
      );
    }
    if (!input.requestState) {
      return alert(
        t("rightsRequestGraph.errors.createRequestStateRequired"),
      );
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          requestType: input.requestType,
          requestState: input.requestState,
          dataSubject: input.dataSubject,
          contact: input.contact,
          details: input.details,
          deadline: input.deadline,
          actionTaken: input.actionTaken,
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateRightsRequest = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateRightsRequestMutation);
  const { t } = useTranslation();

  return (input: {
    id: string;
    requestType?: string;
    requestState?: string;
    dataSubject?: string;
    contact?: string;
    details?: string;
    deadline?: string | null;
    actionTaken?: string;
  }) => {
    if (!input.id) {
      return alert(t("rightsRequestGraph.errors.updateIdRequired"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
