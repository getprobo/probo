import { IconCircleProgress } from "./IconCircleProgress";
import { IconRadioUnchecked } from "./IconRadioUnchecked";

type Props = {
  state: "TODO" | "DONE";
};

export function TaskStateIcon({ state }: Props) {
  return state === "TODO"
    ? (
        <IconRadioUnchecked size={16} className="text-txt-quaternary" />
      )
    : (
        <IconCircleProgress size={16} className="text-txt-accent" />
      );
}
