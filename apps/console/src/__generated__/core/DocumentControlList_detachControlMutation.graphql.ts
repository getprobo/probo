/**
 * @generated SignedSource<<25dfecd5fa953558409d2a4d1daa32c3>>
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
export type DocumentControlList_detachControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlDocumentMappingInput;
};
export type DocumentControlList_detachControlMutation$data = {
  readonly deleteControlDocumentMapping: {
    readonly deletedControlId: string;
  };
};
export type DocumentControlList_detachControlMutation = {
  response: DocumentControlList_detachControlMutation$data;
  variables: DocumentControlList_detachControlMutation$variables;
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
  "name": "deletedControlId",
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
    "name": "DocumentControlList_detachControlMutation",
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
    "name": "DocumentControlList_detachControlMutation",
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
            "name": "deletedControlId",
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
    "cacheID": "5d78356c860e344801d99117b599665d",
    "id": null,
    "metadata": {},
    "name": "DocumentControlList_detachControlMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentControlList_detachControlMutation(\n  $input: DeleteControlDocumentMappingInput!\n) {\n  deleteControlDocumentMapping(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "fc76faa3002641acaf89856a94d37942";

export default node;
