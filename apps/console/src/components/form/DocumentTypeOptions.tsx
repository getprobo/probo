import { documentTypes, getDocumentTypeLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Option } from "@probo/ui";

export function DocumentTypeOptions() {
  const { __ } = useTranslate();

  return (
    <>
      {documentTypes.map(type => (
        <Option key={type} value={type}>
          {getDocumentTypeLabel(__, type)}
        </Option>
      ))}
    </>
  );
}
