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

import {
  formatDatetime,
  formatError,
  getStatusLabel,
  getStatusOptions,
  getStatusVariant,
  type GraphQLError,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Field,
  IconTrashCan,
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { z } from "zod";

import type { FindingDetailsPageDeleteMutation } from "#/__generated__/core/FindingDetailsPageDeleteMutation.graphql";
import type { FindingDetailsPageQuery } from "#/__generated__/core/FindingDetailsPageQuery.graphql";
import type { FindingDetailsPageUpdateMutation } from "#/__generated__/core/FindingDetailsPageUpdateMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FindingsConnectionKey } from "./FindingsPage";

export const findingDetailsPageQuery = graphql`
  query FindingDetailsPageQuery($findingId: ID!) {
    node(id: $findingId) {
      ... on Finding {
        id
        kind
        referenceId
        description
        source
        identifiedOn
        rootCause
        correctiveAction
        dueDate
        status
        priority
        effectivenessCheck
        owner {
          id
        }
        canUpdate: permission(action: "core:finding:update")
        canDelete: permission(action: "core:finding:delete")
      }
    }
  }
`;

const updateFindingMutation = graphql`
  mutation FindingDetailsPageUpdateMutation($input: UpdateFindingInput!) {
    updateFinding(input: $input) {
      finding {
        id
        kind
        referenceId
        description
        source
        identifiedOn
        rootCause
        correctiveAction
        dueDate
        status
        priority
        effectivenessCheck
        owner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

const deleteFindingMutation = graphql`
  mutation FindingDetailsPageDeleteMutation(
    $input: DeleteFindingInput!
    $connections: [ID!]!
  ) {
    deleteFinding(input: $input) {
      deletedFindingId @deleteEdge(connections: $connections)
    }
  }
`;

const updateFindingSchema = z.object({
  description: z.string().optional(),
  source: z.string().optional(),
  identifiedOn: z.string().optional(),
  dueDate: z.string().optional(),
  rootCause: z.string().optional(),
  correctiveAction: z.string().optional(),
  effectivenessCheck: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED", "MITIGATED", "FALSE_POSITIVE", "RISK_ACCEPTED"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH"]),
  ownerId: z.string().nullable().optional(),
});

type Props = {
  queryRef: PreloadedQuery<FindingDetailsPageQuery>;
};

function getKindLabel(kind: string, __: (s: string) => string): string {
  switch (kind) {
    case "MINOR_NONCONFORMITY":
      return __("Minor nonconformity");
    case "MAJOR_NONCONFORMITY":
      return __("Major nonconformity");
    case "OBSERVATION":
      return __("Observation");
    case "EXCEPTION":
      return __("Exception");
    default:
      return kind;
  }
}

export default function FindingDetailsPage(props: Props) {
  const { node: finding } = usePreloadedQuery<FindingDetailsPageQuery>(
    findingDetailsPageQuery,
    props.queryRef,
  );
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();

  const [updateFinding] = useMutation<FindingDetailsPageUpdateMutation>(updateFindingMutation);
  const [deleteFinding] = useMutation<FindingDetailsPageDeleteMutation>(deleteFindingMutation);

  const connections = [
    ConnectionHandler.getConnectionID(
      organizationId,
      FindingsConnectionKey,
      {
        filter: {
          kind: null,
          status: null,
          priority: null,
          ownerId: null,
        },
      },
    ),
    ConnectionHandler.getConnectionID(
      organizationId,
      FindingsConnectionKey,
      {
        filter: {
          kind: finding.kind,
          status: null,
          priority: null,
          ownerId: null,
        },
      },
    ),
  ];

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteFinding({
            variables: {
              input: { findingId: finding.id! },
              connections,
            },
            onCompleted(_, error) {
              if (error) {
                toast({
                  title: __("Error"),
                  description: formatError(
                    __("Failed to delete finding"),
                    error as GraphQLError[],
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Finding deleted successfully"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete finding"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the finding %s. This action cannot be undone.",
          ),
          finding.referenceId!,
        ),
      },
    );
  };

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateFindingSchema, {
      defaultValues: {
        description: finding.description || "",
        source: finding.source || "",
        identifiedOn: finding.identifiedOn?.split("T")[0] || "",
        dueDate: finding.dueDate?.split("T")[0] || "",
        rootCause: finding.rootCause || "",
        correctiveAction: finding.correctiveAction || "",
        effectivenessCheck: finding.effectivenessCheck || "",
        status: finding.status || "OPEN",
        priority: finding.priority || "MEDIUM",
        ownerId: finding.owner?.id ?? null,
      },
    });

  const onSubmit = handleSubmit((formData) => {
    if (!finding.id) return;

    updateFinding({
      variables: {
        input: {
          id: finding.id,
          description: formData.description || undefined,
          source: formData.source || undefined,
          identifiedOn: formatDatetime(formData.identifiedOn) ?? null,
          dueDate: formatDatetime(formData.dueDate) ?? null,
          rootCause: formData.rootCause || undefined,
          correctiveAction: formData.correctiveAction || undefined,
          effectivenessCheck: formData.effectivenessCheck || undefined,
          status: formData.status,
          priority: formData.priority,
          ownerId: formData.ownerId || undefined,
        },
      },
      onCompleted() {
        reset(formData);
        toast({
          title: __("Success"),
          description: __("Finding updated successfully"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update finding"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  });

  const statusOptions = getStatusOptions(__).filter(
    opt => opt.value !== "RISK_ACCEPTED",
  );

  const priorityOptions = [
    { value: "LOW", label: __("Low") },
    { value: "MEDIUM", label: __("Medium") },
    { value: "HIGH", label: __("High") },
  ];

  const breadcrumbFindingsUrl = `/organizations/${organizationId}/findings`;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Findings"),
            to: breadcrumbFindingsUrl,
          },
          {
            label: finding.referenceId || __("Unknown Finding"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl font-semibold">
            {finding.referenceId}
          </div>
          <Badge variant="neutral">
            {getKindLabel(finding.kind || "", __)}
          </Badge>
          <Badge variant={getStatusVariant(finding.status || "OPEN")}>
            {getStatusLabel(finding.status || "OPEN")}
          </Badge>
          <Badge
            variant={
              finding.priority === "HIGH"
                ? "danger"
                : finding.priority === "MEDIUM"
                  ? "warning"
                  : "success"
            }
          >
            {finding.priority === "HIGH"
              ? __("High")
              : finding.priority === "MEDIUM"
                ? __("Medium")
                : __("Low")}
          </Badge>
        </div>
        <ActionDropdown variant="secondary">
          {finding.canDelete && (
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={e => void onSubmit(e)} className="space-y-6">
            <Field label={__("Description")}>
              <Textarea
                {...register("description")}
                placeholder={__("Enter description")}
                rows={3}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field
                label={__("Source")}
                error={formState.errors.source?.message}
              >
                <Input
                  {...register("source")}
                  placeholder={__("Enter source")}
                />
              </Field>

              <PeopleSelectField
                organizationId={organizationId}
                control={control}
                name="ownerId"
                label={__("Owner")}
                error={formState.errors.ownerId?.message}
                optional
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <ControlledField
                control={control}
                name="status"
                type="select"
                label={__("Status")}
                required
              >
                {statusOptions.map(option => (
                  <Option key={option.value} value={option.value}>
                    {option.label}
                  </Option>
                ))}
              </ControlledField>

              <Controller
                control={control}
                name="priority"
                render={({ field }) => (
                  <div>
                    <Label>
                      {__("Priority")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {priorityOptions.map(option => (
                        <Option key={option.value} value={option.value}>
                          {option.label}
                        </Option>
                      ))}
                    </Select>
                    {formState.errors.priority?.message && (
                      <div className="text-red-500 text-sm mt-1">
                        {formState.errors.priority.message}
                      </div>
                    )}
                  </div>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field label={__("Date Identified")}>
                <Input
                  {...register("identifiedOn")}
                  type="date"
                />
              </Field>

              <Field label={__("Due Date")}>
                <Input
                  {...register("dueDate")}
                  type="date"
                />
              </Field>
            </div>

            <Field label={__("Root Cause")}>
              <Textarea
                {...register("rootCause")}
                placeholder={__("Enter root cause")}
                rows={3}
              />
            </Field>

            <Field label={__("Corrective Action")}>
              <Textarea
                {...register("correctiveAction")}
                placeholder={__("Enter corrective action")}
                rows={3}
              />
            </Field>

            <Field label={__("Effectiveness Check")}>
              <Textarea
                {...register("effectivenessCheck")}
                placeholder={__("Enter effectiveness check details")}
                rows={3}
              />
            </Field>

            <div className="flex justify-end">
              {formState.isDirty
                && finding.canUpdate && (
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting ? __("Updating...") : __("Update")}
                </Button>
              )}
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
