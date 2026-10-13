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

import { downloadFile } from "@probo/helpers";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { useDownloadDeviceEvidenceMutation } from "#/__generated__/core/useDownloadDeviceEvidenceMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const downloadDeviceEvidenceMutation = graphql`
  mutation useDownloadDeviceEvidenceMutation($input: DownloadDeviceEvidenceInput!) {
    downloadDeviceEvidence(input: $input) {
      downloadUrl
    }
  }
`;

const evidenceFileName = "device-evidence.pdf";

export function useDownloadDeviceEvidence() {
  const { t } = useTranslation("organizations/devices");
  const [commit, isDownloading] = useMutation<useDownloadDeviceEvidenceMutation>(
    downloadDeviceEvidenceMutation,
  );

  const download = useCallback(async (deviceIds: readonly string[]) => {
    const response = await commit({
      variables: {
        input: {
          deviceIds: [...deviceIds],
        },
      },
    }, {
      errorToast: t("devices.errors.downloadEvidence"),
    });
    downloadFile(response.downloadDeviceEvidence.downloadUrl, evidenceFileName);
  }, [commit, t]);

  return { download, isDownloading };
}
