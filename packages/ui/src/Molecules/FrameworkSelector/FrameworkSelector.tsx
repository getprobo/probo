import { useTranslate } from "@probo/i18n";

import { Button } from "../../Atoms/Button/Button";
import { Dropdown, DropdownItem } from "../../Atoms/Dropdown/Dropdown";
import { TwentyOneCFRPart11 } from "../../Atoms/Frameworks/21CFRPart11";
import { CCPA } from "../../Atoms/Frameworks/CCPA";
import { DORA } from "../../Atoms/Frameworks/DORA";
import { GDPR } from "../../Atoms/Frameworks/GDPR";
import { HDS } from "../../Atoms/Frameworks/HDS";
import { HIPAA } from "../../Atoms/Frameworks/HIPAA";
import { ISO27001 } from "../../Atoms/Frameworks/ISO27001";
import { ISO27701 } from "../../Atoms/Frameworks/ISO27701";
import { ISO42001 } from "../../Atoms/Frameworks/ISO42001";
import { NIS2 } from "../../Atoms/Frameworks/NIS2";
import { SOC2 } from "../../Atoms/Frameworks/SOC2";
import { IconChevronDown, IconPlusLarge } from "../../Atoms/Icons";

const availableFrameworks = [
  {
    id: "ISO27001-2022",
    name: "ISO 27001 (2022)",
    logo: <ISO27001 className="size-8" />,
    description: "Information security management systems",
  },
  {
    id: "SOC2",
    name: "SOC 2",
    logo: <SOC2 className="size-8" />,
    description: "System and Organization Controls 2",
  },
  {
    id: "HIPAA",
    name: "HIPAA",
    logo: <HIPAA className="size-8" />,
    description: "Health Insurance Portability and Accountability Act",
  },
  {
    id: "CCPA",
    name: "CCPA",
    logo: <CCPA className="size-8" />,
    description: "California Consumer Privacy Act",
  },
  {
    id: "NIS2",
    name: "NIS 2",
    logo: <NIS2 className="size-8" />,
    description: "Network and Information Systems Directive 2",
  },
  {
    id: "GDPR",
    name: "GDPR",
    logo: <GDPR className="size-8" />,
    description: "General Data Protection Regulation",
  },
  {
    id: "DORA",
    name: "DORA",
    logo: <DORA className="size-8" />,
    description: "Digital Operational Readiness Assessment",
  },
  {
    id: "ISO27701-2025",
    name: "ISO 27701 (2025)",
    logo: <ISO27701 className="size-8" />,
    description:
            "Information security, cybersecurity and privacy protection",
  },
  {
    id: "ISO42001-2023",
    name: "ISO 42001 (2023)",
    logo: <ISO42001 className="size-8" />,
    description:
            "Information technology, artificial intelligence, management system",
  },
  {
    id: "21CFR-part11",
    name: "21 CFR Part 11",
    logo: <TwentyOneCFRPart11 className="size-8" />,
    description: "21 CFR Part 11",
  },
  {
    id: "HDS",
    name: "HDS",
    logo: <HDS className="size-8" />,
    description: "Hébergement de Données de Santé",
  },
];

type Framework = (typeof availableFrameworks)[number];

type Props = {
  disabled?: boolean;
  onSelect: (frameworkId: string) => void;
};

export function FrameworkSelector({ disabled, onSelect }: Props) {
  const { __ } = useTranslate();
  return (
    <Dropdown
      toggle={(
        <Button
          icon={IconPlusLarge}
          iconAfter={IconChevronDown}
          disabled={disabled}
        >
          {__("New framework")}
        </Button>
      )}
    >
      <FrameworkItem onClick={() => onSelect("custom")} />
      {availableFrameworks.map(framework => (
        <FrameworkItem
          key={framework.id}
          framework={framework}
          onClick={() => onSelect(framework.id)}
        />
      ))}
    </Dropdown>
  );
}

function FrameworkItem(props: { framework?: Framework; onClick: () => void }) {
  const { __ } = useTranslate();
  if (!props.framework) {
    return (
      <DropdownItem onClick={props.onClick} className="">
        <div className="rounded-full size-8 bg-highlight text-txt-primary flex items-center justify-center">
          <IconPlusLarge size={16} />
        </div>
        <div className="space-y-[2px]">
          <div className="text-sm font-medium">
            {__("Custom framework")}
          </div>
          <div className="text-xs text-txt-secondary">
            {__("Start from scratch")}
          </div>
        </div>
      </DropdownItem>
    );
  }
  return (
    <DropdownItem onClick={props.onClick} className="">
      {props.framework.logo}
      <div className="space-y-[2px]">
        <div className="text-sm font-medium">
          {props.framework.name}
        </div>
        <div className="text-xs text-txt-secondary">
          {props.framework.description}
        </div>
      </div>
    </DropdownItem>
  );
}
