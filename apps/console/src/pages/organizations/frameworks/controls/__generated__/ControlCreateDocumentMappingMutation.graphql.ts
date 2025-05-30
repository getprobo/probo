/**
 * @generated SignedSource<<50d6cda002b6548d568947de29ef3971>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateControlDocumentMappingInput = {
  controlId: string;
  documentId: string;
};
export type ControlCreateDocumentMappingMutation$variables = {
  input: CreateControlDocumentMappingInput;
};
export type ControlCreateDocumentMappingMutation$data = {
  readonly createControlDocumentMapping: {
    readonly controlEdge: {
      readonly node: {
        readonly id: string;
      };
    };
  };
};
export type ControlCreateDocumentMappingMutation = {
  response: ControlCreateDocumentMappingMutation$data;
  variables: ControlCreateDocumentMappingMutation$variables;
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
    "concreteType": "CreateControlDocumentMappingPayload",
    "kind": "LinkedField",
    "name": "createControlDocumentMapping",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ControlEdge",
        "kind": "LinkedField",
        "name": "controlEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Control",
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
    "name": "ControlCreateDocumentMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ControlCreateDocumentMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "44eba73b10ac45d65d97c96098260898",
    "id": null,
    "metadata": {},
    "name": "ControlCreateDocumentMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ControlCreateDocumentMappingMutation(\n  $input: CreateControlDocumentMappingInput!\n) {\n  createControlDocumentMapping(input: $input) {\n    controlEdge {\n      node {\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8ffd23b4cc3ce09756891fcde03c33cd";

export default node;
