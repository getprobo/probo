import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const signUpFromIvitationMutation = graphql`
  mutation SignUpFromInvitationPageMutation(
    $input: SignUpFromInvitationInput!
  ) {
    signUpFromInvitation(input: $input) {
      identity {
        id
      }
    }
  }
`;

const schema = z.object({
  password: z.string().min(8),
  fullName: z.string().min(2),
});

type FormData = z.infer<typeof schema>;

export default function SignUpFromInvitationPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const fullNameFromParams = searchParams.get("fullName") || "";

  usePageTitle(__("Sign up"));

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      password: "",
      fullName: fullNameFromParams,
    },
  });

  const [signUpFromInvitation] = useMutation(signUpFromIvitationMutation);

  const onSubmit = (data: FormData) => {
    const token = searchParams.get("token");
    if (!token) {
      toast({
        title: __("Signup failed"),
        description: __("Invalid or missing invitation token"),
        variant: "error",
      });
      return;
    }

    signUpFromInvitation({
      variables: {
        input: {
          token,
          password: data.password,
          fullName: data.fullName,
        },
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Signup failed"),
            description: formatError(__("Signup failed"), e),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __(
            "Account created successfully. Please accept your invitation to join the organization.",
          ),
          variant: "success",
        });
        void navigate("/", { replace: true });
      },
      onError: (e) => {
        toast({
          title: __("Signup failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Create your account")}</h1>
        <p className="text-txt-tertiary">
          {__("Set your password to join the organization")}
        </p>
      </div>

      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <Field
          label={__("Full Name")}
          type="text"
          placeholder={__("John Doe")}
          {...register("fullName")}
          required
          error={formState.errors.fullName?.message}
        />

        <Field
          label={__("Password")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Button type="submit" className="w-xs h-10 mx-auto mt-6" disabled={formState.isLoading}>
          {formState.isLoading
            ? __("Creating account...")
            : __("Create account")}
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
