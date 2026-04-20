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
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieDialogUpdateMutation } from "#/__generated__/core/CookieDialogUpdateMutation.graphql";

const updateCategoryMutation = graphql`
  mutation CookieDialogUpdateMutation($input: UpdateCookieCategoryInput!) {
    updateCookieCategory(input: $input) {
      cookieCategory {
        id
        name
        description
        rank
        cookies {
          name
          duration
          description
        }
        updatedAt
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

interface CookieEntry {
  name: string;
  duration: string;
  description: string;
}

interface Category {
  id: string;
  name: string;
  cookies: ReadonlyArray<CookieEntry>;
}

interface CookieDialogProps {
  categories: ReadonlyArray<Category>;
  onOpenChange: (open: boolean) => void;
}

export function CookieDialog({ categories, onOpenChange }: CookieDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [commitUpdate, isUpdating] = useMutation<CookieDialogUpdateMutation>(updateCategoryMutation);

  const [categoryId, setCategoryId] = useState(categories[0]?.id ?? "");
  const [name, setName] = useState("");
  const [duration, setDuration] = useState("");
  const [description, setDescription] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const category = categories.find(c => c.id === categoryId);
    if (!category) return;

    const existingCookies = category.cookies.map(c => ({
      name: c.name,
      duration: c.duration,
      description: c.description,
    }));

    commitUpdate({
      variables: {
        input: {
          cookieCategoryId: categoryId,
          cookies: [
            ...existingCookies,
            {
              name: name.trim(),
              duration: duration.trim(),
              description: description.trim(),
            },
          ],
        },
      },
      onCompleted() {
        toast({ title: __("Success"), description: __("Cookie added"), variant: "success" });
        dialogRef.current?.close();
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to add cookie"), error as GraphQLError), variant: "error" });
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      defaultOpen
      onClose={() => onOpenChange(false)}
      title={__("Add Cookie")}
      className="max-w-lg"
    >
      <form onSubmit={handleSubmit}>
        <DialogContent padded className="space-y-4">
          <div className="space-y-2">
            <Label>{__("Category")}</Label>
            <Select value={categoryId} onValueChange={id => setCategoryId(id)}>
              {categories.map(cat => (
                <Option key={cat.id} value={cat.id}>{cat.name}</Option>
              ))}
            </Select>
          </div>

          <Field label={__("Cookie name")}>
            <Input value={name} onChange={e => setName(e.target.value)} required />
          </Field>

          <Field label={__("Duration")}>
            <Input
              value={duration}
              onChange={e => setDuration(e.target.value)}
              required
              placeholder={__("e.g. 1 year")}
            />
          </Field>

          <Field label={__("Description")}>
            <Input value={description} onChange={e => setDescription(e.target.value)} required />
          </Field>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={isUpdating}>
            {isUpdating ? __("Adding...") : __("Add Cookie")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
