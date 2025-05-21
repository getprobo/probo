import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Label,
  Select,
  Option,
  Field,
} from "@probo/ui";
import { useMemo, useState, type ReactNode } from "react";
import { useFetchQuery } from "../../../hooks/useFetchQuery";

type Props = {
  trigger: ReactNode;
};

type Risk = {
  category: string;
  name: string;
  description: string;
};

export default function NewRiskDialog({ trigger }: Props) {
  const { __ } = useTranslate();

  const onTemplateChange = (risk: Risk) => {
    console.log(risk);
  };

  return (
    <Dialog
      open
      onClose={() => {}}
      trigger={trigger}
      title={<Breadcrumb items={[__("Risks"), __("New Risk")]}></Breadcrumb>}
    >
      <form>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          {/* Main form */}
          <div className="py-8 px-12 space-y-6">
            <TemplateSelector onChange={onTemplateChange} />
            <Field
              label={__("Risk name")}
              name="name"
              placeholder={__("Service Outage")}
            />
            <Field
              label={__("Description")}
              name="description"
              placeholder={__("Type your description here")}
              type="textarea"
            />
          </div>

          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle"></div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">{__("Create risk")}</Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function TemplateSelector({ onChange }: { onChange: (risk: Risk) => void }) {
  const { __ } = useTranslate();
  const { data: risks } = useFetchQuery<Risk[]>("/data/risks/risks.json", {
    staleTime: 100_000,
  });

  const categories = useMemo(
    () => Array.from(new Set(risks?.map((t) => t.category))),
    [risks]
  );

  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  const templates = useMemo(
    () => risks?.filter((r) => r.category === selectedCategory) ?? [],
    [risks, selectedCategory]
  );

  const onTemplateChange = (template: string) => {
    const risk = risks?.find((r) => r.name === template);
    if (!risk) {
      throw new Error("Risk not found");
    }
    onChange(risk);
  };
  return (
    <div>
      <Label>{__("Risk category")}</Label>
      <div className="grid grid-cols-2 gap-2 mt-[6px]">
        <Select
          placeholder={__("Select a category")}
          onChange={setSelectedCategory}
        >
          {categories.map((category) => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
        <Select
          empty={templates?.length === 0}
          placeholder={__("Select template")}
          onChange={onTemplateChange}
        >
          {templates?.map((template) => (
            <Option key={template.name} value={template.name}>
              {template.name}
            </Option>
          ))}
        </Select>
      </div>
    </div>
  );
}
