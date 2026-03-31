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

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const SnapshotsConnectionKey = "SnapshotsPage_snapshots";

export const snapshotsQuery = graphql`
  query SnapshotGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        canCreateSnapshot: permission(action: "core:snapshot:create")
        ...SnapshotsPageFragment
      }
    }
  }
`;

export const snapshotNodeQuery = graphql`
  query SnapshotGraphNodeQuery($snapshotId: ID!) {
    node(id: $snapshotId) {
      ... on Snapshot {
        id
        name
        description
        type
        organization {
          id
          name
        }
        createdAt
      }
    }
  }
`;

export const createSnapshotMutation = graphql`
  mutation SnapshotGraphCreateMutation(
    $input: CreateSnapshotInput!
    $connections: [ID!]!
  ) {
    createSnapshot(input: $input) {
      snapshotEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          type
          createdAt
        }
      }
    }
  }
`;

export const deleteSnapshotMutation = graphql`
  mutation SnapshotGraphDeleteMutation(
    $input: DeleteSnapshotInput!
    $connections: [ID!]!
  ) {
    deleteSnapshot(input: $input) {
      deletedSnapshotId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteSnapshot = (
  snapshot: { id: string; name: string },
  connectionId: string,
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteSnapshotMutation, {
    successMessage: __("Snapshot deleted successfully"),
    errorMessage: __("Failed to delete snapshot"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              snapshotId: snapshot.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the snapshot %s. This action cannot be undone.",
          ),
          snapshot.name,
        ),
      },
    );
  };
};

export const useCreateSnapshot = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createSnapshotMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    name: string;
    description?: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create snapshot: organization is required"));
    }
    if (!input.name) {
      return alert(__("Failed to create snapshot: name is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          name: input.name,
          description: input.description,
        },
        connections: [connectionId],
      },
    });
  };
};
