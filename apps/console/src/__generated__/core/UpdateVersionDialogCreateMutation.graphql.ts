/**
 * @generated SignedSource<<19fb3f18c8abfaf933f77514a1800538>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type CreateDraftDocumentVersionInput = {
  documentID: string;
};
export type UpdateVersionDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateDraftDocumentVersionInput;
};
export type UpdateVersionDialogCreateMutation$data = {
  readonly createDraftDocumentVersion: {
    readonly documentVersionEdge: {
      readonly node: {
        readonly content: string;
        readonly id: string;
        readonly publishedAt: string | null | undefined;
        readonly signatures: {
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly state: DocumentVersionSignatureState;
            };
          }>;
        };
        readonly status: DocumentStatus;
        readonly updatedAt: string;
        readonly version: number;
      };
    };
  };
};
export type UpdateVersionDialogCreateMutation = {
  response: UpdateVersionDialogCreateMutation$data;
  variables: UpdateVersionDialogCreateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "concreteType": "DocumentVersionEdge",
  "kind": "LinkedField",
  "name": "documentVersionEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "DocumentVersion",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "content",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "status",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "publishedAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "version",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "updatedAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": [
            {
              "kind": "Literal",
              "name": "first",
              "value": 100
            }
          ],
          "concreteType": "DocumentVersionSignatureConnection",
          "kind": "LinkedField",
          "name": "signatures",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "DocumentVersionSignatureEdge",
              "kind": "LinkedField",
              "name": "edges",
              "plural": true,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "DocumentVersionSignature",
                  "kind": "LinkedField",
                  "name": "node",
                  "plural": false,
                  "selections": [
                    (v3/*: any*/),
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "state",
                      "storageKey": null
                    }
                  ],
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": "signatures(first:100)"
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "UpdateVersionDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDraftDocumentVersionPayload",
        "kind": "LinkedField",
        "name": "createDraftDocumentVersion",
        "plural": false,
        "selections": [
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "UpdateVersionDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDraftDocumentVersionPayload",
        "kind": "LinkedField",
        "name": "createDraftDocumentVersion",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "documentVersionEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1a0f515ea5c0ca4fd3ab1ae969025ca0",
    "id": null,
    "metadata": {},
    "name": "UpdateVersionDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation UpdateVersionDialogCreateMutation(\n  $input: CreateDraftDocumentVersionInput!\n) {\n  createDraftDocumentVersion(input: $input) {\n    documentVersionEdge {\n      node {\n        id\n        content\n        status\n        publishedAt\n        version\n        updatedAt\n        signatures(first: 100) {\n          edges {\n            node {\n              id\n              state\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "a78a88558effc43ceab6746d8a4fea46";

export default node;
