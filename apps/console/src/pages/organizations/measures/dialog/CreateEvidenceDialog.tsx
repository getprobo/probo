import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Dropzone,
  Field,
  Spinner,
  TabItem,
  Tabs,
  type DialogRef,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql, useRelayEnvironment } from "react-relay";
import { useState } from "react";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { updateStoreCounter } from "/hooks/useMutationWithIncrement";

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
  const { __ } = useTranslate();
  const [tab, setTab] = useState("upload");
  return (
    <Dialog
      title={
        <Breadcrumb
          items={[
            { label: __("Measures detail") },
            { label: __("Create Evidence") },
          ]}
        />
      }
      ref={props.ref}
      className="max-w-lg"
    >
      <Tabs className="px-6">
        <TabItem active={tab === "upload"} onClick={() => setTab("upload")}>
          {__("Upload")}
        </TabItem>
        <TabItem active={tab === "link"} onClick={() => setTab("link")}>
          {__("Link")}
        </TabItem>
      </Tabs>
      {tab === "upload" && <EvidenceUpload {...props} />}
      {tab === "link" && <EvidenceLink {...props} />}
    </Dialog>
  );
}

function EvidenceUpload({ measureId, connectionId }: Omit<Props, "ref">) {
  const { __ } = useTranslate();

  const relayEnv = useRelayEnvironment();
  const [mutate, isUpdating] = useMutationWithToasts(uploadEvidenceMutation, {
    successMessage: __("Evidence uploaded successfully"),
    errorMessage: __("Failed to create evidence"),
  });
  const handleDrop = (files: File[]) => {
    for (const file of files) {
      mutate({
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
          description={__(
            "Only PDF, DOCX, XLSX, PPTX, JPG, PNG, WEBP, URI files up to 10MB are allowed"
          )}
          isUploading={isUpdating}
          onDrop={handleDrop}
          accept={{
            "text/plain": [".uri"],
            "application/pdf": [".pdf"],
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
              [".docx"],
            "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
              [".xlsx"],
            "application/vnd.openxmlformats-officedocument.presentationml.presentation":
              [".pptx"],
            "image/jpeg": [".jpg", ".jpeg"],
            "image/png": [".png"],
            "image/webp": [".webp"],
          }}
          maxSize={10}
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
  const { __ } = useTranslate();
  const { handleSubmit, register, formState, reset } = useFormWithSchema(
    linkSchema,
    {
      defaultValues: {
        name: "",
        url: "",
      },
    }
  );

  const [mutate] = useMutationWithToasts(uploadEvidenceMutation, {
    successMessage: __("Evidence created successfully"),
    errorMessage: __("Failed to create evidence"),
  });
  const onSubmit = handleSubmit(async (data) => {
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
  });
  return (
    <form onSubmit={onSubmit}>
      <DialogContent padded className="space-y-4">
        <Field
          required
          type="text"
          label={__("Name")}
          placeholder={__("Evidence name")}
          {...register("name")}
          error={formState.errors.name?.message}
        />
        <Field
          required
          type="url"
          label={__("URL")}
          placeholder={__("Evidence URL")}
          {...register("url")}
          error={formState.errors.url?.message}
          help={__("This will create a .uri file with the URL inside")}
        />
      </DialogContent>
      <DialogFooter>
        <Button
          type="submit"
          disabled={formState.isSubmitting}
          icon={formState.isSubmitting ? Spinner : undefined}
        >
          {__("Create")}
        </Button>
      </DialogFooter>
    </form>
  );
}
