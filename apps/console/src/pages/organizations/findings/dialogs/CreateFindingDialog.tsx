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
} from "@probo/helpers";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";
import { z } from "zod";

import type { CreateFindingDialogMutation } from "#/__generated__/core/CreateFindingDialogMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const createFindingMutation = graphql`
  mutation CreateFindingDialogMutation(
    $input: CreateFindingInput!
    $connections: [ID!]!
  ) {
    createFinding(input: $input) {
      findingEdge @prependEdge(connections: $connections) {
        node {
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
          createdAt
          canUpdate: permission(action: "core:finding:update")
          canDelete: permission(action: "core:finding:delete")
        }
      }
    }
  }
`;

const schema = z.object({
  kind: z.enum(["MINOR_NONCONFORMITY", "MAJOR_NONCONFORMITY", "OBSERVATION", "EXCEPTION"]),
  description: z.string().optional(),
  source: z.string().optional(),
  identifiedOn: z.string().optional(),
  rootCause: z.string().optional(),
  correctiveAction: z.string().optional(),
  ownerId: z.string().nullable().optional(),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED", "MITIGATED", "FALSE_POSITIVE"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH"]),
  effectivenessCheck: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateFindingDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionIds?: string[];
}

export function CreateFindingDialog({
  children,
  organizationId,
  connectionIds,
}: CreateFindingDialogProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [createFinding] = useMutation<CreateFindingDialogMutation>(createFindingMutation);
  const statusOptions = ["OPEN", "IN_PROGRESS", "CLOSED", "MITIGATED", "FALSE_POSITIVE"] as const;

  const kindOptions = [
    { value: "MINOR_NONCONFORMITY", label: t("createFindingDialog.kinds.minorNonconformity") },
    { value: "MAJOR_NONCONFORMITY", label: t("createFindingDialog.kinds.majorNonconformity") },
    { value: "OBSERVATION", label: t("createFindingDialog.kinds.observation") },
    { value: "EXCEPTION", label: t("createFindingDialog.kinds.exception") },
  ];

  const priorityOptions = [
    { value: "LOW", label: t("createFindingDialog.priority.low") },
    { value: "MEDIUM", label: t("createFindingDialog.priority.medium") },
    { value: "HIGH", label: t("createFindingDialog.priority.high") },
  ];

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      kind: "MINOR_NONCONFORMITY" as const,
      description: "",
      source: "",
      identifiedOn: "",
      rootCause: "",
      correctiveAction: "",
      ownerId: null,
      dueDate: "",
      status: "OPEN" as const,
      priority: "MEDIUM" as const,
      effectivenessCheck: "",
    },
  });

  const onSubmit = (formData: FormData) => {
    createFinding({
      variables: {
        input: {
          organizationId,
          kind: formData.kind,
          description: formData.description || undefined,
          source: formData.source || undefined,
          identifiedOn: formatDatetime(formData.identifiedOn),
          rootCause: formData.rootCause || undefined,
          correctiveAction: formData.correctiveAction || undefined,
          ownerId: formData.ownerId || undefined,
          dueDate: formatDatetime(formData.dueDate),
          status: formData.status,
          priority: formData.priority,
          effectivenessCheck: formData.effectivenessCheck || undefined,
        },
        connections: connectionIds ?? [],
      },
      onCompleted() {
        toast({
          title: t("createFindingDialog.messages.successTitle"),
          description: t("createFindingDialog.messages.created"),
          variant: "success",
        });
        reset();
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: t("createFindingDialog.errors.title"),
          description: formatError(t("createFindingDialog.errors.create"), error),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[t("createFindingDialog.breadcrumbs.findings"), t("createFindingDialog.breadcrumbs.create")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Controller
            control={control}
            name="kind"
            render={({ field }) => (
              <Field label={t("createFindingDialog.fields.kind")} required>
                <Select
                  variant="editor"
                  placeholder={t("createFindingDialog.placeholders.selectKind")}
                  onValueChange={field.onChange}
                  value={field.value}
                  className="w-full"
                >
                  {kindOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      {option.label}
                    </Option>
                  ))}
                </Select>
                {formState.errors.kind && (
                  <p className="text-sm text-red-500 mt-1">{formState.errors.kind.message}</p>
                )}
              </Field>
            )}
          />

          <div className="space-y-2">
            <Label htmlFor="description">{t("createFindingDialog.fields.description")}</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder={t("createFindingDialog.placeholders.description")}
              rows={2}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field
              label={t("createFindingDialog.fields.source")}
              {...register("source")}
              placeholder={t("createFindingDialog.placeholders.source")}
              error={formState.errors.source?.message}
            />

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={t("createFindingDialog.fields.owner")}
              error={formState.errors.ownerId?.message}
              optional
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field label={t("createFindingDialog.fields.status")}>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={t("createFindingDialog.placeholders.selectStatus")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                  >
                    {statusOptions.map(status => (
                      <Option key={status} value={status}>
                        {t(`createFindingDialog.status.${status.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.status && (
                <p className="text-sm text-red-500 mt-1">{formState.errors.status.message}</p>
              )}
            </Field>

            <Controller
              control={control}
              name="priority"
              render={({ field }) => (
                <div>
                  <Label>
                    {t("createFindingDialog.fields.priority")}
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

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="identifiedOn">{t("createFindingDialog.fields.dateIdentified")}</Label>
              <Input
                id="identifiedOn"
                type="date"
                {...register("identifiedOn")}
              />
              {formState.errors.identifiedOn && (
                <p className="text-sm text-red-500">{formState.errors.identifiedOn.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="dueDate">{t("createFindingDialog.fields.dueDate")}</Label>
              <Input
                id="dueDate"
                type="date"
                {...register("dueDate")}
              />
              {formState.errors.dueDate && (
                <p className="text-sm text-red-500">{formState.errors.dueDate.message}</p>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="rootCause">{t("createFindingDialog.fields.rootCause")}</Label>
            <Textarea
              id="rootCause"
              {...register("rootCause")}
              placeholder={t("createFindingDialog.placeholders.rootCause")}
              rows={3}
            />
            {formState.errors.rootCause && (
              <p className="text-sm text-red-500">{formState.errors.rootCause.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="correctiveAction">{t("createFindingDialog.fields.correctiveAction")}</Label>
            <Textarea
              id="correctiveAction"
              {...register("correctiveAction")}
              placeholder={t("createFindingDialog.placeholders.correctiveAction")}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="effectivenessCheck">{t("createFindingDialog.fields.effectivenessCheck")}</Label>
            <Textarea
              id="effectivenessCheck"
              {...register("effectivenessCheck")}
              placeholder={t("createFindingDialog.placeholders.effectivenessCheck")}
              rows={2}
            />
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting}>
            {formState.isSubmitting
              ? t("createFindingDialog.actions.creating")
              : t("createFindingDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
