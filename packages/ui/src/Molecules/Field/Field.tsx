import { tv } from "tailwind-variants";
import { Label } from "../../Atoms/Label/Label";
import { Input } from "../../Atoms/Input/Input";
import { Textarea } from "../../Atoms/Textarea/Textarea";
import { type ComponentProps, type ReactNode } from "react";
import { Select } from "../../Atoms/Select/Select";

type BaseProps<T extends string, P> = {
    label?: string;
    help?: string;
    error?: string;
    onValueChange?: (s: string) => void;
    type?: T;
    children?: ReactNode;
} & P;

type Props =
    | BaseProps<never, ComponentProps<typeof Input>>
    | BaseProps<"text", ComponentProps<typeof Input>>
    | BaseProps<"email", ComponentProps<typeof Input>>
    | BaseProps<"password", ComponentProps<typeof Input>>
    | BaseProps<"textarea", ComponentProps<typeof Textarea>>
    | BaseProps<"select", ComponentProps<typeof Select>>;

const field = tv({
    slots: {
        base: "flex flex-col",
        label: "mb-[6px]",
        help: "text-xs text-txt-tertiary mt-1",
    },
});

const { base: baseClass, label: labelClass, help: helpClass } = field();

export function Field(props: Props) {
    const showHelp = props.help && !props.error;
    const childrenAsInput = !props.type && props.children;
    return (
        <div className={baseClass()}>
            {props.label && (
                <Label htmlFor={props.name} className={labelClass()}>
                    {props.label}
                </Label>
            )}
            {childrenAsInput ? props.children : getInput(props)}
            {showHelp && <span className={helpClass()}>{props.help}</span>}
            {props.error && (
                <span className="text-txt-danger text-sm mt-1">
                    {props.error}
                </span>
            )}
        </div>
    );
}

function getInput(props: Props) {
    const { label, error, onValueChange, type, ...restProps } = props;
    const baseProps = {
        ["aria-invalid"]: !!error,
        name: props.name,
        id: props.name,
        placeholder: props.placeholder,
    };
    switch (type) {
        case "select":
            return (
                // @ts-expect-error Select is too dynamic
                <Select
                    {...baseProps}
                    {...restProps}
                    onValueChange={onValueChange}
                />
            );
        case "textarea":
            return (
                <Textarea
                    // @ts-expect-error Textarea is too dynamic
                    onChange={(e) => onValueChange?.(e.target.value)}
                    {...baseProps}
                    {...restProps}
                />
            );
        default:
            return (
                <Input
                    type={type}
                    // @ts-expect-error Input is too dynamic
                    onChange={(e) => onValueChange?.(e.target.value)}
                    {...baseProps}
                    {...restProps}
                />
            );
    }
}
