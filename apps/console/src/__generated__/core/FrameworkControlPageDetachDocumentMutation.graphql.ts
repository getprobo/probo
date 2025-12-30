/**
 * @generated SignedSource<<61b6714faa71bf0afe54a1984b719a55>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlDocumentMappingInput = {
  controlId: string;
  documentId: string;
};
export type FrameworkControlPageDetachDocumentMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlDocumentMappingInput;
};
export type FrameworkControlPageDetachDocumentMutation$data = {
  readonly deleteControlDocumentMapping: {
    readonly deletedDocumentId: string;
  };
};
export type FrameworkControlPageDetachDocumentMutation = {
  response: FrameworkControlPageDetachDocumentMutation$data;
  variables: FrameworkControlPageDetachDocumentMutation$variables;
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
    "name": "FrameworkControlPageDetachDocumentMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlDocumentMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlDocumentMapping",
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
    "name": "FrameworkControlPageDetachDocumentMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlDocumentMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlDocumentMapping",
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
    "cacheID": "f442693991b055c11609fb7bc326f1da",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachDocumentMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachDocumentMutation(\n  $input: DeleteControlDocumentMappingInput!\n) {\n  deleteControlDocumentMapping(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "8c20b51090f69e932bc4801d88a6eaa6";

export default node;
