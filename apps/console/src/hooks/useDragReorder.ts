import { useState } from "react";

/**
 * Encapsulates drag-and-drop reorder state and handlers for ranked list tables.
 * The caller owns the actual mutation via `onDrop(fromIndex, targetIndex)`, which
 * keeps field access (e.g. `node.rank`) colocated with the fragment that queries it.
 */
export function useDragReorder<TData>({
  onDrop,
}: {
  onDrop: (fromIndex: number, targetIndex: number) => void;
}) {
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dropIndex, setDropIndex] = useState<number | null>(null);
  const [draggedData, setDraggedData] = useState<TData | null>(null);

  const reset = () => {
    setDraggedIndex(null);
    setDropIndex(null);
    setDraggedData(null);
  };

  const handleDragStart = (index: number, data: TData) => {
    setDraggedIndex(index);
    setDraggedData(data);
    setDropIndex(null);
  };

  const handleDragEnter = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedIndex !== null && draggedIndex !== index) {
      setDropIndex(index);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  };

  const handleDragEnd = () => {
    reset();
  };

  const handleDrop = (e: React.DragEvent, targetIndex: number) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === targetIndex) {
      reset();
      return;
    }
    const fromIndex = draggedIndex;
    reset();
    onDrop(fromIndex, targetIndex);
  };

  return {
    draggedIndex,
    dropIndex,
    setDropIndex,
    draggedData,
    isDragging: draggedIndex !== null,
    handleDragStart,
    handleDragEnter,
    handleDragOver,
    handleDragEnd,
    handleDrop,
  };
}
