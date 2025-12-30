/**
 * @generated SignedSource<<5c230fb025256fb7a5281ee5ecd3ccf4>>
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
export type DocumentControlsTab_detachControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlDocumentMappingInput;
};
export type DocumentControlsTab_detachControlMutation$data = {
  readonly deleteControlDocumentMapping: {
    readonly deletedControlId: string;
  };
};
export type DocumentControlsTab_detachControlMutation = {
  response: DocumentControlsTab_detachControlMutation$data;
  variables: DocumentControlsTab_detachControlMutation$variables;
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
    "name": "DocumentControlsTab_detachControlMutation",
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
    "name": "DocumentControlsTab_detachControlMutation",
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
    "cacheID": "fce801689ab578a47c33a50c60fc4cd7",
    "id": null,
    "metadata": {},
    "name": "DocumentControlsTab_detachControlMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentControlsTab_detachControlMutation(\n  $input: DeleteControlDocumentMappingInput!\n) {\n  deleteControlDocumentMapping(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "39c5b39c08ba37d06f5e0cfedc707a25";

export default node;
