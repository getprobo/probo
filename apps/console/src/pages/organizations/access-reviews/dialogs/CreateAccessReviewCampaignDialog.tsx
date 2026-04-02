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
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, Suspense, useState } from "react";
import { graphql, useLazyLoadQuery, useMutation } from "react-relay";
import { z } from "zod";

import type { CreateAccessReviewCampaignDialogMutation } from "#/__generated__/core/CreateAccessReviewCampaignDialogMutation.graphql";
import type { CreateAccessReviewCampaignDialogSourcesQuery } from "#/__generated__/core/CreateAccessReviewCampaignDialogSourcesQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const createCampaignMutation = graphql`
  mutation CreateAccessReviewCampaignDialogMutation(
    $input: CreateAccessReviewCampaignInput!
    $connections: [ID!]!
  ) {
    createAccessReviewCampaign(input: $input) {
      accessReviewCampaignEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          status
          createdAt
        }
      }
    }
  }
`;

const sourcesQuery = graphql`
  query CreateAccessReviewCampaignDialogSourcesQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        accessSources(first: 500) {
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

const schema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
});

type Props = {
  children: ReactNode;
  organizationId: string;
  connectionId: string;
};

export function CreateAccessReviewCampaignDialog({
  children,
  organizationId,
  connectionId,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const ref = useDialogRef();
  const [selectedSourceIds, setSelectedSourceIds] = useState<string[]>([]);
  const { register, handleSubmit, reset, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
        description: "",
      },
    },
  );

  const [createCampaign, isCreating]
    = useMutation<CreateAccessReviewCampaignDialogMutation>(
      createCampaignMutation,
    );

  const toggleSource = (sourceId: string) => {
    setSelectedSourceIds(prev =>
      prev.includes(sourceId)
        ? prev.filter(id => id !== sourceId)
        : [...prev, sourceId],
    );
  };

  const onSubmit = (data: z.infer<typeof schema>) => {
    createCampaign({
      variables: {
        input: {
          organizationId,
          name: data.name,
          description: data.description || null,
          accessSourceIds:
            selectedSourceIds.length > 0 ? selectedSourceIds : null,
        },
        connections: [connectionId],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to create campaign"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Campaign created successfully."),
          variant: "success",
        });
        reset();
        setSelectedSourceIds([]);
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create campaign"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleClose = () => {
    reset();
    setSelectedSourceIds([]);
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      onClose={handleClose}
      title={(
        <Breadcrumb
          items={[__("Access Reviews"), __("New Campaign")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />
          <Field
            label={__("Description")}
            {...register("description")}
            type="textarea"
          />
          <Suspense
            fallback={(
              <div className="text-sm text-txt-tertiary">
                {__("Loading sources...")}
              </div>
            )}
          >
            <SourceSelector
              organizationId={organizationId}
              selectedSourceIds={selectedSourceIds}
              onToggle={toggleSource}
            />
          </Suspense>
        </DialogContent>
        <DialogFooter>
          <Button disabled={isCreating || formState.isSubmitting} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function SourceSelector({
  organizationId,
  selectedSourceIds,
  onToggle,
}: {
  organizationId: string;
  selectedSourceIds: string[];
  onToggle: (sourceId: string) => void;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<CreateAccessReviewCampaignDialogSourcesQuery>(
    sourcesQuery,
    { organizationId },
    { fetchPolicy: "network-only" },
  );

  const sources
    = data?.organization?.accessSources?.edges
      ?.map(edge => edge.node)
      .filter((node): node is NonNullable<typeof node> => node !== null) ?? [];

  if (sources.length === 0) {
    return (
      <div className="text-sm text-txt-tertiary">
        {__("No sources available. Add sources in the Sources tab first.")}
      </div>
    );
  }

  return (
    <fieldset>
      <legend className="text-sm font-medium mb-2">{__("Sources")}</legend>
      <div className="space-y-2">
        {sources.map(source => (
          <label
            key={source.id}
            className="flex items-center gap-2 cursor-pointer"
          >
            <Checkbox
              checked={selectedSourceIds.includes(source.id)}
              onChange={() => onToggle(source.id)}
            />
            <span className="text-sm">{source.name}</span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
