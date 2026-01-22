import { getSnapshotTypeLabel, snapshotTypes } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Option } from "@probo/ui";

export function SnapshotTypeOptions() {
  const { __ } = useTranslate();

  return (
    <>
      {snapshotTypes.map(type => (
        <Option key={type} value={type}>
          {getSnapshotTypeLabel(__, type)}
        </Option>
      ))}
    </>
  );
}
