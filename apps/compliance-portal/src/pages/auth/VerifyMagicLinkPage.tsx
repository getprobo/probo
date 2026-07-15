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
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getSafeContinueUrl } from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { VerifyMagicLinkPageMutation } from "./__generated__/VerifyMagicLinkPageMutation.graphql";

const verifyMagicLinkMutation = graphql`
  mutation VerifyMagicLinkPageMutation($input: VerifyMagicLinkInput!) {
    verifyMagicLink(input: $input) {
      continue
    }
  }
`;

// Landing page for the magic-link email. It verifies the token on mount and
// forwards to the (validated) continue URL, where any pending access request
// resumes.
export default function VerifyMagicLinkPage() {
  const { t } = useTranslation();
  const toast = Toast.useToastManager();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const submittedRef = useRef(false);

  const [verifyMagicLink] = useMutation<VerifyMagicLinkPageMutation>(
    verifyMagicLinkMutation,
    { errorToast: false },
  );

  useEffect(() => {
    const token = searchParams.get("token");
    if (!token || submittedRef.current) {
      return;
    }
    submittedRef.current = true;

    void verifyMagicLink({
      variables: { input: { token: token.trim() } },
      onCompleted: (response, errors) => {
        const code = (errors?.[0] as GraphQLError | undefined)?.extensions?.code;

        if (code === "ALREADY_AUTHENTICATED") {
          window.location.href = getSafeContinueUrl(null);
          return;
        }
        if (code === "TOKEN_EXPIRED") {
          void navigate("/magic-link-expired");
          return;
        }
        if (code === "TOKEN_ALREADY_USED") {
          void navigate("/magic-link-already-used");
          return;
        }
        if (errors && errors.length > 0) {
          toast.add({ title: t("auth.errors.verifyFailed"), type: "error" });
          return;
        }

        window.location.href = getSafeContinueUrl(response.verifyMagicLink?.continue);
      },
      onError: () => {
        toast.add({ title: t("auth.errors.verifyFailed"), type: "error" });
      },
    }).catch(() => {});
  }, [navigate, searchParams, t, toast, verifyMagicLink]);

  return (
    <div className="flex flex-col items-center gap-2 text-center">
      <Heading level={1} size={5}>{t("auth.verify.title")}</Heading>
      <Text color="neutral">{t("auth.verify.description")}</Text>
    </div>
  );
}
