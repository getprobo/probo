import type { GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import {
  useMutation,
} from "react-relay";
import { useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { getPathPrefix } from "#/utils/pathPrefix";

import type { FullNamePageMutation } from "./__generated__/FullNamePageMutation.graphql";

const updateMutation = graphql`
  mutation FullNamePageMutation($input: UpdateFullNameInput!) {
    updateFullName(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  fullName: z.string().min(2),
});

type FormData = z.infer<typeof schema>;

export default function FullNamePage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();

  const continueUrlParam = searchParams.get("continue");
  let safeContinueUrl: string;
  if (continueUrlParam) {
    try {
      const continueUrl = new URL(continueUrlParam, window.location.origin);
      if (continueUrl.origin === window.location.origin && continueUrl.pathname.startsWith(`${getPathPrefix()}/`)) {
        safeContinueUrl = window.location.origin + continueUrl.pathname + continueUrl.search;
      } else {
        safeContinueUrl = window.location.origin + (getPathPrefix() || "/");
      }
    } catch {
      safeContinueUrl = window.location.origin + (getPathPrefix() || "/");
    }
  } else {
    safeContinueUrl = window.location.origin + (getPathPrefix() || "/");
  }

  const {
    handleSubmit: handleSubmitWrapper,
    register,
    formState,
  } = useFormWithSchema(schema, {
    defaultValues: {
      fullName: "",
    },
  });

  const [update] = useMutation<FullNamePageMutation>(
    updateMutation,
  );

  const handleSubmit = handleSubmitWrapper(({ fullName }: FormData) => {
    update({
      variables: {
        input: {
          fullName,
        },
      },
      onCompleted: (_, errors: GraphQLError[] | null) => {
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ALREADY_AUTHENTICATED") {
              window.location.href = getPathPrefix() || "/";
              return;
            }
          }
          toast({
            title: __("Error"),
            description: __("Cannot send magic link"),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Full name updated!"),
          variant: "success",
        });

        window.location.href = safeContinueUrl;
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  });

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">
          {__("Please set your profile's full name")}
        </h1>
      </div>

      <form onSubmit={e => void handleSubmit(e)} className="space-y-6">
        <Field
          label={__("Full Name")}
          placeholder="John Doe"
          {...register("fullName")}
          type="text"
          required
          error={formState.errors.fullName?.message}
        />

        <Button
          type="submit"
          className="w-xs h-10 mx-auto"
          disabled={formState.isSubmitting}
        >
          {__("Continue")}
        </Button>
      </form>
    </div>
  );
}
