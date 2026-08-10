// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Card, useToast } from "@probo/ui";
import { Suspense, useState } from "react";
import { useLazyLoadQuery, useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { SlackBindPageLoaderConfirmMutation } from "#/__generated__/core/SlackBindPageLoaderConfirmMutation.graphql";
import type { SlackBindPageLoaderQuery } from "#/__generated__/core/SlackBindPageLoaderQuery.graphql";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

const slackBindPageQuery = graphql`
  query SlackBindPageLoaderQuery($token: String!) {
    slackIdentityBindPreview(token: $token) {
      teamId
      slackUserId
    }
  }
`;

const confirmSlackIdentityBindingMutation = graphql`
  mutation SlackBindPageLoaderConfirmMutation($input: ConfirmSlackIdentityBindingInput!) {
    confirmSlackIdentityBinding(input: $input) {
      slackIdentityBinding {
        id
        teamId
        slackUserId
      }
    }
  }
`;

function SlackBindPageContent() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token")?.trim() ?? "";
  const [confirmed, setConfirmed] = useState(false);

  usePageTitle(__("Link Slack account"));

  const data = useLazyLoadQuery<SlackBindPageLoaderQuery>(
    slackBindPageQuery,
    { token },
    { fetchPolicy: "network-only" },
  );

  const [confirmBinding, isConfirming] = useMutation<SlackBindPageLoaderConfirmMutation>(
    confirmSlackIdentityBindingMutation,
  );

  if (token === "") {
    return (
      <div className="space-y-4 w-full max-w-md mx-auto pt-8 text-center">
        <h1 className="text-3xl font-bold">{__("Link Slack account")}</h1>
        <p className="text-txt-tertiary">{__("This link is missing a token.")}</p>
      </div>
    );
  }

  const preview = data.slackIdentityBindPreview;

  const handleConfirm = () => {
    confirmBinding({
      variables: { input: { token } },
      onCompleted: (_response, errors: GraphQLError[] | null) => {
        if (errors) {
          toast({
            title: __("Link failed"),
            description: formatError(__("Link failed"), errors),
            variant: "error",
          });

          return;
        }

        setConfirmed(true);
        toast({
          title: __("Success"),
          description: __("Your Slack account is now linked to Probo."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Link failed"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Link Slack account")}</h1>
        <p className="text-txt-tertiary">
          {__(
            "Confirm that this Slack user should use your Probo identity in Slack.",
          )}
        </p>
      </div>

      <Card className="p-6 space-y-4">
        <dl className="space-y-3 text-sm">
          <div className="flex justify-between gap-4">
            <dt className="text-txt-tertiary">{__("Slack workspace")}</dt>
            <dd className="font-mono text-right break-all">{preview.teamId}</dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="text-txt-tertiary">{__("Slack user")}</dt>
            <dd className="font-mono text-right break-all">{preview.slackUserId}</dd>
          </div>
        </dl>

        {confirmed
          ? (
              <p className="text-txt-secondary text-sm">
                {__(
                  "Linked. You can return to Slack and message the bot again.",
                )}
              </p>
            )
          : (
              <Button
                className="w-full"
                disabled={isConfirming}
                onClick={handleConfirm}
              >
                {__("Confirm link")}
              </Button>
            )}
      </Card>

      <div className="text-center text-sm text-txt-secondary">
        <Link
          to="/"
          className="underline hover:text-txt-primary"
          onClick={(event) => {
            event.preventDefault();
            void navigate("/");
          }}
        >
          {__("Back to Probo")}
        </Link>
      </div>
    </div>
  );
}

function SlackBindPageFallback() {
  const { __ } = useTranslate();

  return (
    <div className="space-y-4 w-full max-w-md mx-auto pt-8 text-center">
      <h1 className="text-3xl font-bold">{__("Link Slack account")}</h1>
      <p className="text-txt-tertiary">{__("Loading…")}</p>
    </div>
  );
}

export default function SlackBindPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<SlackBindPageFallback />}>
        <SlackBindPageContent />
      </Suspense>
    </CoreRelayProvider>
  );
}
