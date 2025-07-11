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
import { useTranslate } from "@probo/i18n";
import { Suspense, useEffect } from "react";
import { useLazyLoadQuery } from "react-relay";
import { evidenceFileQuery } from "/hooks/graph/EvidenceGraph.ts";
import type { EvidenceGraphFileQuery } from "/hooks/graph/__generated__/EvidenceGraphFileQuery.graphql.ts";

type Props = {
  evidenceId: string;
  filename: string;
  onClose: () => void;
};

export function EvidencePreviewDialog({
  evidenceId,
  onClose,
  filename,
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
  const isUriFile =
    evidence.mimeType === "text/uri-list" || evidence.mimeType === "text/uri";
  useEffect(() => {
    if (!isUriFile) {
      return;
    }
    const abortController = new AbortController();
    fetchUrlFromUriFile(evidence.fileUrl ?? "", {
      signal: abortController.signal,
    })
      .then((url) => {
        window.open(url, "_blank");
      })
      .catch((e) => {
        if (e.name === "AbortError") {
          return;
        }
        toast({
          title: __("Error"),
          description: e.message ?? __("Failed to extract URL from URI file"),
          variant: "error",
        });
      })
      .finally(onClose);
    return () => {
      abortController.abort();
    };
  }, [evidence.fileUrl, isUriFile]);

  if (!evidence.fileUrl) {
    return null;
  }

  if (isUriFile) {
    return (
      <div className="flex flex-col items-center gap-2 justify-center">
        <Spinner size={20} />
      </div>
    );
  }

  if (evidence.mimeType?.startsWith("image/")) {
    return (
      <img
        src={evidence.fileUrl}
        alt={evidence.filename}
        className="max-h-[70vh] object-contain"
      />
    );
  }

  if (evidence.mimeType?.includes("pdf")) {
    return (
      <iframe
        src={evidence.fileUrl}
        className="w-full h-[70vh]"
        title={evidence.filename}
      />
    );
  }

  return (
    <div className="flex flex-col items-center gap-2 justify-center">
      <IconWarning size={20} />
      <p className="text-txt-secondary text-center">
        {__("Preview not available for this file type") +
          " " +
          evidence.mimeType}
      </p>
      <Button asChild variant="secondary" icon={IconArrowInbox}>
        <a href={evidence.fileUrl} target="_blank">
          {__("Download File")}
        </a>
      </Button>
    </div>
  );
}
