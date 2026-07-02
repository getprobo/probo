// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
    alias
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
  const filePath = file.alias ?? file.id;
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
              onClick={() => void navigate(`/documents/${filePath}`)}
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
