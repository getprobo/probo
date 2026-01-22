import { useCopy } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
} from "@probo/ui";

export function PersonalAPIKeyTokenDialog(props: {
  dialogRef: React.RefObject<{ open: () => void; close: () => void } | null>;
  token: string;
  onDone: () => void;
}) {
  const { dialogRef, token, onDone } = props;
  const { __ } = useTranslate();
  const [isCopied, copy] = useCopy();

  return (
    <Dialog
      ref={dialogRef}
      title={<Breadcrumb items={[__("API Keys"), __("Token")]} />}
    >
      <DialogContent padded className="space-y-4">
        <div className="bg-gray-100 p-4 rounded-lg flex items-center gap-2">
          <code className="text-sm font-mono break-all flex-1">{token}</code>
          <Button
            variant="secondary"
            onClick={() => copy(token)}
            disabled={!token}
          >
            {isCopied ? __("Copied") : __("Copy")}
          </Button>
        </div>
      </DialogContent>
      <DialogFooter>
        <Button onClick={onDone}>{__("Done")}</Button>
      </DialogFooter>
    </Dialog>
  );
}
