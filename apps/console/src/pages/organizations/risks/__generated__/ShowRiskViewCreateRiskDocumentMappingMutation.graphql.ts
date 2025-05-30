/**
 * @generated SignedSource<<ef56acc9abd55ffea9aee1093be375bf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateRiskDocumentMappingInput = {
  documentId: string;
  riskId: string;
};
export type ShowRiskViewCreateRiskDocumentMappingMutation$variables = {
  input: CreateRiskDocumentMappingInput;
};
export type ShowRiskViewCreateRiskDocumentMappingMutation$data = {
  readonly createRiskDocumentMapping: {
    readonly riskEdge: {
      readonly node: {
        readonly id: string;
      };
    };
  };
};
export type ShowRiskViewCreateRiskDocumentMappingMutation = {
  response: ShowRiskViewCreateRiskDocumentMappingMutation$data;
  variables: ShowRiskViewCreateRiskDocumentMappingMutation$variables;
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
    "concreteType": "CreateRiskDocumentMappingPayload",
    "kind": "LinkedField",
    "name": "createRiskDocumentMapping",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "RiskEdge",
        "kind": "LinkedField",
        "name": "riskEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Risk",
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
    "name": "ShowRiskViewCreateRiskDocumentMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ShowRiskViewCreateRiskDocumentMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d48140a960f31c5ee40799780ad852c0",
    "id": null,
    "metadata": {},
    "name": "ShowRiskViewCreateRiskDocumentMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ShowRiskViewCreateRiskDocumentMappingMutation(\n  $input: CreateRiskDocumentMappingInput!\n) {\n  createRiskDocumentMapping(input: $input) {\n    riskEdge {\n      node {\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "cfd17c609ec34647f8ecafc1b07037ec";

export default node;
