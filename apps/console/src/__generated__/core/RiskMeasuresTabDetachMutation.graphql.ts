/**
 * @generated SignedSource<<c34aab708f8713173639ae4b7aacdf10>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteRiskMeasureMappingInput = {
  measureId: string;
  riskId: string;
};
export type RiskMeasuresTabDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskMeasureMappingInput;
};
export type RiskMeasuresTabDetachMutation$data = {
  readonly deleteRiskMeasureMapping: {
    readonly deletedMeasureId: string;
  };
};
export type RiskMeasuresTabDetachMutation = {
  response: RiskMeasuresTabDetachMutation$data;
  variables: RiskMeasuresTabDetachMutation$variables;
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
  "name": "deletedMeasureId",
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
    "name": "RiskMeasuresTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "deleteRiskMeasureMapping",
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
    "name": "RiskMeasuresTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "deleteRiskMeasureMapping",
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
            "name": "deletedMeasureId",
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
    "cacheID": "81260f49b2f31c4a12f87b38bab7a683",
    "id": null,
    "metadata": {},
    "name": "RiskMeasuresTabDetachMutation",
    "operationKind": "mutation",
    "text": "mutation RiskMeasuresTabDetachMutation(\n  $input: DeleteRiskMeasureMappingInput!\n) {\n  deleteRiskMeasureMapping(input: $input) {\n    deletedMeasureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c5adc6d900f0da15ec3acf7db304d410";

export default node;
