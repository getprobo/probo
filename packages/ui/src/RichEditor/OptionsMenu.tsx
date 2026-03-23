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
import { CodeBlockIcon, DotsSixVerticalIcon, ListBulletsIcon, ListNumbersIcon, QuotesIcon, TextHOneIcon, TextHThreeIcon, TextHTwoIcon, TextTIcon } from "@phosphor-icons/react";
import { NodeSelection, TextSelection } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { type useEditor } from "@tiptap/react";
import { type DragEvent, useEffect, useLayoutEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { MenuButton } from "./MenuButton";

const optionsMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex size-6 items-center justify-center",
      "rounded text-txt-tertiary hover:bg-subtle hover:text-txt-primary cursor-grab",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});

const { trigger, menu } = optionsMenuVariants();

function findClosestRootBlock(element: Element, editorDom: Element): HTMLElement | null {
  let current: Element | null = element;

  while (current?.parentElement && current.parentElement !== editorDom) {
    current = current.parentElement;
  }

  return current?.parentElement === editorDom ? (current as HTMLElement) : null;
}

function startDrag(view: EditorView, slice: ReturnType<NodeSelection["content"]>) {
  view.dragging = { slice, move: true };
}

type OptionsMenuProps = {
  editor: ReturnType<typeof useEditor>;
};

export function OptionsMenu({ editor }: OptionsMenuProps) {
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
    middleware: [offset(8)],
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

  const getNodeAtHoveredBlock = () => {
    if (!hoveredBlock) return null;
    try {
      const pos = editor.view.posAtDOM(hoveredBlock, 0);
      const $pos = editor.state.doc.resolve(pos);
      if ($pos.depth >= 1) {
        return { node: $pos.node(1), pos: $pos.before(1) };
      }
      const nodeAfter = $pos.nodeAfter;
      if (nodeAfter) {
        return { node: nodeAfter, pos: pos };
      }
      return null;
    } catch {
      return null;
    }
  };

  const isNodeType = (type: string, attrs?: Record<string, unknown>) => {
    const data = getNodeAtHoveredBlock();
    if (!data) return false;
    if (data.node.type.name !== type) return false;
    if (attrs) {
      return Object.entries(attrs).every(
        ([key, value]) => data.node.attrs[key] === value,
      );
    }
    return true;
  };

  const handleAction = (
    applyCommand: (chain: ReturnType<typeof editor.chain>) => ReturnType<typeof editor.chain>,
  ) => {
    const data = getNodeAtHoveredBlock();
    if (!data) {
      setMenuOpen(false);
      return;
    }

    try {
      if (!data.node.isTextblock) {
        if (!data.node.firstChild) {
          const paragraph = editor.state.schema.nodes.paragraph.create();
          editor.chain()
            .focus()
            .command(({ tr }) => {
              tr.replaceWith(data.pos, data.pos + data.node.nodeSize, paragraph);
              return true;
            })
            .run();

          const $near = editor.state.doc.resolve(data.pos + 1);
          const textPos = TextSelection.near($near).from;

          applyCommand(
            editor.chain()
              .focus()
              .setTextSelection(textPos),
          ).run();

          setMenuOpen(false);
          return;
        }

        let textBlock = data.node.firstChild;
        if (!textBlock) return;
        while (!textBlock.isTextblock && textBlock.firstChild) {
          textBlock = textBlock.firstChild;
        }
        if (!textBlock.isTextblock) return;

        const firstChildSize = data.node.firstChild.nodeSize;
        const paragraph = editor.state.schema.nodes.paragraph.create(
          null,
          textBlock.content,
        );

        editor.chain()
          .focus()
          .command(({ tr }) => {
            tr.insert(data.pos, paragraph);
            const wrapperPos = data.pos + paragraph.nodeSize;
            const wrapperNode = tr.doc.nodeAt(wrapperPos);
            if (!wrapperNode) return false;

            if (wrapperNode.childCount <= 1) {
              tr.delete(wrapperPos, wrapperPos + wrapperNode.nodeSize);
            } else {
              tr.delete(wrapperPos + 1, wrapperPos + 1 + firstChildSize);
            }

            return true;
          })
          .run();
      }

      const $near = editor.state.doc.resolve(data.pos + 1);
      const textPos = TextSelection.near($near).from;

      applyCommand(
        editor.chain()
          .focus()
          .setTextSelection(textPos),
      ).run();
    } catch {
      // Block may no longer be in the document
    }

    setMenuOpen(false);
  };

  const onDragStart = (e: DragEvent<HTMLButtonElement>) => {
    const data = getNodeAtHoveredBlock();
    if (!data || !hoveredBlock) return;

    try {
      const view = editor.view;
      const selection = NodeSelection.create(view.state.doc, data.pos);
      const slice = selection.content();

      const { tr } = view.state;
      tr.setSelection(selection);
      view.dispatch(tr);

      if (e.dataTransfer) {
        e.dataTransfer.clearData();
        e.dataTransfer.setData("text/plain", "");
        e.dataTransfer.effectAllowed = "move";

        const wrapper = document.createElement("div");
        wrapper.append(hoveredBlock.cloneNode(true));
        wrapper.style.position = "absolute";
        wrapper.style.top = "-10000px";
        document.body.append(wrapper);
        e.dataTransfer.setDragImage(wrapper, 0, 0);
        document.addEventListener("drop", () => wrapper.remove(), { once: true });
      }

      startDrag(view, slice);
    } catch {
      // Block may no longer be in the document
    }
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
        draggable
        onDragStart={onDragStart}
        onDragEnd={() => setHoveredBlock(null)}
        type="button"
        style={{
          ...triggerStyles,
          visibility: isPositioned ? "visible" : "hidden",
        }}
        className={trigger()}
      >
        <DotsSixVerticalIcon size={20} weight="bold" />
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
          <div className="p-1 font-semibold text-sm">Turn into</div>
          <MenuButton
            active={isNodeType("paragraph")}
            onClick={() => handleAction(chain => chain.setParagraph())}
          >
            <TextTIcon size={16} weight="bold" />
            Text
          </MenuButton>
          <MenuButton
            active={isNodeType("heading", { level: 1 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 1 }))}
          >
            <TextHOneIcon size={16} weight="bold" />
            Heading 1
          </MenuButton>
          <MenuButton
            active={isNodeType("heading", { level: 2 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 2 }))}
          >
            <TextHTwoIcon size={16} weight="bold" />
            Heading 2
          </MenuButton>
          <MenuButton
            active={isNodeType("heading", { level: 3 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 3 }))}
          >
            <TextHThreeIcon size={16} weight="bold" />
            Heading 3
          </MenuButton>
          <MenuButton
            active={isNodeType("bulletList")}
            onClick={() => handleAction(chain => chain.toggleBulletList())}
          >
            <ListBulletsIcon size={16} weight="bold" />
            Bullet List
          </MenuButton>
          <MenuButton
            active={isNodeType("orderedList")}
            onClick={() => handleAction(chain => chain.toggleOrderedList())}
          >
            <ListNumbersIcon size={16} weight="bold" />
            Ordered List
          </MenuButton>
          <MenuButton
            active={isNodeType("codeBlock")}
            onClick={() => handleAction(chain => chain.toggleCodeBlock())}
          >
            <CodeBlockIcon size={16} weight="bold" />
            Code Block
          </MenuButton>
          <MenuButton
            active={isNodeType("blockquote")}
            onClick={() => handleAction(chain => chain.toggleBlockquote())}
          >
            <QuotesIcon size={16} weight="bold" />
            Blockquote
          </MenuButton>
        </div>
      )}
    </>
  );
}
