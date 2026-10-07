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

import { Form } from "@base-ui/react/form";
import { toFieldErrors } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { CategoryCreateDialogMutation } from "#/__generated__/core/CategoryCreateDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { categoryCreateDialog } from "../../../variants";
import {
  CATEGORY_DESCRIPTION_MAX_LENGTH,
  CATEGORY_NAME_MAX_LENGTH,
  CATEGORY_SLUG_MAX_LENGTH,
  CATEGORY_SLUG_PATTERN,
  slugFromName,
} from "../_lib/categorySlug";

const createMutation = graphql`
  mutation CategoryCreateDialogMutation(
    $input: CreateCookieCategoryInput!
    $connections: [ID!]!
  ) {
    createCookieCategory(input: $input) {
      cookieCategoryEdge @appendEdge(connections: $connections) {
        node {
          id
          rank
          ...CategoryListItem_cookieCategory
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

interface CategoryCreateDialogProps {
  cookieBannerId: string;
  connectionId: string;
  nextRank: number;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CategoryCreateDialog({
  cookieBannerId,
  connectionId,
  nextRank,
  open,
  onOpenChange,
}: CategoryCreateDialogProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { form, fields, field } = categoryCreateDialog();
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugDirty, setSlugDirty] = useState(false);
  const [description, setDescription] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [createCategory, isCreating] = useMutation<CategoryCreateDialogMutation>(
    createMutation,
    {
      successMessage: t("categoryCreateDialog.messages.created"),
      errorToast: t("categoryCreateDialog.errors.create"),
    },
  );

  function reset() {
    setName("");
    setSlug("");
    setSlugDirty(false);
    setDescription("");
    setErrors({});
  }

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      reset();
    }
    onOpenChange(nextOpen);
  }

  function handleSubmit() {
    void createCategory({
      variables: {
        input: {
          cookieBannerId,
          name: name.trim(),
          slug: slug.trim(),
          description: description.trim(),
          rank: nextRank,
        },
        connections: [connectionId],
      },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
        }
      },
    }).then(
      () => {
        handleOpenChange(false);
      },
      () => {
        // Field errors are mapped in onCompleted; other failures toast.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogPopup>
        <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t("categoryCreateDialog.title")}</DialogTitle>
          </DialogHeader>
          <DialogBody className={fields()}>
            <Field
              required
              label={t("categoryCreateDialog.fields.name")}
              error={errors.name}
            >
              <TextField
                name="name"
                required
                maxLength={CATEGORY_NAME_MAX_LENGTH}
                value={name}
                disabled={isCreating}
                onValueChange={(value) => {
                  setName(value);
                  if (!slugDirty) {
                    setSlug(slugFromName(value));
                  }
                  setErrors({});
                }}
              />
            </Field>
            <div className={field()}>
              <Field
                required
                label={t("categoryCreateDialog.fields.slug")}
                error={errors.slug}
              >
                <TextField
                  name="slug"
                  required
                  maxLength={CATEGORY_SLUG_MAX_LENGTH}
                  pattern={CATEGORY_SLUG_PATTERN}
                  value={slug}
                  disabled={isCreating}
                  onValueChange={(value) => {
                    setSlug(value);
                    setSlugDirty(true);
                    setErrors({});
                  }}
                />
              </Field>
              <Text size={1} color="faint">
                {t("categoryCreateDialog.fields.slugHelp")}
              </Text>
            </div>
            <Field
              required
              label={t("categoryCreateDialog.fields.description")}
              error={errors.description}
            >
              <Textarea
                name="description"
                required
                maxLength={CATEGORY_DESCRIPTION_MAX_LENGTH}
                rows={3}
                value={description}
                disabled={isCreating}
                onChange={(event) => {
                  setDescription(event.target.value);
                  setErrors({});
                }}
              />
            </Field>
          </DialogBody>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral">
                  {t("categoryCreateDialog.actions.cancel")}
                </Button>
              )}
            />
            <Button type="submit" variant="solid" color="neutral" highContrast loading={isCreating}>
              {t("categoryCreateDialog.actions.create")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}
