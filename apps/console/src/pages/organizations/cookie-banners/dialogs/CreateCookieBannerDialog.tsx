import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { z } from "zod";

import { useCreateCookieBannerMutation } from "#/hooks/graph/CookieBannerGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

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
  const dialogRef = useDialogRef();
  const [createCookieBanner] = useCreateCookieBannerMutation();

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

  const onSubmit = handleSubmit(async (formData) => {
    await createCookieBanner({
      variables: {
        input: {
          organizationId,
          name: formData.name,
          domain: formData.domain || null,
        },
        connections: [connection],
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
