import type { GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useEffect, useRef, useState } from "react";
import {
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { getPathPrefix } from "#/utils/pathPrefix";

import type { ConnectPageMutation } from "./__generated__/ConnectPageMutation.graphql";
import type { ConnectPageQuery } from "./__generated__/ConnectPageQuery.graphql";

export const connectPageQuery = graphql`
  query ConnectPageQuery {
    currentTrustCenter @required(action: THROW) {
      organization @required(action: THROW) {
        name
      }
    }
  }
`;

const sendMagicLinkMutation = graphql`
  mutation ConnectPageMutation($input: SendMagicLinkInput!) {
    sendMagicLink(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
});

type FormData = z.infer<typeof schema>;

const timerDurationSeconds = 60;

export function ConnectPage(props: {
  queryRef: PreloadedQuery<ConnectPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();
  const [magicLinkSent, setMagicLinkSent] = useState<boolean>(false);
  const interval = useRef<NodeJS.Timeout>(undefined);
  const [timer, setTimer] = useState<number>(timerDurationSeconds);

  const {
    currentTrustCenter: { organization },
  } = usePreloadedQuery<ConnectPageQuery>(connectPageQuery, queryRef);

  useEffect(() => {
    if (!magicLinkSent && interval.current) {
      clearInterval(interval.current);
      interval.current = undefined;
    }
    if (magicLinkSent) {
      clearInterval(interval.current);
      interval.current = setInterval(() => {
        setTimer(timer => Math.max(timer - 1, 0));
      }, 1000);
    }

    return () => {
      clearInterval(interval.current);
    };
  }, [magicLinkSent]);

  usePageTitle(__(`Connect to ${organization.name}'s Compliance Page`));

  const {
    handleSubmit: handleSubmitWrapper,
    register,
    formState,
  } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
    },
  });

  const [sendMagicLink] = useMutation<ConnectPageMutation>(
    sendMagicLinkMutation,
  );

  const handleSubmit = handleSubmitWrapper(({ email }: FormData) => {
    sendMagicLink({
      variables: {
        input: {
          email,
        },
      },
      onCompleted: (data, errors: GraphQLError[] | null) => {
        console.log(data, errors);
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ALREADY_AUTHENTICATED") {
              const pathPrefix = getPathPrefix();
              window.location.href = pathPrefix ? getPathPrefix() : "/";
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
          description: __("Magic link sent!"),
          variant: "success",
        });
        setTimer(timerDurationSeconds);
        setMagicLinkSent(true);
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
          {__(`Connect to ${organization.name}'s Compliance Page`)}
        </h1>
        <p className="text-txt-tertiary">
          {__(
            "Enter your email address to connect with a magic link and start requesting access to documents",
          )}
        </p>
      </div>

      <form onSubmit={e => void handleSubmit(e)} className="space-y-6">
        <Field
          label={__("Email")}
          placeholder="john.doe@acme.com"
          {...register("email")}
          type="email"
          error={formState.errors.email?.message}
        />

        {magicLinkSent && (
          <p className="text-txt-primary text-sm">
            {__(
              "Magic Link Sent! Check your emails and use the link to connect.",
            )}
          </p>
        )}

        <Button
          type="submit"
          className="w-xs h-10 mx-auto"
          disabled={formState.isSubmitting || (magicLinkSent && timer !== 0)}
        >
          {magicLinkSent
            ? timer === 0
              ? __("Resend Link")
              : `${__("Resend Link in")} ${timer}s`
            : __("Send Magic Link")}
        </Button>
      </form>
    </div>
  );
}
