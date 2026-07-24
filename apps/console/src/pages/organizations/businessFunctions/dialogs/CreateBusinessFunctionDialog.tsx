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
import { graphql, useMutation } from "react-relay";
import { z } from "zod";

import type { CreateBusinessFunctionDialogMutation } from "#/__generated__/core/CreateBusinessFunctionDialogMutation.graphql";
import { AssetsMultiSelectField } from "#/components/form/AssetsMultiSelectField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import {
  businessFunctionClassificationOptions,
  durationMinutesHelperText,
} from "../_lib/businessFunctionHelpers";

const createBusinessFunctionMutation = graphql`
  mutation CreateBusinessFunctionDialogMutation(
    $input: CreateBusinessFunctionInput!
    $connections: [ID!]!
  ) {
    createBusinessFunction(input: $input) {
      businessFunctionEdge @prependEdge(connections: $connections) {
        node {
          id
          referenceId
          name
          classification
          mtdMinutes
          rtoMinutes
          rpoMinutes
          owner {
            id
            fullName
          }
          createdAt
          canUpdate: permission(action: "core:business-function:update")
          canDelete: permission(action: "core:business-function:delete")
        }
      }
    }
  }
`;

const schema = z.object({
  referenceId: z.string().min(1),
  name: z.string().min(1),
  classification: z.enum(["CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"]),
  mtdMinutes: z.coerce.number().int().min(0),
  rtoMinutes: z.coerce.number().int().min(0),
  rpoMinutes: z.coerce.number().int().min(0),
  impactTolerance: z.string().optional(),
  notes: z.string().optional(),
  ownerId: z.string().nullable().optional(),
  assetIds: z.array(z.string()).optional(),
  thirdPartyIds: z.array(z.string()).optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateBusinessFunctionDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionIds?: string[];
}

export function CreateBusinessFunctionDialog({
  children,
  organizationId,
  connectionIds,
}: CreateBusinessFunctionDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [createBusinessFunction] = useMutation<CreateBusinessFunctionDialogMutation>(
    createBusinessFunctionMutation,
  );

  const classificationOptions = businessFunctionClassificationOptions(__);
  const durationHelper = durationMinutesHelperText(__);

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: "",
      name: "",
      classification: "STANDARD" as const,
      mtdMinutes: 0,
      rtoMinutes: 0,
      rpoMinutes: 0,
      impactTolerance: "",
      notes: "",
      ownerId: null,
      assetIds: [],
      thirdPartyIds: [],
    },
  });

  const onSubmit = (formData: FormData) => {
    createBusinessFunction({
      variables: {
        input: {
          organizationId,
          referenceId: formData.referenceId,
          name: formData.name,
          classification: formData.classification,
          mtdMinutes: formData.mtdMinutes,
          rtoMinutes: formData.rtoMinutes,
          rpoMinutes: formData.rpoMinutes,
          impactTolerance: formData.impactTolerance || undefined,
          notes: formData.notes || undefined,
          ownerId: formData.ownerId || undefined,
          assetIds: formData.assetIds,
          thirdPartyIds: formData.thirdPartyIds,
        },
        connections: connectionIds ?? [],
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Business function created successfully"),
          variant: "success",
        });
        reset();
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to create business function"), error),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Business functions"), __("Create")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Field
              label={__("Reference ID")}
              error={formState.errors.referenceId?.message}
              required
            >
              <Input
                {...register("referenceId")}
                placeholder={__("e.g. F-10")}
              />
            </Field>

            <Field
              label={__("Name")}
              error={formState.errors.name?.message}
              required
            >
              <Input
                {...register("name")}
                placeholder={__("Enter name")}
              />
            </Field>
          </div>

          <Controller
            control={control}
            name="classification"
            render={({ field }) => (
              <Field label={__("Classification")} required>
                <Select
                  variant="editor"
                  placeholder={__("Select classification")}
                  onValueChange={field.onChange}
                  value={field.value}
                  className="w-full"
                >
                  {classificationOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      {option.label}
                    </Option>
                  ))}
                </Select>
              </Field>
            )}
          />

          <div className="grid grid-cols-3 gap-4">
            <Field
              label={__("MTD (minutes)")}
              error={formState.errors.mtdMinutes?.message}
              help={durationHelper}
              required
            >
              <Input {...register("mtdMinutes")} type="number" min={0} />
            </Field>

            <Field
              label={__("RTO (minutes)")}
              error={formState.errors.rtoMinutes?.message}
              help={durationHelper}
              required
            >
              <Input {...register("rtoMinutes")} type="number" min={0} />
            </Field>

            <Field
              label={__("RPO (minutes)")}
              error={formState.errors.rpoMinutes?.message}
              help={durationHelper}
              required
            >
              <Input {...register("rpoMinutes")} type="number" min={0} />
            </Field>
          </div>

          <div className="space-y-2">
            <Label htmlFor="impactTolerance">{__("Impact tolerance")}</Label>
            <Textarea
              id="impactTolerance"
              {...register("impactTolerance")}
              placeholder={__("Describe the impact tolerance")}
              rows={2}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">{__("Notes")}</Label>
            <Textarea
              id="notes"
              {...register("notes")}
              placeholder={__("Dependencies and other notes")}
              rows={2}
            />
          </div>

          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={__("Owner")}
            error={formState.errors.ownerId?.message}
            optional
          />

          <AssetsMultiSelectField
            organizationId={organizationId}
            control={control}
            name="assetIds"
            label={__("Assets")}
          />

          <ThirdPartiesMultiSelectField
            organizationId={organizationId}
            control={control}
            name="thirdPartyIds"
            label={__("Third parties")}
          />
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting}>
            {formState.isSubmitting
              ? __("Creating...")
              : __("Create business function")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
