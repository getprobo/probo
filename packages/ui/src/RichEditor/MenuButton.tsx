import { tv } from "tailwind-variants";

const menuButtonVariants = tv({
  base: ["px-2 py-1 text-sm rounded-sm font-semibold bg-level-0 hover:bg-subtle"],
  variants: {
    active: {
      true: ["bg-active"],
    },
  },
});

type MenuButtonProps = {
  label: string;
  active?: boolean;
  onClick: () => void;
};

export function MenuButton({ label, active, onClick }: MenuButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={menuButtonVariants({ active })}
    >
      {label}
    </button>
  );
}
