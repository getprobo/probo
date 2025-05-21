import { tv } from "tailwind-variants";
import { Label } from "../../Atoms/Label/Label";
import { Input } from "../../Atoms/Input/Input";
import { Textarea } from "../../Atoms/Textarea/Textarea";

type Props = {
    label?: string;
    help?: string;
    name?: string;
    placeholder?: string;
    type?: string;
    required?: boolean;
};

const field = tv({
    slots: {
        base: "flex flex-col",
        label: "mb-[6px]",
        help: "text-xs font-semibold text-txt-tertiary mt-1",
    },
});

const { base: baseClass, label: labelClass, help: helpClass } = field();

export function Field({
    label,
    help,
    name,
    placeholder,
    type,
    required,
}: Props) {
    const Comp = type === "textarea" ? Textarea : Input;
    return (
        <div className={baseClass()}>
            {label && (
                <Label htmlFor={name} className={labelClass()}>
                    {label}
                </Label>
            )}
            <Comp
                name={name}
                id={name}
                placeholder={placeholder}
                type={type}
                required={required}
            />
            {help && <span className={helpClass()}>{help}</span>}
        </div>
    );
}
