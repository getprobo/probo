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
import { type useEditor } from "@tiptap/react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { IconPlusSmall } from "../Atoms/Icons";

import { MenuButton } from "./MenuButton";

const blockMenuVariants = tv({
  slots: {
    trigger: [
      "z-50 flex size-6 items-center justify-center",
      "rounded text-txt-tertiary hover:bg-subtle hover:text-txt-primary text-xl font-light cursor-pointer",
    ],
    menu: ["flex items-center gap-1 rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-50"],
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
      const rootPos = $pos.before(1);
      const rootNode = $pos.node(1);
      const insertPos = rootPos + rootNode.nodeSize;

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
        <IconPlusSmall size={16} />
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
          <MenuButton
            label="H1"
            active={editor.isActive("heading", { level: 1 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 1 }))}
          />
          <MenuButton
            label="H2"
            active={editor.isActive("heading", { level: 2 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 2 }))}
          />
          <MenuButton
            label="H3"
            active={editor.isActive("heading", { level: 3 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 3 }))}
          />
          <MenuButton
            label="Bullet List"
            active={editor.isActive("bulletList")}
            onClick={() => handleAction(chain => chain.toggleBulletList())}
          />
          <MenuButton
            label="Ordered List"
            active={editor.isActive("orderedList")}
            onClick={() => handleAction(chain => chain.toggleOrderedList())}
          />
          <MenuButton
            label="Code"
            active={editor.isActive("code")}
            onClick={() => handleAction(chain => chain.toggleCode())}
          />
          <MenuButton
            label="Code Block"
            active={editor.isActive("codeBlock")}
            onClick={() => handleAction(chain => chain.toggleCodeBlock())}
          />
          <MenuButton
            label="Blockquote"
            active={editor.isActive("blockquote")}
            onClick={() => handleAction(chain => chain.toggleBlockquote())}
          />
          <MenuButton
            label="Divider"
            onClick={() => handleAction(chain => chain.setHorizontalRule())}
          />
        </div>
      )}
    </>
  );
}
