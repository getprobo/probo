import {
  autoUpdate,
  flip,
  offset,
  shift,
  useClick,
  useDismiss,
  useFloating,
  useFloatingRootContext,
  useInteractions,
} from "@floating-ui/react";
import { CodeBlockIcon, ListBulletsIcon, ListNumbersIcon, MinusIcon, PlusIcon, QuotesIcon, TextHOneIcon, TextHThreeIcon, TextHTwoIcon, TextTIcon } from "@phosphor-icons/react";
import { type useEditor } from "@tiptap/react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { MenuButton } from "./MenuButton";

const blockMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex size-6 items-center justify-center",
      "rounded text-txt-tertiary hover:bg-subtle hover:text-txt-primary text-xl font-light cursor-pointer",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});

const { trigger, menu } = blockMenuVariants();

function findClosestRootBlock(element: Element, editorDom: Element): HTMLElement | null {
  let current: Element | null = element;

  while (current?.parentElement && current.parentElement !== editorDom) {
    current = current.parentElement;
  }

  return current?.parentElement === editorDom ? (current as HTMLElement) : null;
}

type BlockMenuProps = {
  editor: ReturnType<typeof useEditor>;
};

export function BlockMenu({ editor }: BlockMenuProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [triggerEl, setTriggerEl] = useState<Element | null>(null);
  const [dropdownEl, setDropdownEl] = useState<HTMLElement | null>(null);
  const [hoveredBlock, setHoveredBlock] = useState<HTMLElement | null>(null);
  const rafId = useRef<number | null>(null);

  const {
    refs: triggerRefs,
    floatingStyles: triggerStyles,
    isPositioned,
  } = useFloating({
    strategy: "fixed",
    placement: "left-start",
    middleware: [offset(32)],
    whileElementsMounted: autoUpdate,
  });

  const menuRootContext = useFloatingRootContext({
    open: menuOpen,
    onOpenChange: setMenuOpen,
    elements: { reference: triggerEl, floating: dropdownEl },
  });

  const { refs: menuRefs, floatingStyles: menuStyles } = useFloating({
    rootContext: menuRootContext,
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [offset(4), flip(), shift()],
    whileElementsMounted: autoUpdate,
  });

  const click = useClick(menuRootContext);
  const dismiss = useDismiss(menuRootContext);
  const { getReferenceProps, getFloatingProps } = useInteractions([click, dismiss]);

  useEffect(() => {
    if (!editor || editor.isDestroyed) return;
    const editorDom = editor.view.dom;

    const onMouseMove = (e: MouseEvent) => {
      if (menuOpen) return;

      if (rafId.current) return;
      rafId.current = requestAnimationFrame(() => {
        rafId.current = null;

        if (!editor.isEditable) {
          setHoveredBlock(null);
          return;
        }

        const elements = editorDom.ownerDocument.elementsFromPoint(e.clientX, e.clientY);
        let block: HTMLElement | null = null;

        for (const el of elements) {
          if (!editorDom.contains(el)) continue;
          block = findClosestRootBlock(el, editorDom);
          if (block) break;
        }

        if (block) {
          setHoveredBlock(block);
        }
      });
    };

    editorDom.addEventListener("mousemove", onMouseMove);

    return () => {
      editorDom.removeEventListener("mousemove", onMouseMove);
      if (rafId.current) {
        cancelAnimationFrame(rafId.current);
        rafId.current = null;
      }
    };
  }, [editor, menuOpen]);

  useLayoutEffect(() => {
    triggerRefs.setReference(hoveredBlock);
  }, [hoveredBlock, triggerRefs]);

  const shouldShow = hoveredBlock != null || menuOpen;

  if (!editor || !shouldShow) return null;

  const handleAction = (applyCommand: (chain: ReturnType<typeof editor.chain>) => ReturnType<typeof editor.chain>) => {
    if (!hoveredBlock) return;

    try {
      const pos = editor.view.posAtDOM(hoveredBlock, 0);
      const $pos = editor.state.doc.resolve(pos);

      let insertPos: number;
      if ($pos.depth >= 1) {
        const rootPos = $pos.before(1);
        const rootNode = $pos.node(1);
        insertPos = rootPos + rootNode.nodeSize;
      } else {
        const nodeAfter = $pos.nodeAfter;
        insertPos = pos + (nodeAfter?.nodeSize ?? 1);
      }

      editor.chain()
        .focus()
        .insertContentAt(insertPos, { type: "paragraph" })
        .setTextSelection(insertPos + 1)
        .run();

      applyCommand(editor.chain()).run();
    } catch {
      // Block may no longer be in the document
    }

    setMenuOpen(false);
  };

  return (
    <>
      <button
        ref={(node) => {
          triggerRefs.setFloating(node);
          setTriggerEl(node);
          menuRefs.setReference(node);
        }}
        {...getReferenceProps()}
        onMouseDown={e => e.preventDefault()}
        type="button"
        style={{
          ...triggerStyles,
          visibility: isPositioned ? "visible" : "hidden",
        }}
        className={trigger()}
      >
        <PlusIcon size={16} weight="bold" />
      </button>
      {menuOpen && (
        <div
          ref={(node) => {
            setDropdownEl(node);
            menuRefs.setFloating(node);
          }}
          style={menuStyles}
          {...getFloatingProps()}
          onMouseDown={e => e.preventDefault()}
          className={menu()}
        >
          <div className="p-1 font-semibold text-sm">Style</div>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.setParagraph())}
          >
            <TextTIcon size={16} weight="bold" />
            Text
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 1 }))}
          >
            <TextHOneIcon size={16} weight="bold" />
            Heading 1
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 2 }))}
          >
            <TextHTwoIcon size={16} weight="bold" />
            Heading 2
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 3 }))}
          >
            <TextHThreeIcon size={16} weight="bold" />
            Heading 3
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleBulletList())}
          >
            <ListBulletsIcon size={16} weight="bold" />
            Bullet List
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleOrderedList())}
          >
            <ListNumbersIcon size={16} weight="bold" />
            Ordered List
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleCodeBlock())}
          >
            <CodeBlockIcon size={16} weight="bold" />
            Code Block
          </MenuButton>
          <MenuButton
            active={false}
            onClick={() => handleAction(chain => chain.toggleBlockquote())}
          >
            <QuotesIcon size={16} weight="bold" />
            Blockquote
          </MenuButton>
          <MenuButton
            onClick={() => handleAction(chain => chain.setHorizontalRule())}
          >
            <MinusIcon size={16} weight="bold" />
            Divider
          </MenuButton>
        </div>
      )}
    </>
  );
}
