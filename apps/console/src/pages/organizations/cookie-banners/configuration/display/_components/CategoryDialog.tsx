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
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Textarea,
  useToast,
} from "@probo/ui";
import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategoryDialogCreateMutation } from "#/__generated__/core/CategoryDialogCreateMutation.graphql";

const createMutation = graphql`
  mutation CategoryDialogCreateMutation(
    $input: CreateCookieCategoryInput!
    $connections: [ID!]!
  ) {
    createCookieCategory(input: $input) {
      cookieCategoryEdge @appendEdge(connections: $connections) {
        node {
          id
          rank
          name
          kind
          ...CategorySectionFragment
        }
      }
      cookieBanner {
        id
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface CategoryFormValues {
  name: string;
  slug: string;
  description: string;
}

interface CategoryDialogProps {
  cookieBannerId: string;
  connectionId: string;
  nextRank: number;
  onOpenChange: (open: boolean) => void;
}

export function CategoryDialog({
  cookieBannerId,
  connectionId,
  nextRank,
  onOpenChange,
}: CategoryDialogProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();

  const [create, isCreating] = useMutation<CategoryDialogCreateMutation>(createMutation);

  const { register, handleSubmit, setValue, control, formState } = useForm<CategoryFormValues>({
    defaultValues: { name: "", slug: "", description: "" },
  });

  const nameValue = useWatch({ control, name: "name" });

  useEffect(() => {
    if (!formState.dirtyFields.slug) {
      setValue(
        "slug",
        nameValue
          .toLowerCase()
          .replace(/[^a-z0-9]+/g, "-")
          .replace(/^-|-$/g, ""),
        { shouldDirty: false },
      );
    }
  }, [nameValue, formState.dirtyFields.slug, setValue]);

  const onSubmit = (data: CategoryFormValues) => {
    create({
      variables: {
        input: {
          cookieBannerId,
          name: data.name,
          slug: data.slug,
          description: data.description,
          rank: nextRank,
        },
        connections: [connectionId],
      },
      onCompleted() {
        toast({ title: t("categoryDialog.messages.successTitle"), description: t("categoryDialog.messages.created"), variant: "success" });
        onOpenChange(false);
      },
      onError(error) {
        toast({ title: t("categoryDialog.errors.title"), description: formatError(t("categoryDialog.errors.create"), error), variant: "error" });
      },
    });
  };

  return (
    <Dialog
      defaultOpen
      onClose={() => onOpenChange(false)}
      title={t("categoryDialog.title")}
      className="max-w-lg"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field label={t("categoryDialog.fields.name")}>
            <Input {...register("name")} required />
          </Field>

          <Field label={t("categoryDialog.fields.slug")} help={t("categoryDialog.fields.slugHelp")}>
            <Input
              {...register("slug", {
                pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
              })}
              required
            />
          </Field>

          <Field label={t("categoryDialog.fields.description")}>
            <Textarea {...register("description")} required rows={2} />
          </Field>

        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {isCreating ? t("categoryDialog.actions.saving") : t("categoryDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
