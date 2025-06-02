/**
 * @generated SignedSource<<6f2850042c0413963bd82f90236fe220>>
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
export type FrameworkDetailPageDetachDocumentMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlDocumentMappingInput;
};
export type FrameworkDetailPageDetachDocumentMutation$data = {
  readonly deleteControlDocumentMapping: {
    readonly deletedDocumentId: string;
  };
};
export type FrameworkDetailPageDetachDocumentMutation = {
  response: FrameworkDetailPageDetachDocumentMutation$data;
  variables: FrameworkDetailPageDetachDocumentMutation$variables;
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
    "name": "FrameworkDetailPageDetachDocumentMutation",
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
    "name": "FrameworkDetailPageDetachDocumentMutation",
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
    "cacheID": "53618baa1678e3a1b496a08597bdd364",
    "id": null,
    "metadata": {},
    "name": "FrameworkDetailPageDetachDocumentMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkDetailPageDetachDocumentMutation(\n  $input: DeleteControlDocumentMappingInput!\n) {\n  deleteControlDocumentMapping(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "233b182104744a67a21d21d60b6df67a";

export default node;
