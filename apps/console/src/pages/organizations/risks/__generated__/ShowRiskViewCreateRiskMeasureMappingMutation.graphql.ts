/**
 * @generated SignedSource<<ac6b4af30ca3e03181512667a7ee6d76>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateRiskMeasureMappingInput = {
  measureId: string;
  riskId: string;
};
export type ShowRiskViewCreateRiskMeasureMappingMutation$variables = {
  input: CreateRiskMeasureMappingInput;
};
export type ShowRiskViewCreateRiskMeasureMappingMutation$data = {
  readonly createRiskMeasureMapping: {
    readonly success: boolean;
  };
};
export type ShowRiskViewCreateRiskMeasureMappingMutation = {
  response: ShowRiskViewCreateRiskMeasureMappingMutation$data;
  variables: ShowRiskViewCreateRiskMeasureMappingMutation$variables;
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
    "concreteType": "CreateRiskMeasureMappingPayload",
    "kind": "LinkedField",
    "name": "createRiskMeasureMapping",
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
    "name": "ShowRiskViewCreateRiskMeasureMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ShowRiskViewCreateRiskMeasureMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "f9e85548eb6e6aeb601fc36f92803ab6",
    "id": null,
    "metadata": {},
    "name": "ShowRiskViewCreateRiskMeasureMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ShowRiskViewCreateRiskMeasureMappingMutation(\n  $input: CreateRiskMeasureMappingInput!\n) {\n  createRiskMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "6246bbb8af3e7c1a2e968eb5482c91ce";

export default node;
