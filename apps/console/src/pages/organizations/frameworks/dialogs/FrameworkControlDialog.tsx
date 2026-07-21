// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  Breadcrumb,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Select,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { FrameworkControlDialogFragment$key } from "#/__generated__/core/FrameworkControlDialogFragment.graphql";
import { ControlMaturityLevelOptions } from "#/components/form/ControlMaturityLevelOptions";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

type Props = {
  children: ReactNode;
  control?: FrameworkControlDialogFragment$key;
  frameworkId: string;
  connectionId?: string;
};

const controlFragment = graphql`
    fragment FrameworkControlDialogFragment on Control {
        id
        name
        description
        sectionTitle
        bestPractice
        notImplementedJustification
        maturityLevel
    }
`;

const createMutation = graphql`
    mutation FrameworkControlDialogCreateMutation(
        $input: CreateControlInput!
        $connections: [ID!]!
    ) {
        createControl(input: $input) {
            controlEdge @prependEdge(connections: $connections) {
                node {
                    ...FrameworkControlDialogFragment
                }
            }
        }
    }
`;

const updateMutation = graphql`
    mutation FrameworkControlDialogUpdateMutation($input: UpdateControlInput!) {
        updateControl(input: $input) {
            control {
                ...FrameworkControlDialogFragment
            }
        }
    }
`;

const schema = z.object({
  name: z.string(),
  description: z.string().optional().nullable(),
  sectionTitle: z.string(),
  bestPractice: z.boolean(),
  maturityLevel: z.enum([
    "NONE",
    "INITIAL",
    "MANAGED",
    "DEFINED",
    "QUANTITATIVELY_MANAGED",
    "OPTIMIZING",
  ]),
  notImplementedJustification: z.string().optional().nullable(),
});

export function FrameworkControlDialog(props: Props) {
  const { t } = useTranslation();
  const frameworkControl = useFragment(controlFragment, props.control);
  const dialogRef = useDialogRef();
  const [mutate, isMutating] = useMutationWithToasts(
    props.control ? updateMutation : createMutation,
    {
      successMessage: props.control
        ? t("frameworkControlDialog.messages.updated")
        : t("frameworkControlDialog.messages.created"),
      errorMessage: props.control
        ? t("frameworkControlDialog.errors.update")
        : t("frameworkControlDialog.errors.create"),
    },
  );

  const defaultValues = useMemo<z.infer<typeof schema>>(
    () => ({
      name: frameworkControl?.name ?? "",
      description: frameworkControl?.description ?? "",
      sectionTitle: frameworkControl?.sectionTitle ?? "",
      bestPractice: frameworkControl?.bestPractice ?? true,
      maturityLevel: frameworkControl?.maturityLevel ?? "INITIAL",
      notImplementedJustification: frameworkControl?.notImplementedJustification ?? "",
    }),
    [frameworkControl],
  );

  const { handleSubmit, register, reset, watch, setValue }
    = useFormWithSchema(schema, {
      defaultValues,
    });

  useEffect(() => {
    reset(defaultValues);
  }, [defaultValues, reset]);

  const bestPracticeValue = watch("bestPractice");
  const maturityLevelValue = watch("maturityLevel");

  const onSubmit = async (data: z.infer<typeof schema>) => {
    if (frameworkControl) {
      await mutate({
        variables: {
          input: {
            id: frameworkControl.id,
            name: data.name,
            description: data.description || null,
            sectionTitle: data.sectionTitle,
            bestPractice: data.bestPractice,
            maturityLevel: data.maturityLevel,
            notImplementedJustification: data.maturityLevel === "NONE" ? (data.notImplementedJustification || null) : null,
          },
        },
      });
    } else {
      await mutate({
        variables: {
          input: {
            frameworkId: props.frameworkId,
            name: data.name,
            description: data.description || null,
            sectionTitle: data.sectionTitle,
            bestPractice: data.bestPractice ?? true,
            maturityLevel: data.maturityLevel,
            notImplementedJustification: data.maturityLevel === "NONE" ? (data.notImplementedJustification || null) : null,
          },
          connections: [props.connectionId!],
        },
      });
      reset();
    }
    dialogRef.current?.close();
  };

  return (
    <Dialog
      trigger={props.children}
      ref={dialogRef}
      title={(
        <Breadcrumb
          items={[
            t("frameworkControlDialog.breadcrumb.controls"),
            frameworkControl
              ? t("frameworkControlDialog.breadcrumb.editControl")
              : t("frameworkControlDialog.breadcrumb.newControl"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-2">
          <Input
            id="sectionTitle"
            required
            variant="ghost"
            placeholder={t("frameworkControlDialog.fields.sectionTitle")}
            {...register("sectionTitle")}
          />
          <Input
            id="title"
            required
            variant="title"
            placeholder={t("frameworkControlDialog.fields.name")}
            {...register("name")}
          />
          <Textarea
            id="content"
            variant="ghost"
            autogrow
            placeholder={t("frameworkControlDialog.fields.description")}
            {...register("description")}
          />
          <div className="border border-border-low rounded-xl p-3 space-y-3 mt-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <Checkbox
                checked={bestPracticeValue}
                onChange={checked =>
                  setValue("bestPractice", checked)}
              />
              <span className="text-sm">{t("frameworkControlDialog.fields.bestPractice")}</span>
            </label>
            <div className="flex items-center gap-2">
              <span className="text-sm">{t("frameworkControlDialog.fields.maturityLevel")}</span>
              <Select
                id="maturityLevel"
                value={maturityLevelValue}
                onValueChange={value =>
                  setValue("maturityLevel", value as typeof maturityLevelValue)}
              >
                <ControlMaturityLevelOptions />
              </Select>
            </div>
            {maturityLevelValue === "NONE" && (
              <Textarea
                id="notImplementedJustification"
                variant="ghost"
                autogrow
                placeholder={t("frameworkControlDialog.fields.notImplementedJustification")}
                {...register("notImplementedJustification")}
              />
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isMutating}>
            {props.control
              ? t("frameworkControlDialog.actions.update")
              : t("frameworkControlDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
