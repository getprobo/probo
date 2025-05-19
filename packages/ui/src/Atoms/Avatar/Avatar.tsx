import { tv } from "tailwind-variants";

type Props = {
    fullName: string;
    size?: "s" | "m" | "l" | "xl";
};

const avatar = tv({
    base: "bg-txt-success text-sm text-invert rounded-full text-xxs font-semibold flex items-center justify-center",
    variants: {
        size: {
            s: "size-5",
            m: "size-6",
            l: "size-8",
            xl: "size-16",
        },
    },
    defaultVariants: {
        size: "m",
    },
});

export function Avatar({ fullName, size = "m" }: Props) {
    return <div className={avatar({ size })}>{extractInitials(fullName)}</div>;
}

function extractInitials(name: string) {
    const words = name.split(" ");
    if (words.length === 2) {
        return words
            .map((word) => word[0])
            .join("")
            .toUpperCase();
    }
    return name.substring(0, 2).toUpperCase();
}
