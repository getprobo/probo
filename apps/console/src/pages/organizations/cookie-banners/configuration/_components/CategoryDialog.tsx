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
  Textarea,
  useToast,
} from "@probo/ui";
import { useState } from "react";
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
          name
          slug
          description
          kind
          rank
          createdAt
          updatedAt
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
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [create, isCreating] = useMutation<CategoryDialogCreateMutation>(createMutation);

  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugTouched, setSlugTouched] = useState(false);
  const [description, setDescription] = useState("");

  const handleNameChange = (value: string) => {
    setName(value);
    if (!slugTouched) {
      setSlug(
        value
          .toLowerCase()
          .replace(/[^a-z0-9]+/g, "-")
          .replace(/^-|-$/g, ""),
      );
    }
  };

  const handleSlugChange = (value: string) => {
    setSlugTouched(true);
    setSlug(value);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    create({
      variables: {
        input: {
          cookieBannerId,
          name,
          slug,
          description,
          rank: nextRank,
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
  };

  return (
    <Dialog
      defaultOpen
      onClose={() => onOpenChange(false)}
      title={__("Add Category")}
      className="max-w-lg"
    >
      <form onSubmit={handleSubmit}>
        <DialogContent padded className="space-y-4">
          <Field label={__("Name")}>
            <Input value={name} onChange={e => handleNameChange(e.target.value)} required />
          </Field>

          <Field label={__("Slug")} help={__("Used as the data-cookie-consent attribute value")}>
            <Input value={slug} onChange={e => handleSlugChange(e.target.value)} required pattern="[a-z0-9]+(-[a-z0-9]+)*" />
          </Field>

          <Field label={__("Description")}>
            <Textarea value={description} onChange={e => setDescription(e.target.value)} required rows={2} />
          </Field>

        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {isCreating ? __("Saving...") : __("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
