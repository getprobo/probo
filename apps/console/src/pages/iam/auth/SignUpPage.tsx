import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useMutation } from "react-relay";
import { Link, useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const signUpMutation = graphql`
  mutation SignUpPageMutation($input: SignUpInput!) {
    signUp(input: $input) {
      identity {
        id
      }
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  fullName: z.string().min(2),
});

type FormData = z.infer<typeof schema>;

export default function SignUpPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();

  usePageTitle(__("Sign up"));

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
      password: "",
      fullName: "",
    },
  });

  const [signUp] = useMutation(signUpMutation);

  const onSubmit = (data: FormData) => {
    signUp({
      variables: {
        input: {
          email: data.email,
          password: data.password,
          fullName: data.fullName,
        },
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Registration failed"),
            description: formatError(__("Registration failed"), e),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Account created successfully"),
          variant: "success",
        });
        void navigate("/", { replace: true });
      },
      onError: (e) => {
        toast({
          title: __("Registration failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Sign up")}</h1>
        <p className="text-txt-tertiary">
          {__("Enter your information to create an account")}
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
          label={__("Email")}
          type="email"
          placeholder={__("name@example.com")}
          {...register("email")}
          required
          error={formState.errors.email?.message}
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
            : __("Sign up with email")}
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
