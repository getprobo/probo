import { promisifyMutation, sprintf } from "@probo/helpers";
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
  IconPlusLarge,
  IconTrashCan,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { graphql, useFragment, useMutation } from "react-relay";
import { useOutletContext } from "react-router";
import { z } from "zod";

import type { CookieBannerCategoriesTabFragment$key } from "#/__generated__/core/CookieBannerCategoriesTabFragment.graphql";
import type { CookieBannerGraphNodeQuery$data } from "#/__generated__/core/CookieBannerGraphNodeQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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

const deleteCategoryMutation = graphql`
  mutation CookieBannerCategoriesTabDeleteMutation(
    $input: DeleteCookieCategoryInput!
  ) {
    deleteCookieCategory(input: $input) {
      deletedCookieCategoryId
    }
  }
`;

export default function CookieBannerCategoriesTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerGraphNodeQuery$data["node"];
  }>();

  const { __ } = useTranslate();
  const data = useFragment<CookieBannerCategoriesTabFragment$key>(
    fragment,
    banner,
  );

  const categories = data.categories?.edges.map(edge => edge.node) ?? [];

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
        {categories.map(category => (
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
  category: {
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
  canUpdate: boolean;
}) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  // eslint-disable-next-line relay/generated-typescript-types
  const [deleteCategory] = useMutation(deleteCategoryMutation);

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteCategory)({
          variables: {
            input: { id: category.id },
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete category \"%s\". This action cannot be undone.",
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
                  {category.cookies.map(cookie => (
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
        {canUpdate && !category.required && (
          <ActionDropdown>
            <DropdownItem
              onClick={handleDelete}
              variant="danger"
              icon={IconTrashCan}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>
    </Card>
  );
}

const createCategorySchema = z.object({
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
  const dialogRef = useDialogRef();

  const [createCategory] = useMutationWithToasts(createCategoryMutation, {
    successMessage: __("Category created successfully."),
    errorMessage: __("Failed to create category"),
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useFormWithSchema(createCategorySchema, {
    defaultValues: {
      name: "",
      description: "",
      rank: 0,
    },
  });

  const onSubmit = handleSubmit(async (formData) => {
    await createCategory({
      variables: {
        input: {
          cookieBannerId,
          name: formData.name,
          description: formData.description || null,
          rank: formData.rank,
        },
      },
      onSuccess: () => {
        reset();
        dialogRef.current?.close();
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={__("Add a cookie category")}
    >
      <form onSubmit={e => void onSubmit(e)}>
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
