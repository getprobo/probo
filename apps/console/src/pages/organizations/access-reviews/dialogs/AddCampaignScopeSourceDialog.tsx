// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, Suspense, useState } from "react";
import { graphql, useLazyLoadQuery, useMutation } from "react-relay";

import type { AddCampaignScopeSourceDialogMutation } from "#/__generated__/core/AddCampaignScopeSourceDialogMutation.graphql";
import type { AddCampaignScopeSourceDialogSourcesQuery } from "#/__generated__/core/AddCampaignScopeSourceDialogSourcesQuery.graphql";

const addScopeMutation = graphql`
  mutation AddCampaignScopeSourceDialogMutation(
    $input: AddAccessReviewCampaignScopeSourceInput!
  ) {
    addAccessReviewCampaignScopeSource(input: $input) {
      accessReviewCampaign {
        id
        scopeSources {
          id
          name
          fetchStatus
          fetchedAccountsCount
          entries(first: 50) {
            edges {
              node {
                id
                email
                fullName
                role
                isAdmin
                mfaStatus
                lastLogin
                decision
                flags
              }
            }
            pageInfo {
              hasNextPage
            }
          }
        }
      }
    }
  }
`;

const sourcesQuery = graphql`
  query AddCampaignScopeSourceDialogSourcesQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        accessSources(first: 100) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  organizationId: string;
  campaignId: string;
  existingScopeSourceIds: string[];
};

export function AddCampaignScopeSourceDialog({
  children,
  organizationId,
  campaignId,
  existingScopeSourceIds,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const ref = useDialogRef();
  const [selectedSourceId, setSelectedSourceId] = useState<string>("");

  const [addScopeSource, isAdding]
    = useMutation<AddCampaignScopeSourceDialogMutation>(addScopeMutation);

  const onSubmit = () => {
    if (!selectedSourceId) return;

    addScopeSource({
      variables: {
        input: {
          accessReviewCampaignId: campaignId,
          accessSourceId: selectedSourceId,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to add source"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Source added to campaign."),
          variant: "success",
        });
        setSelectedSourceId("");
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to add source"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={
        <Breadcrumb items={[__("Campaign"), __("Add Source")]} />
      }
    >
      <DialogContent padded className="space-y-4">
        <Suspense
          fallback={
            <Select disabled placeholder={__("Loading...")} />
          }
        >
          <SourceSelect
            organizationId={organizationId}
            existingScopeSourceIds={existingScopeSourceIds}
            value={selectedSourceId}
            onChange={setSelectedSourceId}
          />
        </Suspense>
      </DialogContent>
      <DialogFooter>
        <Button
          disabled={isAdding || !selectedSourceId}
          onClick={onSubmit}
        >
          {__("Add")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}

function SourceSelect({
  organizationId,
  existingScopeSourceIds,
  value,
  onChange,
}: {
  organizationId: string;
  existingScopeSourceIds: string[];
  value: string;
  onChange: (value: string) => void;
}) {
  const { __ } = useTranslate();
  const data
    = useLazyLoadQuery<AddCampaignScopeSourceDialogSourcesQuery>(
      sourcesQuery,
      { organizationId },
      { fetchPolicy: "network-only" },
    );

  const sources
    = data?.organization?.accessSources?.edges
      ?.map(edge => edge.node)
      .filter(
        (node): node is NonNullable<typeof node> =>
          node !== null && !existingScopeSourceIds.includes(node.id),
      ) ?? [];

  if (sources.length === 0) {
    return (
      <p className="text-sm text-txt-tertiary">
        {__("All available sources are already added to this campaign.")}
      </p>
    );
  }

  return (
    <Select
      placeholder={__("Select a source")}
      value={value}
      onValueChange={onChange}
    >
      {sources.map(source => (
        <Option key={source.id} value={source.id}>
          {source.name}
        </Option>
      ))}
    </Select>
  );
}
