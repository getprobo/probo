export function PDFPreview({ src }: { src: string }) {
  const url = new URL(window.location.origin + "/pdfjs/web/viewer.html");
  url.searchParams.set("file", encodeURI(src));

  return <iframe src={url.toString()} className="size-full block" />;
}
