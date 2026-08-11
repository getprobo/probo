// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { CornersInIcon, CornersOutIcon } from "@phosphor-icons/react";
import { clsx } from "clsx";
import { memo, useCallback, useEffect, useId, useRef, useState } from "react";
import svgPanZoom from "svg-pan-zoom";

import { mermaidRenderErrorToast, renderMermaidDiagram } from "../../lib/mermaid";
import { Button } from "../Button/Button";
import { IconMinusLarge, IconPlusLarge, IconRotateCw } from "../Icons";
import { useToast } from "../Toasts/Toasts";

import { addMermaidInteractions } from "./MermaidInteractions";

type Props = {
  chart: string;
};

export function MermaidDiagram({ chart }: Props) {
  const svgRef = useRef<SVGElement>(null);
  const zoomRef = useRef<SvgPanZoom.Instance>(null);
  const wrapper = useRef<HTMLDivElement>(null);
  const style = useRef("");
  const [fullscreen, setFullScreen] = useState(false);

  // Handle fullscreen state change
  useEffect(() => {
    function onFullscreenChange() {
      const isFullScreen = document.fullscreenElement === wrapper.current;
      svgRef.current?.setAttribute("style", isFullScreen ? "width: 100%; height: 100%;" : style.current);
      zoomRef.current?.resize();
      zoomRef.current?.reset();
      setFullScreen(document.fullscreenElement === wrapper.current);
    }

    document.addEventListener("fullscreenchange", onFullscreenChange);
    return () => {
      document.removeEventListener("fullscreenchange", onFullscreenChange);
    };
  }, []);

  const zoomIn = () => {
    zoomRef.current?.zoomIn();
  };
  const zoomOut = () => {
    zoomRef.current?.zoomOut();
  };
  const zoomReset = () => {
    zoomRef.current?.resetZoom();
    zoomRef.current?.resetPan();
  };
  const toggleFullscreen = () => {
    if (document.fullscreenElement === wrapper.current) {
      document.exitFullscreen().catch(console.error);
      return
    }
    wrapper.current?.requestFullscreen().catch(console.error);
  };
  const onSVGLoaded = useCallback((svg: SVGElement, zoom: SvgPanZoom.Instance) => {
    svgRef.current = svg;
    zoomRef.current = zoom;
    style.current = svg.getAttribute("style") ?? "";
  }, []);

  if (!chart) {
    return null;
  }

  return (
    <div className={clsx("relative w-full")} ref={wrapper}>
      <MermaidSVG
        source={chart}
        onSVG={onSVGLoaded}
      />
      <div className={clsx("flex flex-col gap-2 absolute", fullscreen ? "bottom-4 right-4" : "bottom-0 right-0")}>
        <Button
          icon={fullscreen ? CornersInIcon : CornersOutIcon}
          onClick={toggleFullscreen}
          variant="secondary"
        />
        <Button variant="secondary" icon={IconRotateCw} onClick={zoomReset} />
        <Button variant="secondary" icon={IconPlusLarge} onClick={zoomIn} />
        <Button variant="secondary" icon={IconMinusLarge} onClick={zoomOut} />
      </div>
    </div>
  );
}

const MermaidSVG = memo(({ source, onSVG }: { source: string; onSVG: (el: SVGElement, zoom: SvgPanZoom.Instance) => void }) => {
  const id = useId().replace(/:/g, "");
  const { toast } = useToast();
  const [error, setError] = useState(false);
  const div = useRef<HTMLDivElement>(null);

  // Load and render mermaid SVG
  useEffect(() => {
    if (!source) {
      return;
    }
    let destroy = () => {};
    renderMermaidDiagram(`mermaid-${id}`, source)
      .then((r) => {
        const element = div.current;
        if (!element || element.innerHTML) {
          return;
        }
        element.innerHTML = r.svg;
        r.bindFunctions?.(element);
        const svg = element.firstElementChild;
        if (!(svg instanceof SVGElement)) {
          return;
        }
        const { width, height } = svg.getBoundingClientRect();
        svg.style.aspectRatio = `${width}/${height}`;
        svg.style.minHeight = "200px";
        const zoom = svgPanZoom(svg);
        const removeInteractions = addMermaidInteractions(svg);
        destroy = () => {
          zoom.destroy();
          removeInteractions();
        };
        onSVG(svg, zoom);
      })
      .catch(() => {
        setError(true);
        toast(mermaidRenderErrorToast);
      });

    return () => {
      destroy();
    };
  }, [source, id, toast, onSVG]);

  if (error) {
    return (
      <pre className="border border-border-solid rounded p-4 bg-transparent font-mono text-sm overflow-x-auto text-inherit">
        <code>{source}</code>
      </pre>
    );
  }

  return (
    <div id={id} className="contents" ref={div} />
  );
});

MermaidSVG.displayName = "MermaidSVG";
