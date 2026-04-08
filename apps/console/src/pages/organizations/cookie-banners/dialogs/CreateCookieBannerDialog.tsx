import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateCookieBannerDialogMutation } from "#/__generated__/core/CreateCookieBannerDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const createCookieBannerMutation = graphql`
  mutation CreateCookieBannerDialogMutation(
    $input: CreateCookieBannerInput!
    $connections: [ID!]!
  ) {
    createCookieBanner(input: $input) {
      cookieBannerEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          domain
          state
          version
          createdAt
          updatedAt
          canUpdate: permission(action: "core:cookie-banner:update")
          canDelete: permission(action: "core:cookie-banner:delete")
        }
      }
    }
  }
`;

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  domain: z.string().optional(),
});

type Props = {
  children: ReactNode;
  organizationId: string;
  connection: string;
};

export function CreateCookieBannerDialog({
  children,
  organizationId,
  connection,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [createCookieBanner] = useMutation<CreateCookieBannerDialogMutation>(createCookieBannerMutation);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useFormWithSchema(schema, {
    defaultValues: {
      name: "",
      domain: "",
    },
  });

  const onSubmit = handleSubmit((formData) => {
    createCookieBanner({
      variables: {
        input: {
          organizationId,
          name: formData.name,
          domain: formData.domain || null,
        },
        connections: [connection],
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Cookie banner created successfully."),
          variant: "success",
        });
        reset();
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to create cookie banner"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={__("Add a cookie banner")}
    >
      <form onSubmit={e => void onSubmit(e)}>
        <DialogContent className="p-6 space-y-4">
          <Field
            {...register("name")}
            label={__("Name")}
            type="text"
            error={errors.name?.message}
            placeholder={__("e.g. Main Website")}
          />
          <Field
            {...register("domain")}
            label={__("Domain")}
            type="text"
            error={errors.domain?.message}
            placeholder={__("e.g. example.com")}
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
