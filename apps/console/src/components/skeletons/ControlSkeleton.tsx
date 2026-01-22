import { useTranslate } from "@probo/i18n";
import { ActionDropdown, Button, IconPencil, Skeleton } from "@probo/ui";

/**
 * Skeleton state for the framework control panel
 */
export function ControlSkeleton() {
  const { __ } = useTranslate();
  return (
    <div className="space-y-6">
      <div className="flex justify-between">
        <Skeleton style={{ width: 72, height: 34 }} className="mb-3" />
        <div className="flex gap-2">
          <Button icon={IconPencil} variant="secondary" disabled>
            {__("Edit control")}
          </Button>
          <ActionDropdown variant="secondary" />
        </div>
      </div>
      <Skeleton style={{ width: "80%", height: 24 }} />
      <Skeleton style={{ height: 160 }} />
      <Skeleton style={{ height: 160 }} />
    </div>
  );
}
