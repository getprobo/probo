import { Dropdown, DropdownItem } from "../../Atoms/Dropdown/Dropdown";
import { IconChevronDown, IconPlusLarge } from "../../Atoms/Icons";
import { Button } from "../../Atoms/Button/Button";
import { availableFrameworks } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

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
            <img
                src={props.framework.logo}
                alt={props.framework.name}
                className="size-8"
            />
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
