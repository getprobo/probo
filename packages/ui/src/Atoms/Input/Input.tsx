import { tv } from "tailwind-variants";

type Props = {
    label?: string;
    help?: string;
    name?: string;
    placeholder?: string;
    type?: string;
    required?: boolean;
};

const input = tv({
    slots: {
        base: "flex flex-col",
        label: "text-sm font-medium text-txt-primary mb-[6px]",
        input: "py-[6px] bg-secondary border border-border-mid rounded-[10px] hover:border-border-strong focus:shadow-focus text-sm px-3",
        help: "text-xs font-semibold text-txt-tertiary mt-1",
    },
});

const {
    base: baseClass,
    label: labelClass,
    input: inputClass,
    help: helpClass,
} = input();

export function Input({
    label,
    help,
    name,
    placeholder,
    type,
    required,
}: Props) {
    return (
        <div className={baseClass()}>
            {label && (
                <label htmlFor={name} className={labelClass()}>
                    {label}
                </label>
            )}
            <input
                name={name}
                id={name}
                className={inputClass()}
                placeholder={placeholder}
                type={type}
                required={required}
            />
            {help && <span className={helpClass()}>{help}</span>}
        </div>
    );
}
