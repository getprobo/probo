/**
 * @generated SignedSource<<327cdb825c0029c21d8812b782ecf14f>>
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
export type MeasureLinkDialogDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskMeasureMappingInput;
};
export type MeasureLinkDialogDetachMutation$data = {
  readonly deleteRiskMeasureMapping: {
    readonly deletedMeasureId: string;
  };
};
export type MeasureLinkDialogDetachMutation = {
  response: MeasureLinkDialogDetachMutation$data;
  variables: MeasureLinkDialogDetachMutation$variables;
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
    "name": "MeasureLinkDialogDetachMutation",
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
    "name": "MeasureLinkDialogDetachMutation",
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
    "cacheID": "8a54f580a892f60f24c65ec927c7340c",
    "id": null,
    "metadata": {},
    "name": "MeasureLinkDialogDetachMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureLinkDialogDetachMutation(\n  $input: DeleteRiskMeasureMappingInput!\n) {\n  deleteRiskMeasureMapping(input: $input) {\n    deletedMeasureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "e02a17e00e2c07881dcae17fef9513eb";

export default node;
