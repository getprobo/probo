import { times } from "@probo/helpers";
import {
  IconArrowInbox,
  IconChevronLeft,
  IconChevronRight,
  IconPlusLarge,
  Spinner,
} from "@probo/ui";
import { IconMinusLarge } from "@probo/ui/src/Atoms/Icons/IconMinusLarge";
import { type ComponentProps, useRef, useState } from "react";
import { Document, Page, pdfjs } from "react-pdf";

import "react-pdf/dist/Page/TextLayer.css";
import "react-pdf/dist/Page/AnnotationLayer.css";

// Worker for PDF.js
pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`;

const btnClass
  = "size-8 grid place-items-center hover:bg-secondary-hover cursor-pointer rounded-sm disabled:opacity-30 transition-all";

export function PDFPreview({ src, name }: { src: string; name?: string }) {
  const [numPages, setNumPages] = useState(0);
  const [scale, setScale] = useState(1.0);
  const [currentPage, setCurrentPage] = useState(1);
  const documentRef: ComponentProps<typeof Document>["ref"] = useRef(null);
  const wrapperRef = useRef<HTMLDivElement>(null);

  const onDocumentLoadSuccess: ComponentProps<
    typeof Document
  >["onLoadSuccess"] = (document) => {
    setNumPages(document.numPages);
    setCurrentPage(1);
  };

  const zoomFactor = (factor: number) => () => {
    setScale(scale * factor);
  };

  const movePage = (direction: 1 | -1) => () => {
    if (currentPage === 1 && direction === -1) {
      return;
    }
    const newPage = currentPage + direction;
    const page = documentRef.current?.pages.current[newPage - 1];
    if (!page) {
      return;
    }
    page.scrollIntoView({
      behavior: "smooth",
      block: "start",
      inline: "center",
    });
    setCurrentPage(newPage);
  };

  const resolveCurrentPage = () => {
    if (!wrapperRef.current) {
      return;
    }
    const pages = documentRef.current?.pages.current;
    if (!pages?.length) {
      return;
    }
    const parentRect = wrapperRef.current.getBoundingClientRect();
    const parentMiddleY = parentRect.top + parentRect.height / 2;
    for (let i = 0; i < pages.length; i++) {
      const childRect = pages[i].getBoundingClientRect();
      if (childRect.top <= parentMiddleY && childRect.bottom >= parentMiddleY) {
        return setCurrentPage(i + 1);
      }
    }
  };

  const handleDownload = () => {
    const link = document.createElement("a");
    link.href = src;
    link.download = name || "document.pdf";
    link.click();
  };

  return (
    <div className="grid grid-rows-[max-content_1fr] h-full bg-subtle">
      {/* Custom Zoom Controls */}
      <nav className="flex-none flex items-center gap-2 bg-level-1 py-3 text-sm pl-4 pr-3 text-txt-primary">
        <div>{name}</div>
        <div className="mx-auto flex gap-1 items-center">
          <button
            onClick={movePage(-1)}
            className={btnClass}
            disabled={currentPage === 1}
          >
            <IconChevronLeft size={16} />
          </button>
          <div>
            {currentPage}
            {" "}
            /
            {numPages}
          </div>
          <button onClick={movePage(1)} className={btnClass}>
            <IconChevronRight size={16} />
          </button>
        </div>
        <button onClick={zoomFactor(0.8)} className={btnClass}>
          <IconMinusLarge size={16} />
        </button>
        <button onClick={zoomFactor(1.2)} className={btnClass}>
          <IconPlusLarge size={16} />
        </button>
        <button onClick={handleDownload} className={btnClass}>
          <IconArrowInbox size={16} />
        </button>
      </nav>

      {/* PDF Document */}
      <div
        className="overflow-auto scroll-p-6"
        onScrollEnd={resolveCurrentPage}
        ref={wrapperRef}
      >
        <Document
          file={src}
          key={src + "document"}
          onLoadSuccess={onDocumentLoadSuccess}
          className="flex flex-col gap-4 py-10"
          ref={documentRef}
        >
          {numPages === 0 && <Spinner className="mx-auto" />}
          {times(numPages, index => (
            <Page
              className="w-max h-max mx-auto shadow-mid"
              key={index.toString()}
              pageNumber={index + 1}
              scale={scale} // Apply zoom via scale prop
            />
          ))}
        </Document>
      </div>
    </div>
  );
}
