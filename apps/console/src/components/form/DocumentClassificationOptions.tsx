import { useTranslate } from "@probo/i18n";
import { Option } from "@probo/ui";
import {
  documentClassifications,
  getDocumentClassificationLabel,
} from "@probo/helpers";

export function DocumentClassificationOptions() {
  const { __ } = useTranslate();

  return (
    <>
      {documentClassifications.map((classification) => (
        <Option key={classification} value={classification}>
          {getDocumentClassificationLabel(__, classification)}
        </Option>
      ))}
    </>
  );
}
