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
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useTranslate } from "@probo/i18n";
import type { MagicLinkDialogMutation } from "./__generated__/MagicLinkDialogMutation.graphql";
import type { PropsWithChildren } from "react";

const sendMagicLinkMutation = graphql`
  mutation MagicLinkDialogMutation($input: SendMagicLinkInput!) {
    sendMagicLink(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
});

type FormData = z.infer<typeof schema>;

export function MagicLinkDialog(props: PropsWithChildren) {
  const { children } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [sendMagicLink] = useMutation<MagicLinkDialogMutation>(
    sendMagicLinkMutation,
  );

  const {
    handleSubmit: handleSubmitWrapper,
    register,
    formState,
  } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
    },
  });

  const handleSubmit = ({ email }: FormData) => {
    sendMagicLink({
      variables: {
        input: {
          email,
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: __("Cannot send magic link"),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __(
            "Magic link sent! Please check your emails to authenticate",
          ),
          variant: "success",
        });
        dialogRef.current?.close();
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      trigger={children}
      ref={dialogRef}
      className="max-w-[500px] text-txt-primary"
    >
      <form onSubmit={handleSubmitWrapper(handleSubmit)}>
        <DialogTitle className="text-2xl font-semibold mb-4 pt-4 md:pt-8 px-4 md:px-8">
          {__("Create account or sign in")}
        </DialogTitle>
        <DialogContent className="px-4 md:px-8 pb-4 md:pb-8 text-txt-primary">
          <p className="text-txt-secondary mb-4">
            {__(
              "To be able to request access to documents, you need to authenticate. Please enter your email to get a magic link.",
            )}
          </p>
          <Field
            label={__("Email")}
            placeholder="john.doe@acme.com"
            {...register("email")}
            type="email"
            error={formState.errors.email?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {__("Continue")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
