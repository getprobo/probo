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

import {
  formatDatetime,
  formatError,
  getStatusVariant,
} from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
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

export default function FindingDetailsPage(props: Props) {
  const { node: finding } = usePreloadedQuery<FindingDetailsPageQuery>(
    findingDetailsPageQuery,
    props.queryRef,
  );
  const { t } = useTranslation();
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
                  title: t("findingDetails.errors.title"),
                  description: formatError(
                    t("findingDetails.errors.delete"),
                    error,
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: t("findingDetails.messages.successTitle"),
                  description: t("findingDetails.messages.deleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("findingDetails.errors.title"),
                description: formatError(
                  t("findingDetails.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("findingDetails.deleteConfirmation", {
          referenceId: finding.referenceId,
        }),
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
          title: t("findingDetails.messages.successTitle"),
          description: t("findingDetails.messages.updated"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("findingDetails.errors.title"),
          description: formatError(
            t("findingDetails.errors.update"),
            error,
          ),
          variant: "error",
        });
      },
    });
  });

  const statusOptions = ["OPEN", "IN_PROGRESS", "CLOSED", "MITIGATED", "FALSE_POSITIVE"] as const;

  const priorityOptions = [
    { value: "LOW", label: t("findingDetails.priority.low") },
    { value: "MEDIUM", label: t("findingDetails.priority.medium") },
    { value: "HIGH", label: t("findingDetails.priority.high") },
  ];

  const breadcrumbFindingsUrl = `/organizations/${organizationId}/findings`;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("findingDetails.breadcrumbs.findings"),
            to: breadcrumbFindingsUrl,
          },
          {
            label: finding.referenceId || t("findingDetails.unknown"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl font-semibold">
            {finding.referenceId}
          </div>
          <Badge variant="neutral">
            {t(`findingDetails.kinds.${(finding.kind || "").toLowerCase()}`)}
          </Badge>
          <Badge variant={getStatusVariant(finding.status || "OPEN")}>
            {t(`findingDetails.status.${(finding.status || "OPEN").toLowerCase()}`)}
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
              ? t("findingDetails.priority.high")
              : finding.priority === "MEDIUM"
                ? t("findingDetails.priority.medium")
                : t("findingDetails.priority.low")}
          </Badge>
        </div>
        <ActionDropdown variant="secondary">
          {finding.canDelete && (
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {t("findingDetails.actions.delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={e => void onSubmit(e)} className="space-y-6">
            <Field label={t("findingDetails.fields.description")}>
              <Textarea
                {...register("description")}
                placeholder={t("findingDetails.placeholders.description")}
                rows={3}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field
                label={t("findingDetails.fields.source")}
                error={formState.errors.source?.message}
              >
                <Input
                  {...register("source")}
                  placeholder={t("findingDetails.placeholders.source")}
                />
              </Field>

              <PeopleSelectField
                organizationId={organizationId}
                control={control}
                name="ownerId"
                label={t("findingDetails.fields.owner")}
                error={formState.errors.ownerId?.message}
                optional
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <ControlledField
                control={control}
                name="status"
                type="select"
                label={t("findingDetails.fields.status")}
                required
              >
                {statusOptions.map(status => (
                  <Option key={status} value={status}>
                    {t(`findingDetails.status.${status.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledField>

              <Controller
                control={control}
                name="priority"
                render={({ field }) => (
                  <div>
                    <Label>
                      {t("findingDetails.fields.priority")}
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
              <Field label={t("findingDetails.fields.dateIdentified")}>
                <Input
                  {...register("identifiedOn")}
                  type="date"
                />
              </Field>

              <Field label={t("findingDetails.fields.dueDate")}>
                <Input
                  {...register("dueDate")}
                  type="date"
                />
              </Field>
            </div>

            <Field label={t("findingDetails.fields.rootCause")}>
              <Textarea
                {...register("rootCause")}
                placeholder={t("findingDetails.placeholders.rootCause")}
                rows={3}
              />
            </Field>

            <Field label={t("findingDetails.fields.correctiveAction")}>
              <Textarea
                {...register("correctiveAction")}
                placeholder={t("findingDetails.placeholders.correctiveAction")}
                rows={3}
              />
            </Field>

            <Field label={t("findingDetails.fields.effectivenessCheck")}>
              <Textarea
                {...register("effectivenessCheck")}
                placeholder={t("findingDetails.placeholders.effectivenessCheck")}
                rows={3}
              />
            </Field>

            <div className="flex justify-end">
              {formState.isDirty
                && finding.canUpdate && (
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting
                    ? t("findingDetails.actions.updating")
                    : t("findingDetails.actions.update")}
                </Button>
              )}
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
