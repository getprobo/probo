import { clsx } from "clsx";
import type { FC } from "react";

import { IconMonitor } from "../Icons/IconMonitor";
import { IconMoon } from "../Icons/IconMoon";
import { IconSun } from "../Icons/IconSun";
import { useTheme } from "../ThemeProvider/ThemeProvider";

type Theme = "light" | "dark" | "system";

const themes: { value: Theme; icon: FC<{ size: number }> }[] = [
  { value: "light", icon: IconSun },
  { value: "dark", icon: IconMoon },
  { value: "system", icon: IconMonitor },
];

const iconMap: Record<Theme, FC<{ size: number }>> = {
  light: IconSun,
  dark: IconMoon,
  system: IconMonitor,
};

const nextMap: Record<Theme, Theme> = {
  light: "dark",
  dark: "system",
  system: "light",
};

type Props = {
  collapsed?: boolean;
};

export function ThemeToggle({ collapsed }: Props) {
  const { theme, setTheme } = useTheme();
  const CollapsedIcon = iconMap[theme];

  if (collapsed) {
    return (
      <button
        type="button"
        onClick={() => setTheme(nextMap[theme])}
        className="flex items-center justify-center size-8 rounded-full text-txt-tertiary hover:bg-subtle-hover active:bg-subtle-pressed transition-colors"
        aria-label="Toggle theme"
      >
        <CollapsedIcon size={16} />
      </button>
    );
  }

  return (
    <div className="flex items-center w-full rounded-full bg-subtle p-0.5">
      {themes.map(({ value, icon: Icon }) => (
        <button
          key={value}
          type="button"
          onClick={() => setTheme(value)}
          className={clsx(
            "flex-1 flex items-center justify-center py-1.5 rounded-full transition-all",
            theme === value
              ? "bg-level-1 text-txt-primary shadow-base"
              : "text-txt-tertiary hover:text-txt-secondary",
          )}
          aria-label={value}
        >
          <Icon size={16} />
        </button>
      ))}
    </div>
  );
}
