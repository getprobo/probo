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

import { PencilSimpleIcon } from "@phosphor-icons/react";
import { Breadcrumb, Dialog, useDialogRef } from "@probo/ui";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { useQueryLoader } from "react-relay";

import type { EditSAMLConfigurationFormQuery } from "#/__generated__/iam/EditSAMLConfigurationFormQuery.graphql";

import {
  EditSAMLConfigurationForm,
  samlConfigurationFormQuery,
} from "./EditSAMLConfigurationForm";

interface EditSAMLConfigurationDialogProps {
  samlConfigurationId: string;
}

export function EditSAMLConfigurationDialog({
  samlConfigurationId,
}: EditSAMLConfigurationDialogProps) {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [queryRef, loadQuery] = useQueryLoader<EditSAMLConfigurationFormQuery>(
    samlConfigurationFormQuery,
  );

  function handleOpen() {
    loadQuery({ samlConfigurationId }, { fetchPolicy: "network-only" });
    dialogRef.current?.open();
  }

  function handleClose() {
    dialogRef.current?.close();
  }

  return (
    <>
      <IconButton
        size={1}
        variant="surface"
        color="neutral"
        aria-label={t("samlConfigurationList.actions.edit")}
        onClick={handleOpen}
      >
        <PencilSimpleIcon />
      </IconButton>
      <Dialog
        ref={dialogRef}
        onClose={handleClose}
        title={(
          <Breadcrumb
            items={[
              t("samlSsoPage.breadcrumb.settings"),
              t("samlSsoPage.breadcrumb.configure"),
            ]}
          />
        )}
      >
        <Suspense>
          {queryRef && (
            <EditSAMLConfigurationForm
              queryRef={queryRef}
              onUpdate={handleClose}
            />
          )}
        </Suspense>
      </Dialog>
    </>
  );
}
