/**
 * @generated SignedSource<<2cf110e09c6e34e64c013c6fc260e918>>
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
export type ShowRiskViewDeleteRiskMeasureMappingMutation$variables = {
  input: DeleteRiskMeasureMappingInput;
};
export type ShowRiskViewDeleteRiskMeasureMappingMutation$data = {
  readonly deleteRiskMeasureMapping: {
    readonly success: boolean;
  };
};
export type ShowRiskViewDeleteRiskMeasureMappingMutation = {
  response: ShowRiskViewDeleteRiskMeasureMappingMutation$data;
  variables: ShowRiskViewDeleteRiskMeasureMappingMutation$variables;
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
    "concreteType": "DeleteRiskMeasureMappingPayload",
    "kind": "LinkedField",
    "name": "deleteRiskMeasureMapping",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "success",
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
    "name": "ShowRiskViewDeleteRiskMeasureMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ShowRiskViewDeleteRiskMeasureMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "76adf5d3867e6179bb11daeda4de9fc5",
    "id": null,
    "metadata": {},
    "name": "ShowRiskViewDeleteRiskMeasureMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ShowRiskViewDeleteRiskMeasureMappingMutation(\n  $input: DeleteRiskMeasureMappingInput!\n) {\n  deleteRiskMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "c51d0282a2d54bab88d56d9f2fdc9e49";

export default node;
