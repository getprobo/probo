/**
 * @generated SignedSource<<f9c0e3868814f7247aee411c93d7a5eb>>
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
export type MeasureRisksTabDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskMeasureMappingInput;
};
export type MeasureRisksTabDetachMutation$data = {
  readonly deleteRiskMeasureMapping: {
    readonly deletedRiskId: string;
  };
};
export type MeasureRisksTabDetachMutation = {
  response: MeasureRisksTabDetachMutation$data;
  variables: MeasureRisksTabDetachMutation$variables;
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
  "name": "deletedRiskId",
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
    "name": "MeasureRisksTabDetachMutation",
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
    "name": "MeasureRisksTabDetachMutation",
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
            "name": "deletedRiskId",
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
    "cacheID": "ac75591b5e166b2ee994bd99a0afc29a",
    "id": null,
    "metadata": {},
    "name": "MeasureRisksTabDetachMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureRisksTabDetachMutation(\n  $input: DeleteRiskMeasureMappingInput!\n) {\n  deleteRiskMeasureMapping(input: $input) {\n    deletedRiskId\n  }\n}\n"
  }
};
})();

(node as any).hash = "5b42a2d7e8a78ff4c674ce6612df537c";

export default node;
