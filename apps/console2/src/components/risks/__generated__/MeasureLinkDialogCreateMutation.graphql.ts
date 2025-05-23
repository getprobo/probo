/**
 * @generated SignedSource<<bce124e2ad364bd742c5d2b2a92f768b>>
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
export type MeasureLinkDialogCreateMutation$variables = {
  input: CreateRiskMeasureMappingInput;
};
export type MeasureLinkDialogCreateMutation$data = {
  readonly createRiskMeasureMapping: {
    readonly success: boolean;
  };
};
export type MeasureLinkDialogCreateMutation = {
  response: MeasureLinkDialogCreateMutation$data;
  variables: MeasureLinkDialogCreateMutation$variables;
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
    "name": "MeasureLinkDialogCreateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureLinkDialogCreateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "f7117ee68c2d7b6c2ceca56c054a4892",
    "id": null,
    "metadata": {},
    "name": "MeasureLinkDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureLinkDialogCreateMutation(\n  $input: CreateRiskMeasureMappingInput!\n) {\n  createRiskMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "b54abd7d4d4f88ff7d54bfa35121a015";

export default node;
