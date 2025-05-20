import { Button, PageHeader } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { IconPlusLarge } from "@probo/ui";
import NewRiskDialog from "./NewRiskDialog";

export default function RisksPage() {
  const { __ } = useTranslate();
  return (
    <div className="space-y-6">
      <PageHeader title={__("Risks")}>
        <NewRiskDialog
          trigger={<Button icon={IconPlusLarge}>{__("New Risk")}</Button>}
        />
      </PageHeader>
    </div>
  );
}
