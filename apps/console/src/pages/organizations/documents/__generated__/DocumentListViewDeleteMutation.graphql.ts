/**
 * @generated SignedSource<<f20fbf3a0fd45a8f408e5fd364c9dfee>>
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
export type DocumentListViewDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteDocumentInput;
};
export type DocumentListViewDeleteMutation$data = {
  readonly deleteDocument: {
    readonly deletedDocumentId: string;
  };
};
export type DocumentListViewDeleteMutation = {
  response: DocumentListViewDeleteMutation$data;
  variables: DocumentListViewDeleteMutation$variables;
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
    "name": "DocumentListViewDeleteMutation",
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
    "name": "DocumentListViewDeleteMutation",
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
    "cacheID": "61bee9d714b5ee08e1653d0ad426c6f5",
    "id": null,
    "metadata": {},
    "name": "DocumentListViewDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentListViewDeleteMutation(\n  $input: DeleteDocumentInput!\n) {\n  deleteDocument(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "f0da5a5c332d17752d23c2fbfcd0e503";

export default node;
