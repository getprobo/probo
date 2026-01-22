import { downloadFile } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Breadcrumb, Dialog, DialogContent, Spinner } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useLazyLoadQuery } from "react-relay";

import type { EvidenceGraphFileQuery } from "#/__generated__/core/EvidenceGraphFileQuery.graphql";
import { evidenceFileQuery } from "#/hooks/graph/EvidenceGraph";

type Props = {
  evidenceId: string;
  onClose: () => void;
};
export function EvidenceDownloadDialog({ evidenceId, onClose }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog
      className="max-w-sm"
      onClose={onClose}
      defaultOpen
      title={(
        <Breadcrumb
          items={[{ label: __("Evidences") }, { label: __("Download") }]}
        />
      )}
    >
      <DialogContent padded>
        <Suspense
          fallback={(
            <div className="flex gap-2 justify-center">
              <Spinner />
              {__("Generating download link")}
              ...
            </div>
          )}
        >
          <DownloadLink evidenceId={evidenceId} onClose={onClose} />
        </Suspense>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Force the download of an evidence file
 */
function DownloadLink({ evidenceId, onClose }: Props) {
  const data = useLazyLoadQuery<EvidenceGraphFileQuery>(evidenceFileQuery, {
    evidenceId,
  });
  const evidence = data.node;

  useEffect(() => {
    downloadFile(
      evidence.file?.downloadUrl,
      evidence.file?.fileName ?? "evidence",
    );
    onClose();
  }, [evidence.file?.downloadUrl, evidence.file?.fileName, onClose]);

  return null;
}
