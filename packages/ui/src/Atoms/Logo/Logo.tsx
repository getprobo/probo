import clsx from "clsx";

import { useTheme } from "../ThemeProvider/ThemeProvider";

type Props = {
  className?: string;
  withPicto?: boolean;
};

export function Logo({ className, withPicto }: Props) {
  const { resolvedTheme } = useTheme();
  const src = resolvedTheme === "dark" ? "/logos/govrly-white.png" : "/logos/govrly.png";
  const alt = "Govrly";

  if (withPicto) {
    return (
      <img
        src={src}
        alt={alt}
        className={clsx(className, "aspect-auto h-auto")}
      />
    );
  }

  return <img src={src} alt={alt} className={clsx(className, "h-auto")} />;
}
