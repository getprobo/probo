import {
  documentClassifications,
  getDocumentClassificationLabel,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Option } from "@probo/ui";

export function DocumentClassificationOptions() {
  const { __ } = useTranslate();

  return (
    <>
      {documentClassifications.map(classification => (
        <Option key={classification} value={classification}>
          {getDocumentClassificationLabel(__, classification)}
        </Option>
      ))}
    </>
  );
}
