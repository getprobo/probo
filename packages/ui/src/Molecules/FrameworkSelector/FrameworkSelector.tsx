import { Dropdown, DropdownItem } from "../../Atoms/Dropdown/Dropdown";
import { IconChevronDown, IconPlusLarge } from "../../Atoms/Icons";
import { Button } from "../../Atoms/Button/Button";
import { useTranslate } from "@probo/i18n";
import { ISO27001 } from "../../Atoms/Frameworks/ISO27001";
import { SOC2 } from "../../Atoms/Frameworks/SOC2";
import { HIPAA } from "../../Atoms/Frameworks/HIPAA";
import { CCPA } from "../../Atoms/Frameworks/CCPA";
import { GDPR } from "../../Atoms/Frameworks/GDPR";
import { NIS2 } from "../../Atoms/Frameworks/NIS2";
import { DORA } from "../../Atoms/Frameworks/DORA";

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
            toggle={
                <Button
                    icon={IconPlusLarge}
                    iconAfter={IconChevronDown}
                    disabled={disabled}
                >
                    {__("New framework")}
                </Button>
            }
        >
            <FrameworkItem onClick={() => onSelect("custom")} />
            {availableFrameworks.map((framework) => (
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
