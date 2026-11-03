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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { RequestScanButton_cookieBanner$key } from "#/__generated__/core/RequestScanButton_cookieBanner.graphql";
import type { RequestScanButtonMutation } from "#/__generated__/core/RequestScanButtonMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const requestScanButtonFragment = graphql`
  fragment RequestScanButton_cookieBanner on CookieBanner {
    id
    canUpdate: permission(action: "core:cookie-banner:update")
    activeScan {
      id
    }
  }
`;

const requestScanMutation = graphql`
  mutation RequestScanButtonMutation($input: RequestCookieBannerScanInput!) {
    requestCookieBannerScan(input: $input) {
      cookieBannerScan {
        id
        state
      }
      cookieBanner {
        id
        activeScan {
          id
          state
        }
      }
    }
  }
`;

interface RequestScanButtonProps {
  cookieBannerKey: RequestScanButton_cookieBanner$key;
  onRequested: () => void;
}

export function RequestScanButton({ cookieBannerKey, onRequested }: RequestScanButtonProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(requestScanButtonFragment, cookieBannerKey);
  const [requestScan, isRequesting] = useMutation<RequestScanButtonMutation>(requestScanMutation, {
    successMessage: t("scansPage.messages.requested"),
    errorToast: t("scansPage.errors.request"),
  });

  if (!banner.canUpdate) {
    return null;
  }

  function handleRequest() {
    void requestScan({
      variables: { input: { cookieBannerId: banner.id } },
    }).then(() => {
      onRequested();
    }, () => undefined);
  }

  return (
    <Button
      type="button"
      size={2}
      variant="solid"
      color="neutral"
      highContrast
      disabled={banner.activeScan != null}
      loading={isRequesting}
      onClick={handleRequest}
    >
      {t("scansPage.actions.scan")}
    </Button>
  );
}
