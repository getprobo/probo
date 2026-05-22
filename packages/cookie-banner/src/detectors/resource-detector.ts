// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { Detector } from "./detector";
import { isExtensionCaller, isExtensionContext } from "./extension-context";
import type { ReportQueue } from "./report-queue";
import type { ResourceType } from "./types";

// Map browser-reported PerformanceResourceTiming.initiatorType to the
// server-side tracker_resource_type. Anything we cannot classify is
// dropped rather than reported as "other" to keep the table tidy.
function mapInitiatorType(it: string): ResourceType | null {
  switch (it) {
    case "script":
      return "script";
    case "iframe":
      return "iframe";
    case "img":
    case "image":
    case "imageset":
    case "input":
      return "image";
    case "css":
    case "link":
      return "stylesheet";
    case "font":
      return "font";
    case "beacon":
    case "ping":
      return "beacon";
    case "fetch":
    case "xmlhttprequest":
      return "fetch";
    case "video":
    case "audio":
    case "track":
    case "embed":
    case "object":
      return "media";
    default:
      return null;
  }
}

// Attribute names that can carry a resource URL on at least one of the
// element types we track. Used as a fast-path filter at the top of the
// `setAttribute`/`setAttributeNS` wrap so non-resource calls (e.g.
// `setAttribute("class", ...)`) bail before we ever touch the stack.
const RESOURCE_ATTRIBUTES: ReadonlySet<string> = new Set([
  "src",
  "href",
  "data",
]);

// SavedDescriptor lets stop() restore the exact (configurable, enumerable,
// get, set) shape we replaced via Object.defineProperty.
interface SavedDescriptor {
  target: object;
  key: string;
  descriptor: PropertyDescriptor;
}

// Hard cap on the synchronously-marked extension URL set so a malicious
// extension cannot grow the SDK's memory footprint indefinitely. FIFO
// eviction is fine because the most recent injections are also the most
// likely to still be loading and need filtering.
const MAX_EXTENSION_URLS = 256;

export class ResourceDetector implements Detector {
  private readonly queue: ReportQueue;
  private readonly pageOrigin: string;
  private readonly apiOrigin: string;
  private observer: MutationObserver | null = null;
  private perfObserver: PerformanceObserver | null = null;
  private originalSWRegister: typeof ServiceWorkerContainer.prototype.register | null = null;

  // extensionElements is populated synchronously by the property-setter,
  // setAttribute, and HTML-parsing wraps when the call originates from a
  // browser-extension stack. The MutationObserver consults it before
  // reporting, which is the only way we can reliably attribute a
  // resource insertion to an extension -- the observer's own callback
  // fires from a browser-internal stack with no extension frame visible.
  private extensionElements: WeakSet<Element> = new WeakSet();

  // extensionUrls covers the URL-only paths (PerformanceObserver, fetch,
  // XHR, sendBeacon) where we have no element to key on. It mirrors the
  // origin+pathname identifier used in processResource so a hit drops
  // the report without further work. Capped at MAX_EXTENSION_URLS with
  // FIFO eviction.
  private extensionUrls: Set<string> = new Set();

  // Saved property descriptors restored by stop().
  private savedDescriptors: SavedDescriptor[] = [];

  // Saved originals for direct method/global replacements restored by
  // stop().
  private originalSetAttribute: typeof Element.prototype.setAttribute | null = null;
  private originalSetAttributeNS: typeof Element.prototype.setAttributeNS | null = null;
  private originalInsertAdjacentHTML: typeof Element.prototype.insertAdjacentHTML | null = null;
  private originalDocWrite: typeof Document.prototype.write | null = null;
  private originalDocWriteln: typeof Document.prototype.writeln | null = null;
  private originalFetch: typeof window.fetch | null = null;
  private originalXHROpen: typeof XMLHttpRequest.prototype.open | null = null;
  private originalSendBeacon: typeof navigator.sendBeacon | null = null;

  constructor(queue: ReportQueue, apiOrigin: string) {
    this.queue = queue;
    this.pageOrigin = location.origin;
    this.apiOrigin = apiOrigin;
  }

  start(): void {
    this.queue.onNotFound(() => this.stop());

    this.observeMutations();
    this.observePerformance();
    this.wrapServiceWorker();

    this.wrapElementSrcSetters();
    this.wrapSetAttribute();
    this.wrapHTMLParsing();
    this.wrapNetworkAPIs();

    if (isExtensionContext()) return;

    this.scanExisting();
    this.scanServiceWorkers();
  }

  stop(): void {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }

    if (this.perfObserver) {
      this.perfObserver.disconnect();
      this.perfObserver = null;
    }

    if (
      this.originalSWRegister
      && typeof navigator !== "undefined"
      && navigator.serviceWorker
    ) {
      navigator.serviceWorker.register = this.originalSWRegister;
      this.originalSWRegister = null;
    }

    for (const saved of this.savedDescriptors) {
      try {
        Object.defineProperty(saved.target, saved.key, saved.descriptor);
      } catch {
        // Restoration is best-effort: another piece of code may have
        // re-defined the same property after we did. We swallow rather
        // than throw because failing to restore is not a correctness
        // problem -- our wrap simply keeps running until the page
        // unloads.
      }
    }
    this.savedDescriptors = [];

    if (this.originalSetAttribute) {
      Element.prototype.setAttribute = this.originalSetAttribute;
      this.originalSetAttribute = null;
    }

    if (this.originalSetAttributeNS) {
      Element.prototype.setAttributeNS = this.originalSetAttributeNS;
      this.originalSetAttributeNS = null;
    }

    if (this.originalInsertAdjacentHTML) {
      Element.prototype.insertAdjacentHTML = this.originalInsertAdjacentHTML;
      this.originalInsertAdjacentHTML = null;
    }

    if (this.originalDocWrite) {
      Document.prototype.write = this.originalDocWrite;
      this.originalDocWrite = null;
    }

    if (this.originalDocWriteln) {
      Document.prototype.writeln = this.originalDocWriteln;
      this.originalDocWriteln = null;
    }

    if (this.originalFetch && typeof window !== "undefined") {
      window.fetch = this.originalFetch;
      this.originalFetch = null;
    }

    if (this.originalXHROpen && typeof XMLHttpRequest !== "undefined") {
      XMLHttpRequest.prototype.open = this.originalXHROpen;
      this.originalXHROpen = null;
    }

    if (this.originalSendBeacon && typeof navigator !== "undefined") {
      navigator.sendBeacon = this.originalSendBeacon;
      this.originalSendBeacon = null;
    }
  }

  private scanExisting(): void {
    for (const script of document.querySelectorAll<HTMLScriptElement>("script[src]")) {
      this.processResource(script.src, "script");
    }
    for (const iframe of document.querySelectorAll<HTMLIFrameElement>("iframe[src]")) {
      this.processResource(iframe.src, "iframe");
    }
  }

  private observeMutations(): void {
    this.observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        for (const node of mutation.addedNodes) {
          if (!(node instanceof HTMLElement)) continue;
          if (this.extensionElements.has(node)) continue;

          if (node instanceof HTMLScriptElement && node.src) {
            this.processResource(node.src, "script");
          } else if (node instanceof HTMLIFrameElement && node.src) {
            this.processResource(node.src, "iframe");
          }

          for (const script of node.querySelectorAll<HTMLScriptElement>("script[src]")) {
            if (this.extensionElements.has(script)) continue;
            this.processResource(script.src, "script");
          }
          for (const iframe of node.querySelectorAll<HTMLIFrameElement>("iframe[src]")) {
            if (this.extensionElements.has(iframe)) continue;
            this.processResource(iframe.src, "iframe");
          }
        }
      }
    });

    this.observer.observe(document.documentElement, {
      childList: true,
      subtree: true,
    });
  }

  // observePerformance picks up resources the DOM scan misses: tracking
  // pixels (<img>), beacons, fetch/XHR call-homes, CSS-loaded fonts and
  // sub-stylesheets, video/audio embeds. `buffered: true` replays any
  // entries that fired before the observer was attached, so we catch
  // bootstrap resources too.
  private observePerformance(): void {
    if (typeof PerformanceObserver === "undefined") return;

    try {
      this.perfObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries() as PerformanceResourceTiming[]) {
          const rt = mapInitiatorType(entry.initiatorType);
          if (rt) this.processResource(entry.name, rt);
        }
      });
      this.perfObserver.observe({ type: "resource", buffered: true });
    } catch {
      // Older browsers may not support the `type` option or the
      // `'resource'` entry type. Silently degrade to MutationObserver
      // coverage only.
      this.perfObserver = null;
    }
  }

  // wrapServiceWorker intercepts navigator.serviceWorker.register so
  // each registration -- even ones initiated by third-party SDKs --
  // surfaces as a tracker_resource entry keyed on the worker script
  // origin+path.
  private wrapServiceWorker(): void {
    if (typeof navigator === "undefined" || !navigator.serviceWorker) return;

    const sw = navigator.serviceWorker;
    const originalRegister = sw.register.bind(sw);
    this.originalSWRegister = originalRegister;

    const self = this;

    sw.register = function (
      scriptURL: string | URL,
      options?: RegistrationOptions,
    ): Promise<ServiceWorkerRegistration> {
      const url = typeof scriptURL === "string" ? scriptURL : scriptURL.toString();
      self.processResource(url, "service_worker");
      return originalRegister(scriptURL, options);
    };
  }

  // scanServiceWorkers enumerates registrations that pre-date the SDK
  // (e.g. installed on a previous visit, restored from cache).
  private scanServiceWorkers(): void {
    if (typeof navigator === "undefined" || !navigator.serviceWorker) return;

    navigator.serviceWorker
      .getRegistrations()
      .then((registrations) => {
        for (const r of registrations) {
          const url
            = r.active?.scriptURL
              ?? r.installing?.scriptURL
              ?? r.waiting?.scriptURL;
          if (url) this.processResource(url, "service_worker");
        }
      })
      .catch(() => {
        // Some browsers throw NotSupportedError in insecure contexts.
      });
  }

  private processResource(src: string, resourceType: ResourceType): void {
    if (isExtensionCaller()) return;

    let parsed: URL;
    try {
      parsed = new URL(src, location.href);
    } catch {
      return;
    }

    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return;
    // Service workers are always same-origin per browser security rules,
    // so we never drop them based on the page origin -- they are tracked
    // regardless of where the script lives. The apiOrigin guard still
    // applies so we never report our own SDK assets.
    if (parsed.origin === this.apiOrigin) return;
    if (resourceType !== "service_worker" && parsed.origin === this.pageOrigin) return;

    const identifier = parsed.origin + parsed.pathname;
    if (this.extensionUrls.has(identifier)) return;

    this.queue.reportResource({ url: identifier, resource_type: resourceType });
  }

  // identifierOf normalises a raw URL string into the same origin+pathname
  // shape used by processResource, returning null for anything that is
  // not an http(s) URL. Used by the synchronous wraps to populate
  // extensionUrls and stay consistent with the value the async paths
  // will later compare against.
  private identifierOf(src: string): string | null {
    let parsed: URL;
    try {
      parsed = new URL(src, location.href);
    } catch {
      return null;
    }
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return null;
    return parsed.origin + parsed.pathname;
  }

  private markExtensionElement(el: Element): void {
    this.extensionElements.add(el);
  }

  private markExtensionUrl(id: string | null): void {
    if (!id) return;
    if (this.extensionUrls.size >= MAX_EXTENSION_URLS) {
      const first = this.extensionUrls.values().next().value;
      if (first !== undefined) this.extensionUrls.delete(first);
    }
    this.extensionUrls.add(id);
  }

  private markExtension(el: Element | null, url: string | null): void {
    if (el) this.markExtensionElement(el);
    if (url) this.markExtensionUrl(url);
  }

  // wrapElementSrcSetters wraps the `src` (or `href` for <link>) IDL
  // setter on each prototype that loads a tracker-relevant resource.
  // Because the wrap runs on the caller's synchronous stack,
  // isExtensionCaller() actually sees the extension frame here -- which
  // it cannot from MutationObserver/PerformanceObserver. This is the
  // primary point at which extension attribution becomes reliable.
  private wrapElementSrcSetters(): void {
    this.wrapPropertySetter(
      HTMLScriptElement.prototype,
      "src",
      (_el, absolute) => this.processResource(absolute, "script"),
    );
    this.wrapPropertySetter(
      HTMLIFrameElement.prototype,
      "src",
      (_el, absolute) => this.processResource(absolute, "iframe"),
    );
    this.wrapPropertySetter(
      HTMLImageElement.prototype,
      "src",
      (_el, absolute) => this.processResource(absolute, "image"),
    );
    this.wrapPropertySetter(
      HTMLLinkElement.prototype,
      "href",
      (el, absolute) => {
        // <link> only initiates a load for some `rel` values; for the
        // others (icon, manifest, dns-prefetch, ...) we leave reporting
        // to PerformanceObserver if and when an actual load occurs.
        // preload/prefetch/modulepreload are deferred too because their
        // mapped resource type depends on the `as` attribute.
        if ((el as HTMLLinkElement).rel.toLowerCase() === "stylesheet") {
          this.processResource(absolute, "stylesheet");
        }
      },
    );
    if (typeof HTMLSourceElement !== "undefined") {
      this.wrapPropertySetter(
        HTMLSourceElement.prototype,
        "src",
        (_el, absolute) => this.processResource(absolute, "media"),
      );
    }
  }

  private wrapPropertySetter(
    proto: object,
    key: string,
    onPageCaller: (el: Element, absolute: string) => void,
  ): void {
    const desc = Object.getOwnPropertyDescriptor(proto, key);
    if (!desc?.set || !desc?.get) return;

    this.savedDescriptors.push({ target: proto, key, descriptor: desc });

    const originalSet = desc.set;
    const originalGet = desc.get;
    const self = this;

    Object.defineProperty(proto, key, {
      configurable: true,
      enumerable: desc.enumerable,
      get: originalGet,
      set(value: unknown) {
        originalSet.call(this, value);
        if (typeof value !== "string" || value === "") return;

        const fromExtension = isExtensionCaller();
        // The original getter resolves relative URLs against the
        // document base. Reading through it gives the same absolute
        // form processResource would later produce.
        const absolute = (originalGet.call(this) as string) || value;

        if (fromExtension) {
          self.markExtension(this as Element, self.identifierOf(absolute));
          return;
        }
        onPageCaller(this as Element, absolute);
      },
    });
  }

  // wrapSetAttribute installs a single hook on Element.prototype that
  // covers every `el.setAttribute("src" | "href" | "data", ...)` call,
  // regardless of the element type. The first thing it does is bail on
  // attribute names we never care about, keeping the overhead near zero
  // on the hot path of generic setAttribute usage (class names, ARIA
  // attributes, dataset entries, ...).
  private wrapSetAttribute(): void {
    const original = Element.prototype.setAttribute;
    this.originalSetAttribute = original;
    const self = this;

    Element.prototype.setAttribute = function (
      this: Element,
      name: string,
      value: string,
    ): void {
      original.call(this, name, value);
      if (typeof name !== "string") return;
      // Common case: name is already lowercase. Skip the toLowerCase()
      // allocation by checking the canonical names directly first.
      if (
        name !== "src"
        && name !== "href"
        && name !== "data"
        && !RESOURCE_ATTRIBUTES.has(name.toLowerCase())
      ) {
        return;
      }
      const lower = name === "src" || name === "href" || name === "data"
        ? name
        : name.toLowerCase();
      self.handleAttributeMutation(this, lower, value);
    };

    if (typeof Element.prototype.setAttributeNS === "function") {
      const originalNS = Element.prototype.setAttributeNS;
      this.originalSetAttributeNS = originalNS;

      Element.prototype.setAttributeNS = function (
        this: Element,
        ns: string | null,
        name: string,
        value: string,
      ): void {
        originalNS.call(this, ns, name, value);
        if (typeof name !== "string") return;
        // Strip any namespace prefix so xlink:href etc. resolve to the
        // local name we filter on.
        const colon = name.indexOf(":");
        const local = (colon >= 0 ? name.slice(colon + 1) : name).toLowerCase();
        if (!RESOURCE_ATTRIBUTES.has(local)) return;
        self.handleAttributeMutation(this, local, value);
      };
    }
  }

  private handleAttributeMutation(
    el: Element,
    attrName: string,
    value: unknown,
  ): void {
    if (typeof value !== "string" || value === "") return;

    // Extension marking runs before classification because the
    // current attribute alone is not always enough to know whether a
    // load will happen. The canonical case is `<link>`: a stylesheet
    // load is only triggered once both `href` and `rel="stylesheet"`
    // are set, in any order. If the extension sets `href` first we
    // would otherwise miss tagging, and the eventual PerformanceObserver
    // entry -- which has no extension frame on its stack -- would leak
    // through as a page tracker. The same holds for `<link rel=preload>`
    // and any future rel value that initiates a load.
    if (isExtensionCaller()) {
      if (this.couldLoadResource(el, attrName)) {
        this.markExtension(el, this.identifierOf(value));
      }
      return;
    }

    const rt = this.resourceTypeForElement(el, attrName);
    if (rt === null) return;

    this.processResource(value, rt);
  }

  // couldLoadResource reports whether `el` is an element type that can
  // ever initiate a network load via `attrName`, regardless of any
  // other attributes that may or may not be set yet. It is a superset
  // of resourceTypeForElement: it returns true for `<link href>` even
  // when `rel` has not been set, because the rel may change later and
  // turn the href into a real load.
  private couldLoadResource(el: Element, attrName: string): boolean {
    if (attrName === "src") {
      return (
        el instanceof HTMLScriptElement
        || el instanceof HTMLIFrameElement
        || el instanceof HTMLImageElement
        || (typeof HTMLSourceElement !== "undefined" && el instanceof HTMLSourceElement)
        || (typeof HTMLEmbedElement !== "undefined" && el instanceof HTMLEmbedElement)
        || (typeof HTMLMediaElement !== "undefined" && el instanceof HTMLMediaElement)
      );
    }
    if (attrName === "href") {
      return el instanceof HTMLLinkElement;
    }
    if (attrName === "data") {
      return typeof HTMLObjectElement !== "undefined" && el instanceof HTMLObjectElement;
    }
    return false;
  }

  private resourceTypeForElement(el: Element, attrName: string): ResourceType | null {
    if (attrName === "src") {
      if (el instanceof HTMLScriptElement) return "script";
      if (el instanceof HTMLIFrameElement) return "iframe";
      if (el instanceof HTMLImageElement) return "image";
      if (typeof HTMLSourceElement !== "undefined" && el instanceof HTMLSourceElement) return "media";
      if (typeof HTMLEmbedElement !== "undefined" && el instanceof HTMLEmbedElement) return "media";
      if (typeof HTMLMediaElement !== "undefined" && el instanceof HTMLMediaElement) return "media";
      return null;
    }
    if (attrName === "href") {
      if (el instanceof HTMLLinkElement) {
        // Only report stylesheet loads from the sync hook; other rel
        // values either don't load (dns-prefetch, preconnect, icon,
        // manifest) or have a resource type that depends on the `as`
        // attribute, which is best handled by PerformanceObserver.
        return el.rel.toLowerCase() === "stylesheet" ? "stylesheet" : null;
      }
      return null;
    }
    if (attrName === "data") {
      if (typeof HTMLObjectElement !== "undefined" && el instanceof HTMLObjectElement) return "media";
      return null;
    }
    return null;
  }

  // wrapHTMLParsing covers the four entry points where the browser HTML
  // parser builds element trees from a string: setting innerHTML or
  // outerHTML, insertAdjacentHTML, and document.write/writeln. None of
  // those invoke the per-element setter or setAttribute hooks above
  // (the parser writes attributes via internal C++), so we capture the
  // extension verdict synchronously at the entry point and then walk
  // the parsed result to tag the new resource-bearing descendants.
  private wrapHTMLParsing(): void {
    this.wrapInnerHTMLSetter();
    this.wrapOuterHTMLSetter();
    this.wrapInsertAdjacentHTML();
    this.wrapDocumentWrite();
  }

  private wrapInnerHTMLSetter(): void {
    const desc = Object.getOwnPropertyDescriptor(Element.prototype, "innerHTML");
    if (!desc?.set || !desc?.get) return;

    this.savedDescriptors.push({ target: Element.prototype, key: "innerHTML", descriptor: desc });

    const originalSet = desc.set;
    const originalGet = desc.get;
    const self = this;

    Object.defineProperty(Element.prototype, "innerHTML", {
      configurable: true,
      enumerable: desc.enumerable,
      get: originalGet,
      set(value: unknown) {
        const fromExtension = isExtensionCaller();
        originalSet.call(this, value);
        if (!fromExtension) return;
        // innerHTML replaces every existing child, so after the call
        // every descendant is new.
        self.markResourceTree(this as Element, true);
      },
    });
  }

  private wrapOuterHTMLSetter(): void {
    const desc = Object.getOwnPropertyDescriptor(Element.prototype, "outerHTML");
    if (!desc?.set || !desc?.get) return;

    this.savedDescriptors.push({ target: Element.prototype, key: "outerHTML", descriptor: desc });

    const originalSet = desc.set;
    const originalGet = desc.get;
    const self = this;

    Object.defineProperty(Element.prototype, "outerHTML", {
      configurable: true,
      enumerable: desc.enumerable,
      get: originalGet,
      set(value: unknown) {
        const fromExtension = isExtensionCaller();
        const parent = (this as Element).parentNode;
        if (!fromExtension || !parent) {
          originalSet.call(this, value);
          return;
        }
        // outerHTML replaces this element with parsed siblings inside
        // the parent. We can't walk the parent wholesale (pre-existing
        // children would be wrongly marked), so we use a one-shot
        // MutationObserver to capture only the actually-added nodes.
        self.observeAndMark(parent, () => originalSet.call(this, value));
      },
    });
  }

  private wrapInsertAdjacentHTML(): void {
    const original = Element.prototype.insertAdjacentHTML;
    if (typeof original !== "function") return;

    this.originalInsertAdjacentHTML = original;
    const self = this;

    Element.prototype.insertAdjacentHTML = function (
      this: Element,
      position: InsertPosition,
      text: string,
    ): void {
      const fromExtension = isExtensionCaller();
      if (!fromExtension) {
        original.call(this, position, text);
        return;
      }

      // The affected parent depends on the position: beforebegin and
      // afterend insert siblings of `this` (so observe parentNode);
      // afterbegin and beforeend insert children of `this`.
      const root: Node | null
        = position === "beforebegin" || position === "afterend"
          ? this.parentNode
          : this;

      if (!root) {
        original.call(this, position, text);
        return;
      }

      self.observeAndMark(root, () => original.call(this, position, text));
    };
  }

  private wrapDocumentWrite(): void {
    if (typeof Document === "undefined") return;

    const originalWrite = Document.prototype.write;
    const originalWriteln = Document.prototype.writeln;
    if (typeof originalWrite !== "function") return;

    this.originalDocWrite = originalWrite;
    this.originalDocWriteln = typeof originalWriteln === "function" ? originalWriteln : null;

    const self = this;

    Document.prototype.write = function (this: Document, ...args: string[]): void {
      const fromExtension = isExtensionCaller();
      if (!fromExtension) {
        originalWrite.apply(this, args);
        return;
      }
      self.observeAndMark(this, () => originalWrite.apply(this, args));
    };

    if (typeof originalWriteln === "function") {
      Document.prototype.writeln = function (this: Document, ...args: string[]): void {
        const fromExtension = isExtensionCaller();
        if (!fromExtension) {
          originalWriteln.apply(this, args);
          return;
        }
        self.observeAndMark(this, () => originalWriteln.apply(this, args));
      };
    }
  }

  // observeAndMark wraps a single synchronous DOM mutation in a
  // disposable MutationObserver so we get a precise list of the nodes
  // the operation actually inserted. takeRecords() drains the queue
  // synchronously, before the main observer's microtask runs, so we
  // can tag the new elements before observeMutations() sees them.
  private observeAndMark(root: Node, fn: () => void): void {
    if (typeof MutationObserver === "undefined") {
      fn();
      return;
    }

    const observer = new MutationObserver(() => {});
    try {
      observer.observe(root, { childList: true, subtree: true });
    } catch {
      // observe() throws on detached or unusual roots. Fall back to
      // running the operation untracked rather than failing the page.
      fn();
      return;
    }

    try {
      fn();
    } finally {
      const records = observer.takeRecords();
      observer.disconnect();
      for (const record of records) {
        for (const node of record.addedNodes) {
          if (node instanceof HTMLElement) {
            this.markResourceTree(node, true);
          }
        }
      }
    }
  }

  // markResourceTree walks `root` and its descendants and tags any
  // resource-bearing element. `fromParser` distinguishes elements
  // produced by the HTML parser (innerHTML/outerHTML/insertAdjacentHTML/
  // document.write): parser-inserted <script> elements have their
  // already-started flag set and never fetch their src, so we tag them
  // for MutationObserver suppression but deliberately do not poison
  // extensionUrls -- that URL might still be a legitimate page tracker
  // loaded from elsewhere.
  private markResourceTree(root: Element, fromParser: boolean): void {
    this.tryMarkElement(root, fromParser);
    for (const el of root.querySelectorAll("*")) {
      this.tryMarkElement(el, fromParser);
    }
  }

  private tryMarkElement(el: Element, fromParser: boolean): void {
    let url: string | null = null;
    let willLoad = true;

    if (el instanceof HTMLScriptElement) {
      if (!el.src) return;
      url = el.src;
      // Parser-inserted scripts are flagged already-started by the
      // spec and never fetch their src. Tag the element for the
      // MutationObserver path but skip URL marking.
      willLoad = !fromParser;
    } else if (el instanceof HTMLIFrameElement) {
      if (!el.src) return;
      url = el.src;
    } else if (el instanceof HTMLImageElement) {
      if (!el.src) return;
      url = el.src;
    } else if (el instanceof HTMLLinkElement) {
      if (!el.href) return;
      url = el.href;
    } else if (typeof HTMLSourceElement !== "undefined" && el instanceof HTMLSourceElement) {
      if (!el.src) return;
      url = el.src;
    } else if (typeof HTMLEmbedElement !== "undefined" && el instanceof HTMLEmbedElement) {
      if (!el.src) return;
      url = el.src;
    } else if (typeof HTMLObjectElement !== "undefined" && el instanceof HTMLObjectElement) {
      if (!el.data) return;
      url = el.data;
    } else {
      return;
    }

    this.markExtensionElement(el);
    if (willLoad) this.markExtensionUrl(this.identifierOf(url));
  }

  // wrapNetworkAPIs covers the resource paths that never produce a DOM
  // element: fetch, XMLHttpRequest, and sendBeacon. Each of these has a
  // synchronous entry point with the caller's stack on top, so the
  // extension check here is reliable in a way that PerformanceObserver
  // never can be.
  private wrapNetworkAPIs(): void {
    this.wrapFetch();
    this.wrapXHR();
    this.wrapSendBeacon();
  }

  private wrapFetch(): void {
    if (typeof window === "undefined" || typeof window.fetch !== "function") return;

    const original = window.fetch;
    this.originalFetch = original;
    const self = this;

    window.fetch = function (
      input: RequestInfo | URL,
      init?: RequestInit,
    ): Promise<Response> {
      const url = self.urlFromFetchInput(input);
      if (url) {
        if (isExtensionCaller()) {
          self.markExtensionUrl(self.identifierOf(url));
        } else {
          self.processResource(url, "fetch");
        }
      }
      return original.call(this, input, init);
    };
  }

  private urlFromFetchInput(input: RequestInfo | URL): string | null {
    try {
      if (typeof input === "string") return input;
      if (input instanceof URL) return input.toString();
      if (typeof Request !== "undefined" && input instanceof Request) return input.url;
    } catch {
      return null;
    }
    return null;
  }

  private wrapXHR(): void {
    if (typeof XMLHttpRequest === "undefined") return;

    const original = XMLHttpRequest.prototype.open;
    this.originalXHROpen = original;
    const self = this;

    // The ECMA signature of XHR.open is variadic (2 to 5 args) and
    // treats `undefined` for the trailing async/username/password as
    // "not passed" with spec-defined defaults. Forwarding via apply
    // preserves whatever the caller actually provided.
    function wrappedOpen(this: XMLHttpRequest): void {
      const args = arguments as unknown as [string, string | URL, ...unknown[]];
      const rawUrl = args[1];
      if (typeof rawUrl === "string" || rawUrl instanceof URL) {
        const url = typeof rawUrl === "string" ? rawUrl : rawUrl.toString();
        if (isExtensionCaller()) {
          self.markExtensionUrl(self.identifierOf(url));
        } else {
          self.processResource(url, "fetch");
        }
      }
      // eslint-disable-next-line prefer-rest-params
      (original as (...a: unknown[]) => void).apply(this, args);
    }

    XMLHttpRequest.prototype.open = wrappedOpen as typeof XMLHttpRequest.prototype.open;
  }

  private wrapSendBeacon(): void {
    if (typeof navigator === "undefined" || typeof navigator.sendBeacon !== "function") return;

    const original = navigator.sendBeacon;
    this.originalSendBeacon = original;
    const self = this;

    navigator.sendBeacon = function (
      this: Navigator,
      url: string | URL,
      data?: BodyInit | null,
    ): boolean {
      const urlStr = typeof url === "string" ? url : url.toString();
      if (isExtensionCaller()) {
        self.markExtensionUrl(self.identifierOf(urlStr));
      } else {
        self.processResource(urlStr, "beacon");
      }
      return original.call(this, url, data);
    };
  }
}
