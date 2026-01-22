import { tv } from "tailwind-variants";

type Props = {
  name: string;
  src?: string | null;
  size?: "s" | "m" | "l" | "xl";
  className?: string;
};

const avatar = tv({
  base: "bg-txt-success text-txt-invert rounded-full font-semibold flex items-center justify-center flex-none",
  variants: {
    size: {
      s: "size-5 text-xss",
      m: "size-6 text-xxs",
      l: "size-8 text-sm",
      xl: "size-16 text-3xl",
    },
  },
  defaultVariants: {
    size: "m",
  },
});

export function Avatar(props: Props) {
  const className = avatar(props);
  if (props.src) {
    return <img className={className} src={props.src} alt={props.name} />;
  }
  return <div className={className}>{extractInitials(props.name ?? "")}</div>;
}

function extractInitials(name: string) {
  const words = name.split(" ");
  if (words.length === 2) {
    return words
      .map(word => word[0])
      .join("")
      .toUpperCase();
  }
  return name.substring(0, 2).toUpperCase();
}
