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
  default:
    throw new Error(`unsupported action: ${request.action}`);
}
