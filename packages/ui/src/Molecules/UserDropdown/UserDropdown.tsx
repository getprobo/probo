import { Avatar } from "../../Atoms/Avatar/Avatar";
import { Button } from "../../Atoms/Button/Button";
import {
    Dropdown,
    DropdownItem,
    DropdownSeparator,
} from "../../Atoms/Dropdown/Dropdown";
import { IconChevronDown } from "../../Atoms/Icons";
import type { FC, PropsWithChildren } from "react";
import type { IconProps } from "../../Atoms/Icons/type";
import { Link } from "react-router";

type Props = PropsWithChildren<{ fullName: string; email: string }>;

export function UserDropdown({ fullName, children, email }: Props) {
    return (
        <Dropdown
            className="w-60"
            toggle={
                <Button variant="tertiary">
                    <Avatar name={fullName} />
                    <span>{fullName}</span>
                    <IconChevronDown size={16} />
                </Button>
            }
        >
            <div className="flex gap-2 items-center">
                <Avatar name={fullName} size="l" />
                <div>
                    <p className="text-sm font-medium text-txt-primary">
                        {fullName}
                    </p>
                    <p className="text-xxs text-txt-tertiary">{email}</p>
                </div>
            </div>
            <DropdownSeparator />
            {children}
        </Dropdown>
    );
}

export function UserDropdownItem({
    to,
    icon: IconComponent,
    label,
    variant = "tertiary",
}: {
    to: string;
    icon: FC<IconProps>;
    label: string;
    variant?: "tertiary" | "danger";
}) {
    return (
        <DropdownItem asChild variant={variant}>
            <Link to={to}>
                {IconComponent && <IconComponent size={16} />}
                {label}
            </Link>
        </DropdownItem>
    );
}
