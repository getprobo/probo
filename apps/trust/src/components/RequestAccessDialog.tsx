import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { useTrustCenter } from "/hooks/useTrustCenter";

type Props = {
  children: React.ReactNode;
};

const schema = z.object({
  name: z.string(),
  email: z.string().email(),
});

const requestAccessMutation = graphql`
  mutation RequestAccessDialogMutation($input: CreateTrustCenterAccessInput!) {
    createTrustCenterAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

export function RequestAccessDialog({ children }: Props) {
  const trustCenter = useTrustCenter();
  const { toast } = useToast();
  const { __ } = useTranslate();
  const { handleSubmit, register } = useFormWithSchema(schema, {
    defaultValues: {
      name: "",
      email: "",
    },
  });
  const dialogRef = useDialogRef();
  const [commitRequestAccess, isRequestingAccess] = useMutationWithToasts(
    requestAccessMutation,
    {
      onSuccess: () => {
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully"),
          variant: "success",
        });
        dialogRef.current?.close();
      },
    }
  );

  const onSubmit = handleSubmit((data) => {
    commitRequestAccess({
      variables: {
        input: {
          trustCenterId: trustCenter.id,
          name: data.name,
          email: data.email,
        },
      },
    });
  });
  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      className="max-w-[500px] text-txt-primary"
    >
      <form onSubmit={onSubmit}>
        <DialogTitle className="text-2xl font-semibold mb-4 pt-4 md:pt-8 px-4 md:px-8">
          {__("Request access to documentation")}
        </DialogTitle>
        <DialogContent className="px-4 md:px-8 pb-4 md:pb-8 text-txt-primary">
          <p className="text-txt-secondary mb-4">
            {sprintf(
              __(
                "Request access to %s's Trust Center. Your request will be reviewed and you will receive an email notification with access instructions if approved."
              ),
              name
            )}
          </p>
          <div className="space-y-4">
            <Field
              label={__("Full name")}
              placeholder="John Doe"
              {...register("name")}
              type="text"
            />
            <Field
              label={__("Email")}
              placeholder="john.doe@acme.com"
              {...register("email")}
              type="email"
            />
          </div>
        </DialogContent>
        <DialogFooter>
          <Button disabled={isRequestingAccess} type="submit">
            {__("Continue")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
