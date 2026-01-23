import { clsx } from "clsx";

type Props = {
  level: number;
};

export function PriorityLevel({ level }: Props) {
  return (
    <div className="w-max p-[2px] flex gap-[2px] items-end">
      <div
        className={clsx(
          "h-1 w-[3px] bg-txt-quaternary rounded",
          level >= 1 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
      <div
        className={clsx(
          "h-2 w-[3px] bg-txt-quaternary rounded",
          level >= 2 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
      <div
        className={clsx(
          "h-3 w-[3px] bg-txt-quaternary rounded",
          level >= 3 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
    </div>
  );
}
