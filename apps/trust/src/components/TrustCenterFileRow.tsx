// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { UnAuthenticatedError } from "@probo/relay";
import {
  Button,
  IconArrowLink,
  IconLock,
  IconPageTextLine,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

import type { TrustCenterFileRow_requestAccessMutation } from "./__generated__/TrustCenterFileRow_requestAccessMutation.graphql";
import type { TrustCenterFileRowFragment$key } from "./__generated__/TrustCenterFileRowFragment.graphql";

const requestAccessMutation = graphql`
  mutation TrustCenterFileRow_requestAccessMutation(
    $input: RequestTrustCenterFileAccessInput!
  ) {
    requestTrustCenterFileAccess(input: $input) {
      file {
        access {
          id
          status
        }
      }
    }
  }
`;

const trustCenterFileRowFragment = graphql`
  fragment TrustCenterFileRowFragment on TrustCenterFile {
    id
    name
    isUserAuthorized
    access {
      id
      status
    }
  }
`;

export function TrustCenterFileRow(props: {
  file: TrustCenterFileRowFragment$key;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const file = useFragment(trustCenterFileRowFragment, props.file);
  const hasRequested = file.access?.status === "REQUESTED";

  const [requestAccess, isRequestingAccess]
    = useMutation<TrustCenterFileRow_requestAccessMutation>(
      requestAccessMutation,
    );

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          trustCenterFileId: file.id,
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot request access"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully."),
          variant: "success",
        });
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          const pathPrefix = getPathPrefix();
          searchParams.set("request-file-id", file.id);
          const urlSearchParams = new URLSearchParams([[
            "continue",
            window.location.origin + pathPrefix + location.pathname + "?" + searchParams.toString(),
          ]]);
          void navigate(`/connect?${urlSearchParams.toString()}`);

          return;
        }

        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot request access"),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconPageTextLine size={16} className=" flex-none text-txt-tertiary" />
        {file.name}
      </div>
      {file.isUserAuthorized
        ? (
            <Button
              className="w-full md:w-max"
              variant="secondary"
              icon={IconArrowLink}
              onClick={() => void navigate(`/documents/${file.id}`)}
            >
              {__("View")}
            </Button>
          )
        : (
            <Button
              disabled={hasRequested || isRequestingAccess}
              className="w-full md:w-max"
              variant="secondary"
              icon={IconLock}
              onClick={handleRequestAccess}
            >
              {hasRequested ? __("Access requested") : __("Request access")}
            </Button>
          )}
    </div>
  );
}
