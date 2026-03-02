import { useTranslate } from "@probo/i18n";
import { Dialog, useDialogRef } from "@probo/ui";
import { type PropsWithChildren } from "react";
import type { DataID } from "relay-runtime";

import { PersonForm } from "./PersonForm";

export function AddPersonDialog(props: PropsWithChildren<{
  connectionId: DataID;
}>) {
  const { children, connectionId } = props;
  const dialogRef = useDialogRef();
  const { __ } = useTranslate();

  return (
    <Dialog
      title={__("Add Person")}
      trigger={children}
      className="max-w-xl"
      ref={dialogRef}
    >
      <div className="p-4">
        <PersonForm connectionId={connectionId} onSubmit={() => dialogRef.current?.close()} />
      </div>
    </Dialog>
  );
}
