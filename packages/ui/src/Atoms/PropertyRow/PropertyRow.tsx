import { Children, type ReactNode } from "react";
import { Label } from "../Label/Label";

type Props = {
    label: string;
    id?: string;
    children: ReactNode;
    error?: string;
};

export function PropertyRow({ id, label, children, error, ...props }: Props) {
    const [firstChild, ...restChildren] = Children.toArray(children);

    return (
        <div className="py-3 border-b border-border-low space-y-2" {...props}>
            <div className="flex items-center justify-between">
                <Label
                    className="text-sm text-txt-secondary font-medium mb-0"
                    htmlFor={id}
                >
                    {label}
                </Label>
                <div>{firstChild}</div>
            </div>
            {error && <div className="text-xs text-txt-danger">{error}</div>}
            {restChildren}
        </div>
    );
}
