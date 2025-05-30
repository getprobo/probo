/**
 * @generated SignedSource<<70539298ec9a6ffd64b14fa98fdd283e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type CreateDraftDocumentVersionInput = {
  documentID: string;
};
export type ShowDocumentViewCreateDraftMutation$variables = {
  input: CreateDraftDocumentVersionInput;
};
export type ShowDocumentViewCreateDraftMutation$data = {
  readonly createDraftDocumentVersion: {
    readonly documentVersionEdge: {
      readonly node: {
        readonly id: string;
        readonly status: DocumentStatus;
        readonly version: number;
      };
    };
  };
};
export type ShowDocumentViewCreateDraftMutation = {
  response: ShowDocumentViewCreateDraftMutation$data;
  variables: ShowDocumentViewCreateDraftMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "CreateDraftDocumentVersionPayload",
    "kind": "LinkedField",
    "name": "createDraftDocumentVersion",
    "plural": false,
    "selections": [
      {
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
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "id",
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
                "name": "status",
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ShowDocumentViewCreateDraftMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ShowDocumentViewCreateDraftMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6ca94698731c59c1d5881a1dd3092277",
    "id": null,
    "metadata": {},
    "name": "ShowDocumentViewCreateDraftMutation",
    "operationKind": "mutation",
    "text": "mutation ShowDocumentViewCreateDraftMutation(\n  $input: CreateDraftDocumentVersionInput!\n) {\n  createDraftDocumentVersion(input: $input) {\n    documentVersionEdge {\n      node {\n        id\n        version\n        status\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "70edae5e8a303869b1686bb02fe4b22e";

export default node;
