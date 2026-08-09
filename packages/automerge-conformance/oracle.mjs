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

import * as Automerge from "@automerge/automerge";
import { Buffer } from "node:buffer";
import process from "node:process";

const chunks = [];
for await (const chunk of process.stdin) {
  chunks.push(chunk);
}

const request = JSON.parse(Buffer.concat(chunks).toString("utf8"));

switch (request.action) {
  case "runScenario": {
    const scenario = request.scenario;
    let document = Automerge.init({ actor: scenario.actor });
    let pending = [];
    for (const operation of scenario.operations) {
      if (operation.action !== "commit") {
        pending.push(operation);
        continue;
      }
      document = Automerge.change(
        document,
        {
          message: operation.message,
          time: operation.timestamp,
        },
        draft => {
          for (const mutation of pending) {
            applyScenarioMutation(draft, mutation);
          }
        },
      );
      pending = [];
    }
    if (pending.length !== 0) {
      throw new Error("scenario contains uncommitted operations");
    }
    process.stdout.write(
      JSON.stringify({
        data: normalize(document),
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "inspectScenario": {
    const document = Automerge.load(Buffer.from(request.document, "base64"));
    process.stdout.write(
      JSON.stringify({
        data: normalize(document),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "create": {
    let document = Automerge.init({ actor: request.actor });
    document = Automerge.change(
      document,
      {
        message: request.message,
        time: request.timestamp,
      },
      draft => {
        draft.body = "";
        Automerge.splice(draft, ["body"], 0, 0, request.text);
      },
    );
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createRichText": {
    let document = Automerge.from({ body: "" }, { actor: request.actor });
    document = Automerge.change(document, draft => {
      Automerge.updateSpans(draft, ["body"], [
        {
          type: "block",
          value: {
            type: "heading",
            parents: [],
            isEmbed: false,
            attrs: { level: 2 },
          },
        },
        {
          type: "text",
          value: "Policy",
          marks: { strong: true },
        },
      ]);
    });
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createChange": {
    let document = Automerge.init({ actor: request.actor });
    document = Automerge.change(
      document,
      {
        message: request.message,
        time: request.timestamp,
      },
      draft => {
        draft.title = "Policy";
      },
    );
    const changes = Automerge.getAllChanges(document);
    process.stdout.write(
      JSON.stringify({
        change: Buffer.from(changes[0]).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createEmptyChange": {
    let document = Automerge.init({ actor: request.actor });
    document = Automerge.emptyChange(document, {
      message: request.message,
      time: request.timestamp,
    });
    const changes = Automerge.getAllChanges(document);
    process.stdout.write(
      JSON.stringify({
        change: Buffer.from(changes[0]).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createConcurrentChanges": {
    const base = Automerge.from({ body: "" }, { actor: request.actor });
    let left = Automerge.load(Automerge.save(base), { actor: request.actorB });
    let right = Automerge.load(Automerge.save(base), { actor: request.actorC });
    left = Automerge.change(left, draft => {
      Automerge.splice(draft, ["body"], 0, 0, "L");
    });
    right = Automerge.change(right, draft => {
      Automerge.splice(draft, ["body"], 0, 0, "R");
    });
    const changes = [
      ...Automerge.getAllChanges(base),
      ...Automerge.getChanges(base, left),
      ...Automerge.getChanges(base, right),
    ];
    const merged = Automerge.merge(left, right);
    process.stdout.write(
      JSON.stringify({
        body: merged.body,
        changes: changes.map(change => Buffer.from(change).toString("base64")),
        heads: Automerge.getHeads(merged),
      }),
    );
    break;
  }
  case "createSyncMessage": {
    const document = Automerge.from(
      { title: "Policy" },
      { actor: request.actor },
    );
    const [, message] = Automerge.generateSyncMessage(
      document,
      Automerge.initSyncState(),
    );
    if (!message) throw new Error("expected an initial sync message");
    process.stdout.write(
      JSON.stringify({
        sync: Buffer.from(message).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "inspectChange": {
    const change = Buffer.from(request.change, "base64");
    const decoded = Automerge.decodeChange(change);
    const [document] = Automerge.applyChanges(Automerge.init(), [change]);
    process.stdout.write(
      JSON.stringify({
        body: document.title,
        data: normalize(decoded),
        message: decoded.message,
        heads: [decoded.hash],
      }),
    );
    break;
  }
  case "createComplexRichText": {
    let document = Automerge.from({ body: "" }, { actor: request.actor });
    document = Automerge.change(document, draft => {
      Automerge.updateSpans(draft, ["body"], [
        {
          type: "block",
          value: {
            type: "paragraph",
            parents: [],
            isEmbed: false,
            attrs: {},
          },
        },
        {
          type: "text",
          value: "A",
          marks: { strong: true, em: true },
        },
        {
          type: "text",
          value: "B",
          marks: { em: true },
        },
      ]);
    });
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createBoundaryMarks": {
    let document = Automerge.from({ body: "ABC" }, { actor: request.actor });
    document = Automerge.change(document, draft => {
      Automerge.mark(
        draft,
        ["body"],
        { start: 0, end: 1, expand: "none" },
        "strong",
        true,
      );
      Automerge.mark(
        draft,
        ["body"],
        { start: 1, end: 3, expand: "both" },
        "em",
        true,
      );
    });
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createSplitMarks": {
    let document = Automerge.from({ body: "ABCD" }, { actor: request.actor });
    document = Automerge.change(document, draft => {
      Automerge.mark(
        draft,
        ["body"],
        { start: 0, end: 4, expand: "both" },
        "strong",
        true,
      );
      Automerge.unmark(
        draft,
        ["body"],
        { start: 1, end: 3, expand: "none" },
        "strong",
      );
    });
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createUnicodeMarks": {
    let document = Automerge.from({ body: "😀😀" }, { actor: request.actor });
    document = Automerge.change(document, draft => {
      Automerge.mark(
        draft,
        ["body"],
        { start: 2, end: 4, expand: "none" },
        "strong",
        true,
      );
      Automerge.splice(draft, ["body"], 0, 0, "🙃");
    });
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createTableRichText": {
    let document = Automerge.from({ body: "" }, { actor: request.actor });
    document = Automerge.change(document, draft => {
      Automerge.updateSpans(draft, ["body"], [
        tableBlock("table", []),
        tableBlock("table-row", ["table"]),
        tableBlock("table-header", ["table", "table-row"], {
          colspan: 1,
          rowspan: 1,
          colwidth: null,
        }),
        { type: "text", value: "A" },
        tableBlock("table-cell", ["table", "table-row"], {
          colspan: 1,
          rowspan: 1,
          colwidth: null,
        }),
        { type: "text", value: "B" },
        tableBlock("table-row", ["table"]),
        tableBlock("table-cell", ["table", "table-row"], {
          colspan: 1,
          rowspan: 1,
          colwidth: null,
        }),
        { type: "text", value: "C" },
      ]);
    });
    process.stdout.write(
      JSON.stringify({
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "createDataModel": {
    let document = Automerge.init({ actor: request.actor });
    document = Automerge.change(document, draft => {
      draft.values = {
        string: "value",
        integer: -42,
        float: 3.25,
        true: true,
        false: false,
        null: null,
        bytes: new Uint8Array([0, 1, 254, 255]),
        timestamp: new Date("2026-08-08T00:00:00.000Z"),
        counter: new Automerge.Counter(5),
      };
      draft.list = ["first", 2, true, null];
      draft.text = "";
      Automerge.splice(draft, ["text"], 0, 0, "A😀B");
    });
    process.stdout.write(
      JSON.stringify({
        data: normalize(document),
        document: Buffer.from(Automerge.save(document)).toString("base64"),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "inspectDataModel": {
    const document = Automerge.load(Buffer.from(request.document, "base64"));
    process.stdout.write(
      JSON.stringify({
        data: normalize(document),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "inspect": {
    const document = Automerge.load(Buffer.from(request.document, "base64"));
    process.stdout.write(
      JSON.stringify({
        body: document.body,
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  case "inspectChanges": {
    const document = Automerge.load(Buffer.from(request.document, "base64"));
    process.stdout.write(
      JSON.stringify({
        changes: Automerge.getAllChanges(document).map(change =>
          Buffer.from(change).toString("base64")
        ),
        heads: Automerge.getHeads(document),
      }),
    );
    break;
  }
  default:
    throw new Error(`unsupported action: ${request.action}`);
}

function tableBlock(type, parents, attrs = {}) {
  return {
    type: "block",
    value: {
      type,
      parents,
      attrs,
      isEmbed: false,
    },
  };
}

function normalize(value) {
  if (typeof value === "bigint") {
    return { type: "bigint", value: value.toString() };
  }
  if (value instanceof Uint8Array) {
    return { type: "bytes", value: Buffer.from(value).toString("base64") };
  }
  if (value instanceof Date) {
    return { type: "timestamp", value: value.toISOString() };
  }
  if (value instanceof Automerge.Counter) {
    return { type: "counter", value: value.value };
  }
  if (Automerge.isImmutableString(value)) {
    return value.val;
  }
  if (Array.isArray(value)) {
    return value.map(normalize);
  }
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.keys(value)
        .sort()
        .map(key => [key, normalize(value[key])]),
    );
  }
  return value;
}

function applyScenarioMutation(document, operation) {
  const parent = valueAtPath(document, operation.path);
  switch (operation.action) {
    case "createObject":
      parent[operation.key] = scenarioObject(operation.objectType);
      return;
    case "putScalar":
      parent[operation.key] = scenarioScalar(operation.scalar);
      return;
    case "insertScalar":
      parent.splice(operation.index, 0, scenarioScalar(operation.scalar));
      return;
    case "putScalarAt":
      parent[operation.index] = scenarioScalar(operation.scalar);
      return;
    case "deleteIndex":
      parent.splice(operation.index, 1);
      return;
    case "createText":
      parent[operation.key] = "";
      return;
    case "spliceText":
      Automerge.splice(
        document,
        operation.path,
        operation.index,
        operation.deleteCount,
        operation.text,
      );
      return;
    case "increment":
      parent[operation.key].increment(operation.delta);
      return;
    default:
      throw new Error(`unsupported scenario mutation: ${operation.action}`);
  }
}

function valueAtPath(document, path) {
  let value = document;
  for (const property of path) {
    value = value[property];
  }
  return value;
}

function scenarioObject(type) {
  switch (type) {
    case "map":
      return {};
    case "list":
      return [];
    default:
      throw new Error(`unsupported scenario object type: ${type}`);
  }
}

function scenarioScalar(scalar) {
  switch (scalar.type) {
    case "null":
      return null;
    case "boolean":
      return scalar.bool;
    case "uint":
      return new Automerge.Uint(scalar.uint);
    case "int":
      return new Automerge.Int(scalar.int);
    case "float64":
      return new Automerge.Float64(floatFromBits(scalar.floatBits));
    case "string":
      return new Automerge.ImmutableString(scalar.string);
    case "bytes":
      return Uint8Array.from(Buffer.from(scalar.bytes, "hex"));
    case "timestamp":
      return new Date(scalar.int);
    case "counter":
      return new Automerge.Counter(scalar.int);
    default:
      throw new Error(`unsupported scenario scalar type: ${scalar.type}`);
  }
}

function floatFromBits(value) {
  const data = new DataView(new ArrayBuffer(8));
  data.setBigUint64(0, BigInt(value), false);
  return data.getFloat64(0, false);
}
