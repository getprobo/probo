import { focusSiblingElement } from "@probo/helpers";
import * as Popover from "@radix-ui/react-popover";
import {
  type CSSProperties,
  type KeyboardEventHandler,
  type ReactNode,
  useEffect,
  useRef,
  useState,
} from "react";

import { Cell } from "../../Atoms/DataTable/DataTable";

import { useEditableRowContext } from "./EditableRow";

export function useEditableCellRef() {
  return useRef<{ close: () => void } | null>(null);
}

/**
 * Base component to create an editable table cell
 */
export function EditableCell(props: {
  // Name of the field (used to retrieve errors)
  name?: string;
  // Label displayed inside the cell
  label: ReactNode;
  // Callback when the edit is dismissed
  onClose: () => void;
  // Content of the popover (e.g. a field)
  children: ReactNode;
  // Ref used to control the popover (used to close it programatically)
  ref: ReturnType<typeof useEditableCellRef>;
}) {
  const { children, onClose, label, name, ref } = props;
  const { errors } = useEditableRowContext();
  const [isOpen, setOpen] = useState(false);
  const [height, setHeight] = useState<number | undefined>(undefined);
  const [padding, setPadding] = useState("12px");
  const td = useRef<HTMLTableCellElement>(null);

  // When opening the popover, remember the height and padding of the cell
  const onOpenChange = (open: boolean) => {
    if (open) {
      setHeight(td.current?.offsetHeight ?? undefined);
      setPadding(getComputedStyle(td.current!).paddingLeft);
    } else {
      onClose();
      setHeight(undefined);
    }
    setOpen(open);
  };

  useEffect(() => {
    if (ref) {
      ref.current = {
        close: () => onOpenChange(false),
      };
    }
  });

  // Handle keyboard navigation inside the cells
  const onKeyDown: KeyboardEventHandler<HTMLButtonElement> = (e) => {
    if (e.key === "ArrowRight") {
      focusSiblingElement(1);
    } else if (e.key === "ArrowLeft") {
      focusSiblingElement(-1);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      focusSiblingElement(td.current?.parentNode?.children.length ?? 0);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      focusSiblingElement(
        (td.current?.parentNode?.children.length ?? 0) * -1,
      );
    }
  };
  const hasError = errors && name && name in errors;

  return (
    <Popover.Root onOpenChange={onOpenChange} open={isOpen}>
      <Popover.Trigger asChild>
        <Cell ref={td} asChild>
          {/* Keep the height of the cell when the popover is open, so that the popover doesn't jump when it opens/closes. */}
          <button
            onKeyDown={onKeyDown}
            className="flex flex-row justify-start hover:bg-level-2 flex-wrap gap-1 items-center relative"
            style={{ height }}
          >
            {label}
            {hasError && (
              <div className="size-2 bg-txt-danger rounded-full top-1/2 right-3 absolute -translate-y-1/2 animate-pulse" />
            )}
          </button>
        </Cell>
      </Popover.Trigger>

      <Popover.Portal>
        <Popover.Content
          side="bottom"
          align="start"
          sideOffset={height ? height * -1 : 0}
          style={
            {
              "minHeight": height,
              "--padding": padding,
              "--height": height + "px",
            } as CSSProperties
          }
          className="border border-border-low bg-level-2 min-w-[200px] flex flex-col justify-center rounded-sm"
        >
          {children}
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
