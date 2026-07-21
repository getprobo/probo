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
import { useTranslation } from "react-i18next";
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
        accessReviewSources(first: 500) {
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
  const { t } = useTranslation();
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
          accessReviewSourceIds:
            selectedSourceIds.length > 0 ? selectedSourceIds : null,
        },
        connections: [connectionId],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("createAccessReviewCampaignDialog.messages.error"),
            description: formatError(
              t("createAccessReviewCampaignDialog.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("createAccessReviewCampaignDialog.messages.success"),
          description: t("createAccessReviewCampaignDialog.messages.created"),
          variant: "success",
        });
        reset();
        setSelectedSourceIds([]);
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: t("createAccessReviewCampaignDialog.messages.error"),
          description: formatError(
            t("createAccessReviewCampaignDialog.errors.create"),
            error,
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
          items={[
            t("createAccessReviewCampaignDialog.breadcrumb.accessReviews"),
            t("createAccessReviewCampaignDialog.breadcrumb.newCampaign"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={t("createAccessReviewCampaignDialog.fields.name")}
            {...register("name")}
            type="text"
            required
          />
          <Field
            label={t("createAccessReviewCampaignDialog.fields.description")}
            {...register("description")}
            type="textarea"
          />
          <Suspense
            fallback={(
              <div className="text-sm text-txt-tertiary">
                {t("createAccessReviewCampaignDialog.loadingSources")}
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
            {t("createAccessReviewCampaignDialog.actions.create")}
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
  const { t } = useTranslation();
  const data = useLazyLoadQuery<CreateAccessReviewCampaignDialogSourcesQuery>(
    sourcesQuery,
    { organizationId },
    { fetchPolicy: "network-only" },
  );

  const sources
    = data?.organization?.accessReviewSources?.edges
      ?.map(edge => edge.node)
      .filter((node): node is NonNullable<typeof node> => node !== null) ?? [];

  if (sources.length === 0) {
    return (
      <div className="text-sm text-txt-tertiary">
        {t("createAccessReviewCampaignDialog.emptySources")}
      </div>
    );
  }

  return (
    <fieldset>
      <legend className="text-sm font-medium mb-2">
        {t("createAccessReviewCampaignDialog.fields.sources")}
      </legend>
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
