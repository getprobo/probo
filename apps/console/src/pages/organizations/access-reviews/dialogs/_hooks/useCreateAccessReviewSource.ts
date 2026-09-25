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

import { useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";

import { integrationListPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";

type UseCreateAccessReviewSourceParams = {
  organizationId: string;
  onSuccess: () => void;
};

// After a connector is created, open its Settings page. Connecting does not
// create an access-review source.
export function useCreateAccessReviewSource({
  organizationId,
  onSuccess,
}: UseCreateAccessReviewSourceParams) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const navigate = useNavigate();

  const openConnector = (
    _connectorId: string,
    _displayName: string,
    onDone: () => void,
  ) => {
    onDone();
    toast({
      title: t("listPage.messages.connected"),
      description: t("listPage.messages.connectedDescription"),
      variant: "success",
    });
    onSuccess();
    void navigate(integrationListPath(organizationId));
  };

  return openConnector;
}
