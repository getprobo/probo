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
  Badge,
  Button,
  Card,
  IconPencil,
  IconPlusSmall,
  IconTrashCan,
  Input,
  Tbody,
  Td,
  Textarea,
  Th,
  Thead,
  Tr,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategorySectionFragment$key } from "#/__generated__/core/CategorySectionFragment.graphql";
import type { CategorySectionUpdateMutation } from "#/__generated__/core/CategorySectionUpdateMutation.graphql";

export const categorySectionFragment = graphql`
  fragment CategorySectionFragment on CookieCategory {
    id
    name
    description
    required
    cookies {
      name
      duration
      description
    }
  }
`;

const updateCategoryMutation = graphql`
  mutation CategorySectionUpdateMutation(
    $input: UpdateCookieCategoryInput!
  ) {
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

interface CategorySectionProps {
  categoryKey: CategorySectionFragment$key;
}

export function CategorySection({ categoryKey }: CategorySectionProps) {
  const category = useFragment(categorySectionFragment, categoryKey);
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [commitUpdate, isUpdating]
    = useMutation<CategorySectionUpdateMutation>(updateCategoryMutation);

  const [isEditingCategory, setIsEditingCategory] = useState(false);
  const [editName, setEditName] = useState(category.name);
  const [editDescription, setEditDescription] = useState(category.description);

  const [editingCookieIndex, setEditingCookieIndex] = useState<number | null>(
    null,
  );
  const [isAddingCookie, setIsAddingCookie] = useState(false);
  const [cookieForm, setCookieForm] = useState<CookieEntry>({
    name: "",
    duration: "",
    description: "",
  });

  const doUpdate = (
    input: Record<string, unknown>,
    onSuccess?: () => void,
  ) => {
    commitUpdate({
      variables: {
        input: {
          cookieCategoryId: category.id,
          ...input,
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Category updated"),
          variant: "success",
        });
        onSuccess?.();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update category"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleSaveCategory = () => {
    doUpdate({ name: editName, description: editDescription }, () => {
      setIsEditingCategory(false);
    });
  };

  const handleCancelCategoryEdit = () => {
    setEditName(category.name);
    setEditDescription(category.description);
    setIsEditingCategory(false);
  };

  const handleStartEditCookie = (index: number) => {
    const c = category.cookies[index];
    setCookieForm({
      name: c.name,
      duration: c.duration,
      description: c.description,
    });
    setEditingCookieIndex(index);
    setIsAddingCookie(false);
  };

  const handleSaveEditCookie = () => {
    if (editingCookieIndex === null) return;
    const newCookies = category.cookies.map((c, i) =>
      i === editingCookieIndex
        ? { ...cookieForm }
        : { name: c.name, duration: c.duration, description: c.description },
    );
    doUpdate({ cookies: newCookies }, () => {
      setEditingCookieIndex(null);
      setCookieForm({ name: "", duration: "", description: "" });
    });
  };

  const handleCancelEditCookie = () => {
    setEditingCookieIndex(null);
    setCookieForm({ name: "", duration: "", description: "" });
  };

  const handleDeleteCookie = (index: number) => {
    const newCookies = category.cookies
      .filter((_, i) => i !== index)
      .map(c => ({
        name: c.name,
        duration: c.duration,
        description: c.description,
      }));
    doUpdate({ cookies: newCookies });
  };

  const handleStartAddCookie = () => {
    setCookieForm({ name: "", duration: "", description: "" });
    setIsAddingCookie(true);
    setEditingCookieIndex(null);
  };

  const handleSaveNewCookie = () => {
    if (!cookieForm.name.trim()) return;
    const newCookies = [
      ...category.cookies.map(c => ({
        name: c.name,
        duration: c.duration,
        description: c.description,
      })),
      { ...cookieForm },
    ];
    doUpdate({ cookies: newCookies }, () => {
      setIsAddingCookie(false);
      setCookieForm({ name: "", duration: "", description: "" });
    });
  };

  const handleCancelAddCookie = () => {
    setIsAddingCookie(false);
    setCookieForm({ name: "", duration: "", description: "" });
  };

  return (
    <Card className="border overflow-hidden">
      <div className="p-4">
        {isEditingCategory
          ? (
              <div className="space-y-3">
                <Input
                  value={editName}
                  onChange={e => setEditName(e.target.value)}
                  placeholder={__("Category name")}
                />
                <Textarea
                  value={editDescription}
                  onChange={e => setEditDescription(e.target.value)}
                  placeholder={__("Category description")}
                  rows={2}
                />
                <div className="flex items-center gap-2">
                  <Button
                    onClick={handleSaveCategory}
                    disabled={isUpdating}
                  >
                    {isUpdating ? __("Saving...") : __("Save")}
                  </Button>
                  <Button
                    variant="secondary"
                    onClick={handleCancelCategoryEdit}
                  >
                    {__("Cancel")}
                  </Button>
                </div>
              </div>
            )
          : (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{category.name}</span>
                  {category.required && (
                    <Badge variant="neutral">{__("Required")}</Badge>
                  )}
                </div>
                <Button
                  variant="secondary"
                  onClick={() => setIsEditingCategory(true)}
                >
                  <IconPencil size={14} />
                  {__("Edit")}
                </Button>
              </div>
            )}
        {!isEditingCategory && (
          <p className="mt-1 text-sm text-muted-foreground">
            {category.description}
          </p>
        )}
      </div>

      <table className="w-full text-left">
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Duration")}</Th>
            <Th>{__("Description")}</Th>
            <Th className="w-20" />
          </Tr>
        </Thead>
        <Tbody>
          {category.cookies.map((cookie, index) =>
            editingCookieIndex === index
              ? (
                  <Tr key={index}>
                    <Td>
                      <Input
                        value={cookieForm.name}
                        onChange={e =>
                          setCookieForm({ ...cookieForm, name: e.target.value })}
                        placeholder={__("Cookie name")}
                      />
                    </Td>
                    <Td>
                      <Input
                        value={cookieForm.duration}
                        onChange={e =>
                          setCookieForm({
                            ...cookieForm,
                            duration: e.target.value,
                          })}
                        placeholder={__("e.g. 1 year")}
                      />
                    </Td>
                    <Td>
                      <Input
                        value={cookieForm.description}
                        onChange={e =>
                          setCookieForm({
                            ...cookieForm,
                            description: e.target.value,
                          })}
                        placeholder={__("Description")}
                      />
                    </Td>
                    <Td>
                      <div className="flex items-center gap-1">
                        <Button
                          onClick={handleSaveEditCookie}
                          disabled={isUpdating}
                        >
                          {__("Save")}
                        </Button>
                        <Button
                          variant="secondary"
                          onClick={handleCancelEditCookie}
                        >
                          {__("Cancel")}
                        </Button>
                      </div>
                    </Td>
                  </Tr>
                )
              : (
                  <Tr key={index}>
                    <Td>
                      <code className="text-sm font-mono">{cookie.name}</code>
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {cookie.duration}
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {cookie.description}
                    </Td>
                    <Td>
                      <div className="flex items-center gap-1">
                        <button
                          type="button"
                          onClick={() => handleStartEditCookie(index)}
                          className="p-1 rounded cursor-pointer text-muted-foreground hover:text-foreground"
                        >
                          <IconPencil size={14} />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDeleteCookie(index)}
                          className="p-1 rounded cursor-pointer text-muted-foreground hover:text-red-500"
                        >
                          <IconTrashCan size={14} />
                        </button>
                      </div>
                    </Td>
                  </Tr>
                ),
          )}
          {isAddingCookie && (
            <Tr>
              <Td className="pr-3">
                <Input
                  value={cookieForm.name}
                  onChange={e =>
                    setCookieForm({ ...cookieForm, name: e.target.value })}
                  placeholder={__("Cookie name")}
                />
              </Td>
              <Td className="pr-3">
                <Input
                  value={cookieForm.duration}
                  onChange={e =>
                    setCookieForm({ ...cookieForm, duration: e.target.value })}
                  placeholder={__("e.g. 1 year")}
                />
              </Td>
              <Td className="pr-3">
                <Input
                  value={cookieForm.description}
                  onChange={e =>
                    setCookieForm({
                      ...cookieForm,
                      description: e.target.value,
                    })}
                  placeholder={__("Description")}
                />
              </Td>
              <Td>
                <div className="flex items-center gap-2">
                  <Button
                    onClick={handleSaveNewCookie}
                    disabled={isUpdating}
                  >
                    {__("Save")}
                  </Button>
                  <Button
                    variant="secondary"
                    onClick={handleCancelAddCookie}
                  >
                    {__("Cancel")}
                  </Button>
                </div>
              </Td>
            </Tr>
          )}
        </Tbody>
      </table>

      {!isAddingCookie && (
        <div className="p-3 border-t border-border-low">
          <Button
            variant="secondary"
            onClick={handleStartAddCookie}
          >
            <IconPlusSmall size={14} />
            {__("Add Cookie")}
          </Button>
        </div>
      )}
    </Card>
  );
}
