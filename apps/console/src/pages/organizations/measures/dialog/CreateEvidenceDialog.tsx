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

import {
  acceptData,
  acceptDocument,
  acceptImage,
  acceptPresentation,
  acceptSpreadsheet,
  acceptText,
} from "@probo/helpers";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  type DialogRef,
  Dropzone,
  Field,
  Spinner,
  TabItem,
  Tabs,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRelayEnvironment } from "react-relay";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const uploadEvidenceMutation = graphql`
  mutation CreateEvidenceDialogUploadMutation(
    $input: UploadMeasureEvidenceInput!
    $connections: [ID!]!
  ) {
    uploadMeasureEvidence(input: $input) {
      evidenceEdge @appendEdge(connections: $connections) {
        node {
          id
          ...MeasureEvidencesTabFragment_evidence
        }
      }
    }
  }
`;

type Props = {
  measureId: string;
  connectionId: string;
  ref: DialogRef;
};

export function CreateEvidenceDialog(props: Props) {
  const { ref, ...rest } = props;
  const { t } = useTranslation();
  const [tab, setTab] = useState("upload");
  return (
    <Dialog
      title={(
        <Breadcrumb
          items={[
            { label: t("createEvidenceDialog.breadcrumb.measureDetail") },
            { label: t("createEvidenceDialog.breadcrumb.createEvidence") },
          ]}
        />
      )}
      ref={ref}
      className="max-w-lg"
    >
      <Tabs className="px-6">
        <TabItem active={tab === "upload"} onClick={() => setTab("upload")}>
          {t("createEvidenceDialog.tabs.upload")}
        </TabItem>
        <TabItem active={tab === "link"} onClick={() => setTab("link")}>
          {t("createEvidenceDialog.tabs.link")}
        </TabItem>
      </Tabs>
      {tab === "upload" && <EvidenceUpload {...rest} />}
      {tab === "link" && <EvidenceLink ref={ref} {...rest} />}
    </Dialog>
  );
}

function EvidenceUpload({ measureId, connectionId }: Omit<Props, "ref">) {
  const { t } = useTranslation();

  const relayEnv = useRelayEnvironment();
  const [mutate, isUpdating] = useMutationWithToasts(uploadEvidenceMutation, {
    successMessage: t("createEvidenceDialog.messages.uploaded"),
    errorMessage: t("createEvidenceDialog.errors.create"),
  });
  const handleDrop = async (files: File[]) => {
    for (const file of files) {
      await mutate({
        variables: {
          connections: [connectionId],
          input: {
            measureId: measureId,
            file: null,
          },
        },
        uploadables: {
          "input.file": file,
        },
        onSuccess: () => {
          updateStoreCounter(relayEnv, measureId, "evidences(first:0)", 1);
        },
      });
    }
  };
  return (
    <>
      <DialogContent padded>
        <Dropzone
          description={t("createEvidenceDialog.uploadDescription")}
          isUploading={isUpdating}
          onDrop={files => void handleDrop(files)}
          accept={{
            ...acceptDocument,
            ...acceptSpreadsheet,
            ...acceptPresentation,
            ...acceptText,
            ...acceptData,
            ...acceptImage,
          }}
          maxSize={20}
        />
      </DialogContent>
    </>
  );
}

const linkSchema = z.object({
  name: z.string(),
  url: z.string().url(),
});

function EvidenceLink({ measureId, connectionId, ref }: Props) {
  const { t } = useTranslation();
  const { handleSubmit, register, formState, reset } = useFormWithSchema(
    linkSchema,
    {
      defaultValues: {
        name: "",
        url: "",
      },
    },
  );

  const [mutate] = useMutationWithToasts(uploadEvidenceMutation, {
    successMessage: t("createEvidenceDialog.messages.created"),
    errorMessage: t("createEvidenceDialog.errors.create"),
  });
  const onSubmit = async (data: z.infer<typeof linkSchema>) => {
    const fileName = `${data.name.trim()}.uri`;
    const file = new File([data.url.trim()], fileName, {
      type: "text/uri-list",
    });
    await mutate({
      variables: {
        connections: [connectionId],
        input: {
          measureId: measureId,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
    });
    ref.current?.close();
    reset();
  };

  return (
    <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
      <DialogContent padded className="space-y-4">
        <Field
          required
          type="text"
          label={t("createEvidenceDialog.fields.name")}
          placeholder={t("createEvidenceDialog.fields.namePlaceholder")}
          {...register("name")}
          error={formState.errors.name?.message}
        />
        <Field
          required
          type="url"
          label={t("createEvidenceDialog.fields.url")}
          placeholder={t("createEvidenceDialog.fields.urlPlaceholder")}
          {...register("url")}
          error={formState.errors.url?.message}
          help={t("createEvidenceDialog.fields.urlHelp")}
        />
      </DialogContent>
      <DialogFooter>
        <Button
          type="submit"
          disabled={formState.isSubmitting}
          icon={formState.isSubmitting ? Spinner : undefined}
        >
          {t("createEvidenceDialog.actions.create")}
        </Button>
      </DialogFooter>
    </form>
  );
}
