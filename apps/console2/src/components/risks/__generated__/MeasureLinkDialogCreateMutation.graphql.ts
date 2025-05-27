/**
 * @generated SignedSource<<043e8cc5beda1ac8949a7c02adcff4bf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type CreateRiskMeasureMappingInput = {
  measureId: string;
  riskId: string;
};
export type MeasureLinkDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateRiskMeasureMappingInput;
};
export type MeasureLinkDialogCreateMutation$data = {
  readonly createRiskMeasureMapping: {
    readonly measureEdge: {
      readonly node: {
        readonly category: string;
        readonly description: string;
        readonly id: string;
        readonly name: string;
        readonly state: MeasureState;
      };
    };
  };
};
export type MeasureLinkDialogCreateMutation = {
  response: MeasureLinkDialogCreateMutation$data;
  variables: MeasureLinkDialogCreateMutation$variables;
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
  "concreteType": "MeasureEdge",
  "kind": "LinkedField",
  "name": "measureEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Measure",
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
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "name",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "description",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "category",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "state",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "MeasureLinkDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRiskMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "createRiskMeasureMapping",
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
    "name": "MeasureLinkDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRiskMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "createRiskMeasureMapping",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "measureEdge",
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
    "cacheID": "309f3c90d91577e2f1c95ea69848ce5a",
    "id": null,
    "metadata": {},
    "name": "MeasureLinkDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureLinkDialogCreateMutation(\n  $input: CreateRiskMeasureMappingInput!\n) {\n  createRiskMeasureMapping(input: $input) {\n    measureEdge {\n      node {\n        id\n        name\n        description\n        category\n        state\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bd012f06e39d5a04f4c1a2d2026f497e";

export default node;
