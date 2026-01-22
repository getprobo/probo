import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, PageHeader, Skeleton } from "@probo/ui";

export function RisksPageSkeleton() {
  const { __ } = useTranslate();
  return (
    <div className="space-y-6">
      <PageHeader title={__("Risks")}>
        <Button icon={IconPlusLarge} disabled>
          {__("New Risk")}
        </Button>
      </PageHeader>
      <div className="grid grid-cols-2 gap-4">
        <Skeleton className="aspect-square" />
        <Skeleton className="aspect-square" />
      </div>
      <div>
        <Skeleton style={{ aspectRatio: "1280/280" }} />
      </div>
    </div>
  );
}
