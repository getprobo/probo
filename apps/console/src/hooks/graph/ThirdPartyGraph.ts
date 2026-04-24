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
import { useMemo } from "react";
import { useLazyLoadQuery, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ThirdPartyGraphCreateMutation } from "#/__generated__/core/ThirdPartyGraphCreateMutation.graphql";
import type { ThirdPartyGraphDeleteMutation } from "#/__generated__/core/ThirdPartyGraphDeleteMutation.graphql";
import type { ThirdPartyGraphSelectQuery } from "#/__generated__/core/ThirdPartyGraphSelectQuery.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

const createThirdPartyMutation = graphql`
  mutation ThirdPartyGraphCreateMutation(
    $input: CreateThirdPartyInput!
    $connections: [ID!]!
  ) {
    createThirdParty(input: $input) {
      thirdPartyEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          websiteUrl
          createdAt
          updatedAt
          canUpdate: permission(action: "core:thirdParty:update")
          canDelete: permission(action: "core:thirdParty:delete")
        }
      }
    }
  }
`;

export function useCreateThirdPartyMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<ThirdPartyGraphCreateMutation>(
    createThirdPartyMutation,
    {
      successMessage: __("Third party created successfully."),
      errorMessage: __("Failed to create third party"),
    },
  );
}

const deleteThirdPartyMutation = graphql`
  mutation ThirdPartyGraphDeleteMutation(
    $input: DeleteThirdPartyInput!
    $connections: [ID!]!
  ) {
    deleteThirdParty(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteThirdParty = (
  thirdParty: { id?: string; name?: string },
  connectionId: string,
) => {
  const [mutate] = useMutation<ThirdPartyGraphDeleteMutation>(deleteThirdPartyMutation);
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!thirdParty.id || !thirdParty.name) {
      return alert(__("Failed to delete third party: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              thirdPartyId: thirdParty.id!,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete thirdParty \"%s\". This action cannot be undone.",
          ),
          thirdParty.name,
        ),
      },
    );
  };
};

export const thirdPartyConnectionKey = "ThirdPartiesPage_thirdParties";

export const thirdPartiesQuery = graphql`
  query ThirdPartyGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        id
        canCreateThirdParty: permission(action: "core:thirdParty:create")
        ...ThirdPartyGraphPaginatedFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const paginatedThirdPartiesFragment = graphql`
  fragment ThirdPartyGraphPaginatedFragment on Organization
  @refetchable(queryName: "ThirdPartiesListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    thirdParties(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { snapshotId: $snapshotId }
    ) @connection(key: "ThirdPartiesListQuery_thirdParties", filters: ["filter"]) {
      __id
      edges {
        node {
          id
          snapshotId
          name
          websiteUrl
          updatedAt
          riskAssessments(
            first: 1
            orderBy: { direction: DESC, field: CREATED_AT }
          ) {
            edges {
              node {
                id
                createdAt
                expiresAt
                dataSensitivity
                businessImpact
              }
            }
          }
          canUpdate: permission(action: "core:thirdParty:update")
          canDelete: permission(action: "core:thirdParty:delete")
        }
      }
    }
  }
`;

export const thirdPartyNodeQuery = graphql`
  query ThirdPartyGraphNodeQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      id
      ... on ThirdParty {
        snapshotId
        name
        websiteUrl
        canAssess: permission(action: "core:thirdParty:assess")
        canUpdate: permission(action: "core:thirdParty:update")
        canDelete: permission(action: "core:thirdParty:delete")
        canUploadComplianceReport: permission(
          action: "core:thirdParty-compliance-report:upload"
        )
        canCreateRiskAssessment: permission(
          action: "core:thirdParty-risk-assessment:create"
        )
        canCreateContact: permission(action: "core:thirdParty-contact:create")
        canCreateService: permission(action: "core:thirdParty-service:create")
        canUploadBAA: permission(
          action: "core:thirdParty-business-associate-agreement:upload"
        )
        canUploadDPA: permission(
          action: "core:thirdParty-data-privacy-agreement:upload"
        )
        ...useThirdPartyFormFragment
        ...ThirdPartyComplianceTabFragment
        ...ThirdPartyContactsTabFragment
        ...ThirdPartyServicesTabFragment
        ...ThirdPartyRiskAssessmentTabFragment
        ...ThirdPartyOverviewTabBusinessAssociateAgreementFragment
        ...ThirdPartyOverviewTabDataPrivacyAgreementFragment
      }
    }
    viewer {
      id
    }
  }
`;

export const thirdPartiesSelectQuery = graphql`
  query ThirdPartyGraphSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        thirdParties(first: 100, orderBy: { direction: ASC, field: NAME }) {
          edges {
            node {
              id
              name
              websiteUrl
            }
          }
        }
      }
    }
  }
`;

export function useThirdParties(organizationId: string) {
  const data = useLazyLoadQuery<ThirdPartyGraphSelectQuery>(
    thirdPartiesSelectQuery,
    {
      organizationId: organizationId,
    },
    { fetchPolicy: "network-only" },
  );
  return useMemo(() => {
    return data.organization?.thirdParties?.edges.map(edge => edge.node) ?? [];
  }, [data]);
}
