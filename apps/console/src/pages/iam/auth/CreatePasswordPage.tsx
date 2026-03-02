import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreatePasswordPageMutation } from "#/__generated__/iam/CreatePasswordPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const createPasswordMutation = graphql`
  mutation CreatePasswordPageMutation($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  password: z.string().min(8),
});

export default function CreatePasswordPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  usePageTitle(__("Create Password"));

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      password: "",
    },
  });

  const [createPassword, isCreatingPassword] = useMutation<CreatePasswordPageMutation>(createPasswordMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    createPassword({
      variables: {
        input: {
          password: data.password,
          token: searchParams.get("token") ?? "",
        },
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Password creation failed"),
            description: formatError(__("Password creation failed"), e),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Account created successfully"),
          variant: "success",
        });
        void navigate("/auth/login", { replace: true });
      },
      onError: (e) => {
        toast({
          title: __("Password creation failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Create a password")}</h1>
        <p className="text-txt-tertiary">
          {__("Set a password for your account, with at least 8 characters")}
        </p>
      </div>

      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <Field
          label={__("Password")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Button type="submit" className="w-xs h-10 mx-auto mt-6" disabled={formState.isLoading || isCreatingPassword}>
          {__("Save")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {__("Already have an account?")}
          {" "}
          <Link
            to="/auth/login"
            className="underline text-txt-primary hover:text-txt-secondary"
          >
            {__("Log in here")}
          </Link>
        </p>
      </div>
    </div>
  );
}
