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
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import type { FormEvent } from "react";
import { Suspense, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";

import type { UpdateServiceAccountDialog_serviceAccount$key } from "#/__generated__/iam/UpdateServiceAccountDialog_serviceAccount.graphql";
import type { UpdateServiceAccountDialogMutation } from "#/__generated__/iam/UpdateServiceAccountDialogMutation.graphql";
import type { UpdateServiceAccountDialogQuery } from "#/__generated__/iam/UpdateServiceAccountDialogQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { ScopePicker } from "./ScopePicker";

const updateServiceAccountDialogFragment = graphql`
  fragment UpdateServiceAccountDialog_serviceAccount on ServiceAccount {
    id
    name
    description
    scopes
  }
`;

const updateServiceAccountDialogQuery = graphql`
  query UpdateServiceAccountDialogQuery {
    oauth2ScopesSupported
  }
`;

const updateServiceAccountMutation = graphql`
  mutation UpdateServiceAccountDialogMutation(
    $input: UpdateServiceAccountInput!
  ) {
    updateServiceAccount(input: $input) {
      serviceAccount {
        ...ServiceAccountListItem_serviceAccount
      }
    }
  }
`;

interface UpdateServiceAccountDialogProps {
  open: boolean;
  serviceAccountKey: UpdateServiceAccountDialog_serviceAccount$key;
  onOpenChange: (open: boolean) => void;
}

export function UpdateServiceAccountDialog(
  props: UpdateServiceAccountDialogProps,
) {
  const { open, serviceAccountKey, onOpenChange } = props;
  const [queryRef, loadQuery, disposeQuery]
    = useQueryLoader<UpdateServiceAccountDialogQuery>(
      updateServiceAccountDialogQuery,
    );

  useEffect(() => {
    if (open) {
      loadQuery({}, { fetchPolicy: "store-or-network" });
    } else {
      disposeQuery();
    }
  }, [disposeQuery, loadQuery, open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup className="max-w-3xl">
        <Suspense fallback={<UpdateServiceAccountDialogLoading />}>
          {queryRef && (
            <UpdateServiceAccountDialogForm
              queryRef={queryRef}
              serviceAccountKey={serviceAccountKey}
              onUpdated={() => onOpenChange(false)}
            />
          )}
        </Suspense>
      </DialogPopup>
    </Dialog>
  );
}

function UpdateServiceAccountDialogLoading() {
  const { t } = useTranslation("iam/organizations/service-accounts");
  return (
    <>
      <DialogHeader>
        <DialogTitle>{t("update.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className="h-48 animate-pulse rounded-3 bg-sand-3" />
    </>
  );
}

interface UpdateServiceAccountDialogFormProps {
  queryRef: PreloadedQuery<UpdateServiceAccountDialogQuery>;
  serviceAccountKey: UpdateServiceAccountDialog_serviceAccount$key;
  onUpdated: () => void;
}

function UpdateServiceAccountDialogForm(
  props: UpdateServiceAccountDialogFormProps,
) {
  const { queryRef, serviceAccountKey, onUpdated } = props;
  const { t } = useTranslation("iam/organizations/service-accounts");
  const serviceAccount = useFragment(
    updateServiceAccountDialogFragment,
    serviceAccountKey,
  );
  const data = usePreloadedQuery<UpdateServiceAccountDialogQuery>(
    updateServiceAccountDialogQuery,
    queryRef,
  );
  const supportedScopes = [...data.oauth2ScopesSupported].sort();
  const [name, setName] = useState(serviceAccount.name);
  const [description, setDescription] = useState(
    serviceAccount.description ?? "",
  );
  const [scopes, setScopes] = useState<string[]>([...serviceAccount.scopes]);
  const [scopeError, setScopeError] = useState(false);
  const [updateServiceAccount, isUpdating]
    = useMutation<UpdateServiceAccountDialogMutation>(
      updateServiceAccountMutation,
      {
        successMessage: t("messages.updated"),
        errorToast: t("errors.update"),
      },
    );

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (scopes.length === 0) {
      setScopeError(true);
      return;
    }

    try {
      await updateServiceAccount({
        variables: {
          input: {
            serviceAccountId: serviceAccount.id,
            name: name.trim(),
            description: description.trim() || null,
            scopes,
          },
        },
      });
      onUpdated();
    } catch {
      return;
    }
  }

  return (
    <form onSubmit={event => void submit(event)}>
      <DialogHeader>
        <DialogTitle>{t("update.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className="space-y-5">
        <Field label={t("fields.name")}>
          <TextField
            name="name"
            required
            maxLength={120}
            value={name}
            onValueChange={setName}
          />
        </Field>
        <Field label={t("fields.description")}>
          <Textarea
            name="description"
            maxLength={500}
            value={description}
            onChange={event => setDescription(event.target.value)}
          />
        </Field>
        <div>
          <ScopePicker
            disabled={isUpdating}
            scopes={supportedScopes}
            selectedScopes={scopes}
            onChange={(nextScopes) => {
              setScopes(nextScopes);
              setScopeError(nextScopes.length === 0);
            }}
          />
          {scopeError && (
            <p className="mt-2 text-1 text-red-11">
              {t("validation.scopeRequired")}
            </p>
          )}
        </div>
      </DialogBody>
      <DialogFooter>
        <DialogClose render={<Button variant="soft" />}>
          {t("actions.cancel")}
        </DialogClose>
        <Button type="submit" loading={isUpdating}>
          {t("actions.save")}
        </Button>
      </DialogFooter>
    </form>
  );
}
