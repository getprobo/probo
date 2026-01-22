import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  IconArrowInbox,
  IconWarning,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useLazyLoadQuery } from "react-relay";

import type { EvidenceGraphFileQuery } from "#/__generated__/core/EvidenceGraphFileQuery.graphql";
import { evidenceFileQuery } from "#/hooks/graph/EvidenceGraph";

type Props = {
  evidenceId: string;
  filename: string;
  onClose: () => void;
};

export function EvidencePreviewDialog({
  evidenceId,
  filename,
  onClose,
}: Props) {
  const { __ } = useTranslate();
  const ref = useDialogRef();
  return (
    <Dialog
      ref={ref}
      defaultOpen
      title={
        <Breadcrumb items={[{ label: __("Evidences") }, { label: filename }]} />
      }
      onClose={onClose}
    >
      <DialogContent padded>
        <Suspense fallback={<Spinner />}>
          <EvidencePreviewContent
            evidenceId={evidenceId}
            onClose={() => ref.current?.close()}
          />
        </Suspense>
      </DialogContent>
    </Dialog>
  );
}

const fetchUrlFromUriFile = async (
  fileUrl: string,
  options?: { signal?: AbortSignal },
): Promise<string> => {
  const response = await fetch(fileUrl, options);
  const text = await response.text();
  // URI files typically have the URL on the first line
  const firstLine = text.trim().split("\n")[0];
  if (!firstLine) {
    throw new Error("No URL found in URI file");
  }
  return firstLine;
};

function EvidencePreviewContent({
  evidenceId,
  onClose,
}: Omit<Props, "filename">) {
  const evidence = useLazyLoadQuery<EvidenceGraphFileQuery>(
    evidenceFileQuery,
    { evidenceId: evidenceId },
    { fetchPolicy: "network-only" },
  ).node;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const isUriFile
    = evidence.file?.mimeType === "text/uri-list"
      || evidence.file?.mimeType === "text/uri";
  useEffect(() => {
    if (!isUriFile) {
      return;
    }
    const abortController = new AbortController();
    fetchUrlFromUriFile(evidence.file?.downloadUrl ?? "", {
      signal: abortController.signal,
    })
      .then((url) => {
        window.open(url, "_blank");
      })
      .catch((e) => {
        if (e instanceof Error) {
          if (e.name === "AbortError") {
            return;
          }
          toast({
            title: __("Error"),
            description: e.message ?? __("Failed to extract URL from URI file"),
            variant: "error",
          });
        } else {
          toast({
            title: __("Error"),
            description: __("Failed to extract URL from URI file"),
            variant: "error",
          });
        }
      })
      .finally(onClose);
    return () => {
      abortController.abort();
    };
  }, [evidence.file?.downloadUrl, isUriFile, onClose, __, toast]);

  if (!evidence.file?.downloadUrl) {
    return null;
  }

  if (isUriFile) {
    return (
      <div className="flex flex-col items-center gap-2 justify-center">
        <Spinner size={20} />
      </div>
    );
  }

  if (evidence.file.mimeType?.startsWith("image/")) {
    return (
      <img
        src={evidence.file.downloadUrl}
        alt={evidence.file.fileName}
        className="max-h-[70vh] object-contain"
      />
    );
  }

  if (evidence.file.mimeType?.includes("pdf")) {
    return (
      <iframe
        src={evidence.file.downloadUrl}
        className="w-full h-[70vh]"
        title={evidence.file.fileName}
      />
    );
  }

  return (
    <div className="flex flex-col items-center gap-2 justify-center">
      <IconWarning size={20} />
      <p className="text-txt-secondary text-center">
        {__("Preview not available for this file type")
          + " "
          + evidence.file.mimeType}
      </p>
      <Button asChild variant="secondary" icon={IconArrowInbox}>
        <a href={evidence.file.downloadUrl} target="_blank" rel="noreferrer">
          {__("Download File")}
        </a>
      </Button>
    </div>
  );
}
