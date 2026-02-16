import { useTranslate } from "@probo/i18n";
import { Dialog } from "@probo/ui";
import { type PropsWithChildren } from "react";
import type { DataID } from "relay-runtime";

import { PersonForm } from "./PersonForm";

export function AddPersonDialog(props: PropsWithChildren<{
  connectionId: DataID;
}>) {
  const { children, connectionId } = props;

  const { __ } = useTranslate();

  return (
    <Dialog
      title={__("Add Person")}
      trigger={children}
      className="max-w-xl"
    >
      <div className="p-4">
        <PersonForm connectionId={connectionId} />
      </div>
    </Dialog>
  );
}
