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

import { Toast } from "@base-ui/react/toast";
import type { GraphQLError } from "@probo/helpers";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import {
  buildRequestAllContinueUrl,
  REQUEST_ALL_PARAM,
} from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { useResumeAccessRequestMutation } from "./__generated__/useResumeAccessRequestMutation.graphql";

const requestAllAccessesMutation = graphql`
  mutation useResumeAccessRequestMutation {
    requestAllAccesses {
      trustCenterAccess {
        id
      }
    }
  }
`;

// After a user signs in through the dialog, they land back on the page that
// carried the request-all marker. This hook fires the deferred
// `requestAllAccesses` mutation once (when authenticated), routes to the
// full-name gate when the backend asks for it, and clears the marker so a
// refresh never re-triggers it.
export function useResumeAccessRequest(isAuthenticated: boolean) {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  const firedRef = useRef(false);

  const [requestAllAccesses] = useMutation<useResumeAccessRequestMutation>(
    requestAllAccessesMutation,
    { errorToast: false },
  );

  const shouldResume
    = isAuthenticated && searchParams.get(REQUEST_ALL_PARAM) === "true";

  useEffect(() => {
    if (!shouldResume || firedRef.current) {
      return;
    }
    firedRef.current = true;

    // Drop the marker up front so a reload can't queue a second request.
    searchParams.delete(REQUEST_ALL_PARAM);
    setSearchParams(searchParams, { replace: true });

    void requestAllAccesses({
      variables: {},
      onCompleted: (_response, errors) => {
        const code = (errors?.[0] as GraphQLError | undefined)?.extensions?.code;

        // The backend gates access behind a completed profile; send the user to
        // the full-name step, preserving the marker so the request resumes.
        if (code === "FULL_NAME_REQUIRED") {
          const continueUrl = buildRequestAllContinueUrl();
          void navigate(`/full-name?continue=${encodeURIComponent(continueUrl)}`);
          return;
        }

        if (errors && errors.length > 0) {
          toast.add({
            title:
              code === "NDA_SIGNATURE_REQUIRED"
                ? t("auth.errors.ndaRequired")
                : t("auth.errors.requestFailed"),
            type: "error",
          });
          return;
        }

        toast.add({ title: t("auth.requestAccess.success"), type: "success" });
      },
      onError: () => {
        toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
      },
      // The awaitable wrapper rejects on failure; toasts are handled above, so
      // swallow the rejection to avoid an unhandled promise.
    }).catch(() => {});
  }, [
    shouldResume,
    navigate,
    requestAllAccesses,
    searchParams,
    setSearchParams,
    t,
    toast,
  ]);
}
