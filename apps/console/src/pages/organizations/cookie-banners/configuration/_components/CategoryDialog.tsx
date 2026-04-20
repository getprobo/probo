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
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Textarea,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategoryDialogCreateMutation } from "#/__generated__/core/CategoryDialogCreateMutation.graphql";
import type { CategoryDialogUpdateMutation } from "#/__generated__/core/CategoryDialogUpdateMutation.graphql";

const createMutation = graphql`
  mutation CategoryDialogCreateMutation(
    $input: CreateCookieCategoryInput!
    $connections: [ID!]!
  ) {
    createCookieCategory(input: $input) {
      cookieCategoryEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          description
          required
          rank
          cookies {
            name
            duration
            description
          }
          createdAt
          updatedAt
        }
      }
    }
  }
`;

const updateMutation = graphql`
  mutation CategoryDialogUpdateMutation($input: UpdateCookieCategoryInput!) {
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
    }
  }
`;

interface CookieEntry {
  name: string;
  duration: string;
  description: string;
}

interface CategoryDialogProps {
  cookieBannerId: string;
  connectionId: string;
  category?: {
    id: string;
    name: string;
    description: string;
    required: boolean;
    rank: number;
    cookies: ReadonlyArray<CookieEntry>;
  };
  onOpenChange: (open: boolean) => void;
}

export function CategoryDialog({
  cookieBannerId,
  connectionId,
  category,
  onOpenChange,
}: CategoryDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const isEditing = !!category;

  const [commitCreate, isCreating] = useMutation<CategoryDialogCreateMutation>(createMutation);
  const [commitUpdate, isUpdating] = useMutation<CategoryDialogUpdateMutation>(updateMutation);

  const [name, setName] = useState(category?.name ?? "");
  const [description, setDescription] = useState(category?.description ?? "");
  const [required, setRequired] = useState(category?.required ?? false);
  const [cookies, setCookies] = useState<CookieEntry[]>(
    category?.cookies ? [...category.cookies.map(c => ({ ...c }))] : [],
  );

  const addCookie = () => {
    setCookies([...cookies, { name: "", duration: "", description: "" }]);
  };

  const removeCookie = (index: number) => {
    setCookies(cookies.filter((_, i) => i !== index));
  };

  const updateCookie = (index: number, field: keyof CookieEntry, value: string) => {
    setCookies(cookies.map((c, i) => (i === index ? { ...c, [field]: value } : c)));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const cookieItems = cookies.filter(c => c.name.trim() !== "");

    if (isEditing) {
      commitUpdate({
        variables: {
          input: {
            cookieCategoryId: category.id,
            name,
            description,
            cookies: cookieItems,
          },
        },
        onCompleted() {
          toast({ title: __("Success"), description: __("Category updated"), variant: "success" });
          onOpenChange(false);
        },
        onError(error) {
          toast({ title: __("Error"), description: formatError(__("Failed to update category"), error as GraphQLError), variant: "error" });
        },
      });
    } else {
      commitCreate({
        variables: {
          input: {
            cookieBannerId,
            name,
            description,
            required,
            rank: 0,
            cookies: cookieItems.length > 0 ? cookieItems : null,
          },
          connections: [connectionId],
        },
        onCompleted() {
          toast({ title: __("Success"), description: __("Category created"), variant: "success" });
          onOpenChange(false);
        },
        onError(error) {
          toast({ title: __("Error"), description: formatError(__("Failed to create category"), error as GraphQLError), variant: "error" });
        },
      });
    }
  };

  return (
    <Dialog
      defaultOpen
      onClose={() => onOpenChange(false)}
      title={isEditing ? __("Edit Category") : __("Add Category")}
      className="max-w-lg"
    >
      <form onSubmit={handleSubmit}>
        <DialogContent padded className="space-y-4">
          <Field label={__("Name")}>
            <Input value={name} onChange={(e) => setName(e.target.value)} required />
          </Field>

          <Field label={__("Description")}>
            <Textarea value={description} onChange={(e) => setDescription(e.target.value)} required rows={2} />
          </Field>

          {!isEditing && (
            <div className="flex items-center gap-2">
              <Checkbox
                checked={required}
                onCheckedChange={(v) => setRequired(v === true)}
                id="required"
              />
              <Label htmlFor="required">{__("Required")}</Label>
            </div>
          )}

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>{__("Cookies")}</Label>
              <Button type="button" variant="secondary" onClick={addCookie}>
                {__("Add Cookie")}
              </Button>
            </div>
            {cookies.map((cookie, i) => (
              <div key={i} className="rounded border p-3 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-muted-foreground">{__("Cookie")} {i + 1}</span>
                  <button
                    type="button"
                    onClick={() => removeCookie(i)}
                    className="text-xs text-red-500 hover:text-red-700"
                  >
                    {__("Remove")}
                  </button>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <Input
                    placeholder={__("Cookie name")}
                    value={cookie.name}
                    onChange={(e) => updateCookie(i, "name", e.target.value)}
                  />
                  <Input
                    placeholder={__("Duration (e.g. 1 year)")}
                    value={cookie.duration}
                    onChange={(e) => updateCookie(i, "duration", e.target.value)}
                  />
                </div>
                <Input
                  placeholder={__("Description")}
                  value={cookie.description}
                  onChange={(e) => updateCookie(i, "description", e.target.value)}
                />
              </div>
            ))}
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={isCreating || isUpdating}>
            {isCreating || isUpdating ? __("Saving...") : isEditing ? __("Update") : __("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
