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

import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
  const frameworkControl = useFragment(controlFragment, props.control);
  const dialogRef = useDialogRef();
  const [mutate, isMutating] = useMutationWithToasts(
    props.control ? updateMutation : createMutation,
    {
      successMessage: __(
        `Control ${props.control ? "updated" : "created"} successfully.`,
      ),
      errorMessage: __(
        `Failed to ${props.control ? "update" : "create"} control`,
      ),
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
            __("Controls"),
            frameworkControl
              ? __("Edit Control")
              : __("New Control"),
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
            placeholder={__("Section title")}
            {...register("sectionTitle")}
          />
          <Input
            id="title"
            required
            variant="title"
            placeholder={__("Control name")}
            {...register("name")}
          />
          <Textarea
            id="content"
            variant="ghost"
            autogrow
            placeholder={__("Add description")}
            {...register("description")}
          />
          <div className="border border-border-low rounded-xl p-3 space-y-3 mt-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <Checkbox
                checked={bestPracticeValue}
                onChange={checked =>
                  setValue("bestPractice", checked)}
              />
              <span className="text-sm">{__("Best Practice")}</span>
            </label>
            <div className="flex items-center gap-2">
              <span className="text-sm">{__("Maturity level")}</span>
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
                placeholder={__("Justification for non-implementation")}
                {...register("notImplementedJustification")}
              />
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isMutating}>
            {props.control
              ? __("Update control")
              : __("Create control")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
