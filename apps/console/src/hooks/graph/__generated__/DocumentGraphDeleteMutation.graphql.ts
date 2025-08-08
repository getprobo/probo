/**
 * @generated SignedSource<<6b7916e6ff46a629bf97a7b072db2d8b>>
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
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedDocumentId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "DeleteDocumentPayload",
        "kind": "LinkedField",
        "name": "deleteDocument",
        "plural": false,
        "selections": [
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "DeleteDocumentPayload",
        "kind": "LinkedField",
        "name": "deleteDocument",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteRecord",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedDocumentId"
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

(node as any).hash = "471020248ed9d398ab80655e04cb823d";

export default node;
