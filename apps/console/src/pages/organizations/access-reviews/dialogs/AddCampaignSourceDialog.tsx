// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError } from "@probo/helpers";
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

import type { AddCampaignSourceDialogMutation } from "#/__generated__/core/AddCampaignSourceDialogMutation.graphql";
import type { AddCampaignSourceDialogSourcesQuery } from "#/__generated__/core/AddCampaignSourceDialogSourcesQuery.graphql";

const addScopeMutation = graphql`
  mutation AddCampaignSourceDialogMutation(
    $input: AddAccessReviewCampaignSourceInput!
  ) {
    addAccessReviewCampaignSource(input: $input) {
      accessReviewCampaign {
        id
        sources {
          id
          name
          fetchAttempts(first: 1) {
            edges {
              node {
                status
                fetchedAccountsCount
                error
              }
            }
          }
          entries(first: 50) {
            edges {
              node {
                id
                email
                fullName
                roles
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
  query AddCampaignSourceDialogSourcesQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        accessReviewSources(first: 100) {
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
  existingCampaignSourceIds: string[];
};

export function AddCampaignSourceDialog({
  children,
  organizationId,
  campaignId,
  existingCampaignSourceIds,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const ref = useDialogRef();
  const [selectedSourceId, setSelectedSourceId] = useState<string>("");

  const [addCampaignSource, isAdding]
    = useMutation<AddCampaignSourceDialogMutation>(addScopeMutation);

  const onSubmit = () => {
    if (!selectedSourceId) return;

    addCampaignSource({
      variables: {
        input: {
          accessReviewCampaignId: campaignId,
          accessReviewSourceId: selectedSourceId,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to add source"),
              errors,
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
            error,
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
            existingCampaignSourceIds={existingCampaignSourceIds}
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
  existingCampaignSourceIds,
  value,
  onChange,
}: {
  organizationId: string;
  existingCampaignSourceIds: string[];
  value: string;
  onChange: (value: string) => void;
}) {
  const { __ } = useTranslate();
  const data
    = useLazyLoadQuery<AddCampaignSourceDialogSourcesQuery>(
      sourcesQuery,
      { organizationId },
      { fetchPolicy: "network-only" },
    );

  const sources
    = data?.organization?.accessReviewSources?.edges
      ?.map(edge => edge.node)
      .filter(
        (node): node is NonNullable<typeof node> =>
          node !== null && !existingCampaignSourceIds.includes(node.id),
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
