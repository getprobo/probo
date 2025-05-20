import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
} from "@probo/ui";
import type { ReactNode } from "react";

type Props = {
  trigger: ReactNode;
};

export default function NewRiskDialog({ trigger }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog
      onClose={() => {}}
      trigger={trigger}
      title={<Breadcrumb items={[__("Risks"), __("New Risk")]}></Breadcrumb>}
    >
      <form>
        <DialogContent>
          Lorem ipsum dolor, sit amet consectetur adipisicing elit. Velit harum
          maiores dolor hic nisi repudiandae autem quod rem eveniet iure
          quibusdam sint iste numquam, suscipit ad delectus aspernatur. Quo,
          amet.
        </DialogContent>
        <DialogFooter>
          <Button type="submit">{__("Create risk")}</Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
