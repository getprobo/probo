/**
 * @generated SignedSource<<64c2c92a2a0806f34243f8fc9904167c>>
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
export type DocumentGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteDocumentInput;
};
export type DocumentGraphDeleteMutation$data = {
  readonly deleteDocument: {
    readonly deletedDocumentId: string;
  };
};
export type DocumentGraphDeleteMutation = {
  response: DocumentGraphDeleteMutation$data;
  variables: DocumentGraphDeleteMutation$variables;
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
    "name": "DocumentGraphDeleteMutation",
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
    "name": "DocumentGraphDeleteMutation",
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
    "cacheID": "84c747138470346796393f93a85b7127",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentGraphDeleteMutation(\n  $input: DeleteDocumentInput!\n) {\n  deleteDocument(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "355f3a70caecabb2146075657665633e";

export default node;
