/**
 * @generated SignedSource<<588b3953b36814ede0cd490eeaa6f99d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteDocumentInput = {
  documentId: string;
};
export type DocumentListItem_deleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteDocumentInput;
};
export type DocumentListItem_deleteMutation$data = {
  readonly deleteDocument: {
    readonly deletedDocumentId: string;
  };
};
export type DocumentListItem_deleteMutation = {
  response: DocumentListItem_deleteMutation$data;
  variables: DocumentListItem_deleteMutation$variables;
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
  "name": "deletedDocumentId",
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
    "name": "DocumentListItem_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteDocumentPayload",
        "kind": "LinkedField",
        "name": "deleteDocument",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "DocumentListItem_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteDocumentPayload",
        "kind": "LinkedField",
        "name": "deleteDocument",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedDocumentId",
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
    "cacheID": "de0d3cda546f944c8ae3a0ae98aa8bcb",
    "id": null,
    "metadata": {},
    "name": "DocumentListItem_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentListItem_deleteMutation(\n  $input: DeleteDocumentInput!\n) {\n  deleteDocument(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "91fbf2fc8563f4ffe8732193a815de77";

export default node;
