// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { detectSocialName } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
import { ConnectionHandler, graphql, readInlineData } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageCustomLinkDialog_createMutation } from "#/__generated__/core/CompliancePageCustomLinkDialog_createMutation.graphql";
import type { CompliancePageCustomLinkDialog_customLink$key } from "#/__generated__/core/CompliancePageCustomLinkDialog_customLink.graphql";
import type { CompliancePageCustomLinkDialog_updateMutation } from "#/__generated__/core/CompliancePageCustomLinkDialog_updateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutation } from "#/lib/relay/useMutation";

const customLinkFragment = graphql`
  fragment CompliancePageCustomLinkDialog_customLink on ComplianceCustomLink @inline {
    id
    name
    url
  }
`;

const createMutation = graphql`
  mutation CompliancePageCustomLinkDialog_createMutation($input: CreateComplianceCustomLinkInput!) {
    createComplianceCustomLink(input: $input) {
      complianceCustomLinkEdge {
        node {
          id
          name
          url
          rank
          ...CompliancePageCustomLinkListItem_customLink
          ...CompliancePageCustomLinkDialog_customLink
        }
      }
    }
  }
`;

const updateMutation = graphql`
  mutation CompliancePageCustomLinkDialog_updateMutation($input: UpdateComplianceCustomLinkInput!) {
    updateComplianceCustomLink(input: $input) {
      complianceCustomLink {
        id
        name
        url
        rank
        ...CompliancePageCustomLinkListItem_customLink
        ...CompliancePageCustomLinkDialog_customLink
      }
    }
  }
`;

export interface CompliancePageCustomLinkDialogRef {
  openCreate: (compliancePageId: string, connectionId: string) => void;
  openEdit: (customLinkKey: CompliancePageCustomLinkDialog_customLink$key) => void;
}

export const CompliancePageCustomLinkDialog = forwardRef<CompliancePageCustomLinkDialogRef>(
  function CompliancePageCustomLinkDialog(_, ref) {
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();
    const [mode, setMode] = useState<"create" | "edit">("create");
    const [compliancePageId, setCompliancePageId] = useState("");
    const [connectionId, setConnectionId] = useState("");
    const [editId, setEditId] = useState<string | null>(null);

    const schema = z.object({
      name: z.string().min(1, __("Name is required")),
      url: z.string().url(__("Please enter a valid URL")),
    });

    const [create, isCreating] = useMutation<CompliancePageCustomLinkDialog_createMutation>(
      createMutation,
      { successMessage: __("Link added successfully."), errorToast: __("Failed to add link.") },
    );

    const [update, isUpdating] = useMutation<CompliancePageCustomLinkDialog_updateMutation>(
      updateMutation,
      { successMessage: __("Link updated successfully."), errorToast: __("Failed to update link.") },
    );

    const { register, handleSubmit, formState: { errors }, reset, setValue, watch } = useFormWithSchema(schema, {
      defaultValues: { name: "", url: "" },
    });

    const [nameAutoDetected, setNameAutoDetected] = useState(false);

    const handleUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const url = e.target.value;
      const detected = detectSocialName(url);
      if (detected && (nameAutoDetected || watch("name") === "")) {
        setValue("name", detected, { shouldValidate: true });
        setNameAutoDetected(true);
      } else if (!detected && nameAutoDetected) {
        setValue("name", "", { shouldValidate: false });
        setNameAutoDetected(false);
      }
    };

    useImperativeHandle(ref, () => ({
      openCreate: (pageId, cId) => {
        setMode("create");
        setCompliancePageId(pageId);
        setConnectionId(cId);
        setEditId(null);
        setNameAutoDetected(false);
        reset({ name: "", url: "" });
        dialogRef.current?.open();
      },
      openEdit: (customLinkKey) => {
        const customLink = readInlineData(customLinkFragment, customLinkKey);
        setMode("edit");
        setEditId(customLink.id);
        setNameAutoDetected(false);
        reset({ name: customLink.name, url: customLink.url });
        dialogRef.current?.open();
      },
    }));

    const onSubmit = async (data: z.infer<typeof schema>) => {
      if (mode === "create") {
        await create({
          variables: {
            input: { trustCenterId: compliancePageId, name: data.name, url: data.url },
          },
          updater: (store) => {
            const payload = store.getRootField("createComplianceCustomLink");
            const edge = payload?.getLinkedRecord("complianceCustomLinkEdge");
            if (!edge) return;
            const connection = store.get(connectionId);
            if (!connection) return;
            ConnectionHandler.insertEdgeAfter(connection, edge);
          },
        });
      } else if (editId) {
        await update({
          variables: { input: { id: editId, name: data.name, url: data.url } },
        });
      }

      reset();
      dialogRef.current?.close();
    };

    const isSubmitting = isCreating || isUpdating;
    const title = mode === "create" ? __("Add link") : __("Edit link");

    return (
      <Dialog ref={dialogRef} title={title} onClose={() => reset()}>
        <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
          <DialogContent padded className="space-y-6">
            <Field
              {...register("url", { onChange: handleUrlChange })}
              label={__("URL")}
              type="url"
              required
              placeholder="https://example.com"
              error={errors.url?.message}
            />
            <Field
              {...register("name")}
              label={__("Name")}
              type="text"
              required
              placeholder={__("e.g. Twitter, LinkedIn")}
              error={errors.name?.message}
            />
          </DialogContent>
          <DialogFooter>
            <Button type="submit" disabled={isSubmitting} icon={isSubmitting ? Spinner : undefined}>
              {mode === "create" ? __("Add link") : __("Save changes")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    );
  },
);
