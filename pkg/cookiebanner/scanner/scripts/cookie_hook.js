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

((bindingName) => {
  var report = globalThis[bindingName];
  try { delete globalThis[bindingName]; } catch (e) { }
  var proto = Document.prototype;
  var desc = Object.getOwnPropertyDescriptor(proto, "cookie");
  if (!desc || typeof desc.set !== "function") return;
  var nativeSet = desc.set;
  function hostIsIP(host) {
    if (host.indexOf(":") !== -1) return true;
    var parts = host.split(".");
    if (parts.length !== 4) return false;
    for (var i = 0; i < parts.length; i++) {
      if (!/^\d{1,3}$/.test(parts[i]) || Number(parts[i]) > 255) return false;
    }
    return true;
  }
  function cookieDomain(host, domain) {
    host = String(host || "").replace(/\.$/, "").toLowerCase();
    domain = String(domain || "").replace(/^\./, "").replace(/\.+$/, "").toLowerCase();
    if (!host || !domain || domain.indexOf(".") === -1 || hostIsIP(host)) return "";
    if (host !== domain && host.slice(-(domain.length + 1)) !== "." + domain) return "";
    return domain;
  }
  Object.defineProperty(proto, "cookie", {
    configurable: true,
    enumerable: desc.enumerable,
    get: desc.get,
    set: function (value) {
      var accepted = nativeSet.call(this, value);
      try {
        var raw = String(value);
        var name = raw.split("=")[0].trim();
        var domain = "";
        var parts = raw.split(";");
        for (var j = 1; j < parts.length; j++) {
          var part = parts[j].trim();
          if (part.length >= 7 && part.slice(0, 7).toLowerCase() === "domain=") {
            domain = part.slice(7).trim();
          }
        }
        var initiator = "";
        var stack = "";
        try { stack = (new Error()).stack || ""; } catch (e) { }
        var lines = stack.split("\n");
        for (var i = 0; i < lines.length; i++) {
          var match = lines[i].match(/https?:\/\/[^\s)]+/);
          if (!match) continue;
          initiator = match[0].replace(/:\d+(?::\d+)?$/, "");
          break;
        }
        var host = "";
        try { host = location.hostname || ""; } catch (e) { }
        if (typeof report === "function") {
          report(JSON.stringify({
            name: name,
            initiator: initiator,
            host: host,
            domain: cookieDomain(host, domain)
          }));
        }
      } catch (e) { }
      return accepted;
    }
  });
})(__PROBO_BINDING__);
