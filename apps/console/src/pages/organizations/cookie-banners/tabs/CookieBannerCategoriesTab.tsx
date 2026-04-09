import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Field,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { graphql, useFragment, useMutation } from "react-relay";
import { useOutletContext } from "react-router";
import { z } from "zod";

import type { CookieBannerCategoriesTabCreateMutation } from "#/__generated__/core/CookieBannerCategoriesTabCreateMutation.graphql";
import type { CookieBannerCategoriesTabDeleteMutation } from "#/__generated__/core/CookieBannerCategoriesTabDeleteMutation.graphql";
import type { CookieBannerCategoriesTabFragment$key } from "#/__generated__/core/CookieBannerCategoriesTabFragment.graphql";
import type { CookieBannerCategoriesTabUpdateMutation } from "#/__generated__/core/CookieBannerCategoriesTabUpdateMutation.graphql";
import type { CookieBannerDetailPageQuery$data } from "#/__generated__/core/CookieBannerDetailPageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const fragment = graphql`
  fragment CookieBannerCategoriesTabFragment on CookieBanner {
    id
    canUpdate: permission(action: "core:cookie-banner:update")
    categories(first: 50, orderBy: { field: RANK, direction: ASC }) {
      edges {
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
        }
      }
    }
  }
`;

const createCategoryMutation = graphql`
  mutation CookieBannerCategoriesTabCreateMutation(
    $input: CreateCookieCategoryInput!
  ) {
    createCookieCategory(input: $input) {
      cookieCategory {
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
      }
    }
  }
`;

const updateCategoryMutation = graphql`
  mutation CookieBannerCategoriesTabUpdateMutation(
    $input: UpdateCookieCategoryInput!
  ) {
    updateCookieCategory(input: $input) {
      cookieCategory {
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
      }
    }
  }
`;

const deleteCategoryMutation = graphql`
  mutation CookieBannerCategoriesTabDeleteMutation(
    $input: DeleteCookieCategoryInput!
  ) {
    deleteCookieCategory(input: $input) {
      deletedCookieCategoryId
    }
  }
`;

type CategoryNode = {
  id: string;
  name: string;
  description: string;
  required: boolean;
  rank: number;
  cookies: readonly {
    readonly name: string;
    readonly duration: string;
    readonly description: string;
  }[];
};

export default function CookieBannerCategoriesTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerDetailPageQuery$data["node"];
  }>();

  const { __ } = useTranslate();
  const data = useFragment<CookieBannerCategoriesTabFragment$key>(
    fragment,
    banner,
  );

  const categories = data.categories?.edges.map((edge) => edge.node) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-base font-medium">{__("Cookie Categories")}</h2>
        {data.canUpdate && (
          <CreateCategoryDialog cookieBannerId={data.id}>
            <Button icon={IconPlusLarge} variant="secondary">
              {__("Add category")}
            </Button>
          </CreateCategoryDialog>
        )}
      </div>

      <div className="space-y-4">
        {categories.map((category) => (
          <CategoryCard
            key={category.id}
            category={category}
            canUpdate={data.canUpdate}
          />
        ))}
      </div>
    </div>
  );
}

function CategoryCard({
  category,
  canUpdate,
}: {
  category: CategoryNode;
  canUpdate: boolean;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const [deleteCategory] = useMutation<CookieBannerCategoriesTabDeleteMutation>(deleteCategoryMutation);

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCategory({
            variables: {
              input: { id: category.id },
            },
            onCompleted() {
              toast({
                title: __("Success"),
                description: __("Category deleted successfully."),
                variant: "success",
              });
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(__("Failed to delete category"), error as GraphQLError),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete category "%s". This action cannot be undone.',
          ),
          category.name,
        ),
      },
    );
  };

  return (
    <Card padded>
      <div className="flex justify-between items-start">
        <div className="space-y-1 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-medium">{category.name}</span>
            {category.required && (
              <Badge variant="info">{__("Required")}</Badge>
            )}
          </div>
          <p className="text-sm text-txt-secondary">{category.description}</p>
          {category.cookies.length > 0 && (
            <div className="mt-3">
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr className="text-left text-txt-secondary">
                    <th className="pb-2 pr-4 font-medium">{__("Cookie")}</th>
                    <th className="pb-2 pr-4 font-medium">{__("Duration")}</th>
                    <th className="pb-2 font-medium">{__("Description")}</th>
                  </tr>
                </thead>
                <tbody>
                  {category.cookies.map((cookie) => (
                    <tr key={cookie.name} className="border-t border-border">
                      <td className="py-2 pr-4 font-mono text-xs">
                        {cookie.name}
                      </td>
                      <td className="py-2 pr-4 text-txt-secondary">
                        {cookie.duration}
                      </td>
                      <td className="py-2 text-txt-secondary">
                        {cookie.description}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
        {canUpdate && (
          <ActionDropdown>
            <EditCategoryDialog category={category}>
              <DropdownItem icon={IconPencil}>{__("Edit")}</DropdownItem>
            </EditCategoryDialog>
            {!category.required && (
              <DropdownItem
                onClick={handleDelete}
                variant="danger"
                icon={IconTrashCan}
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        )}
      </div>
    </Card>
  );
}

const categorySchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  rank: z.number().min(0),
});

function CreateCategoryDialog({
  children,
  cookieBannerId,
}: {
  children: React.ReactNode;
  cookieBannerId: string;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [createCategory] = useMutation<CookieBannerCategoriesTabCreateMutation>(createCategoryMutation);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useFormWithSchema(categorySchema, {
    defaultValues: {
      name: "",
      description: "",
      rank: 0,
    },
  });

  const onSubmit = handleSubmit((formData) => {
    createCategory({
      variables: {
        input: {
          cookieBannerId,
          name: formData.name,
          description: formData.description || null,
          rank: formData.rank,
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Category created successfully."),
          variant: "success",
        });
        reset();
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to create category"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={__("Add a cookie category")}
    >
      <form onSubmit={(e) => void onSubmit(e)}>
        <DialogContent className="p-6 space-y-4">
          <Field
            {...register("name")}
            label={__("Name")}
            type="text"
            error={errors.name?.message}
          />
          <Field
            {...register("description")}
            label={__("Description")}
            type="textarea"
            error={errors.description?.message}
          />
          <Field
            {...register("rank", { valueAsNumber: true })}
            label={__("Rank")}
            type="number"
            error={errors.rank?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting}>
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

const editCategorySchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  rank: z.number().min(0),
  cookies: z.array(
    z.object({
      name: z.string().min(1, "Cookie name is required"),
      duration: z.string().min(1, "Duration is required"),
      description: z.string(),
    }),
  ),
});

function EditCategoryDialog({
  children,
  category,
}: {
  children: React.ReactNode;
  category: CategoryNode;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [updateCategory] = useMutation<CookieBannerCategoriesTabUpdateMutation>(updateCategoryMutation);

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useFormWithSchema(editCategorySchema, {
    defaultValues: {
      name: category.name,
      description: category.description,
      rank: category.rank,
      cookies: category.cookies.map((c) => ({
        name: c.name,
        duration: c.duration,
        description: c.description,
      })),
    },
  });

  const cookies = watch("cookies");

  const addCookie = () => {
    setValue(
      "cookies",
      [...cookies, { name: "", duration: "", description: "" }],
      { shouldDirty: true },
    );
  };

  const removeCookie = (index: number) => {
    setValue(
      "cookies",
      cookies.filter((_, i) => i !== index),
      { shouldDirty: true },
    );
  };

  const onSubmit = handleSubmit((formData) => {
    updateCategory({
      variables: {
        input: {
          id: category.id,
          name: formData.name,
          description: formData.description || null,
          rank: formData.rank,
          cookies: formData.cookies,
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Category updated successfully."),
          variant: "success",
        });
        reset(formData);
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to update category"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={sprintf(__('Edit "%s"'), category.name)}
    >
      <form onSubmit={(e) => void onSubmit(e)}>
        <DialogContent className="max-h-[70vh] overflow-y-auto p-6 space-y-4">
          <Field
            {...register("name")}
            label={__("Name")}
            type="text"
            error={errors.name?.message}
          />
          <Field
            {...register("description")}
            label={__("Description")}
            type="textarea"
            error={errors.description?.message}
          />
          <Field
            {...register("rank", { valueAsNumber: true })}
            label={__("Rank")}
            type="number"
            error={errors.rank?.message}
          />

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-txt-primary">
                {__("Cookies")}
              </label>
              <button
                type="button"
                onClick={addCookie}
                className="text-sm text-txt-link hover:underline"
              >
                {__("+ Add cookie")}
              </button>
            </div>

            {cookies.map((_, index) => (
              <div
                key={index}
                className="rounded-lg border border-border p-3 space-y-2"
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-txt-secondary">
                    {sprintf(__("Cookie %d"), index + 1)}
                  </span>
                  <button
                    type="button"
                    onClick={() => removeCookie(index)}
                    className="text-xs text-txt-danger hover:underline"
                  >
                    {__("Remove")}
                  </button>
                </div>
                <Field
                  {...register(`cookies.${index}.name`)}
                  label={__("Name")}
                  type="text"
                  error={errors.cookies?.[index]?.name?.message}
                />
                <Field
                  {...register(`cookies.${index}.duration`)}
                  label={__("Duration")}
                  type="text"
                  placeholder="e.g. 1 year, Session"
                  error={errors.cookies?.[index]?.duration?.message}
                />
                <Field
                  {...register(`cookies.${index}.description`)}
                  label={__("Description")}
                  type="text"
                  error={errors.cookies?.[index]?.description?.message}
                />
              </div>
            ))}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting}>
            {__("Update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
