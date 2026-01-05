/**
 * @generated SignedSource<<9af67842ec7b635682d0c10d86385b88>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateRiskMeasureMappingInput = {
  measureId: string;
  riskId: string;
};
export type MeasureRisksTabCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateRiskMeasureMappingInput;
};
export type MeasureRisksTabCreateMutation$data = {
  readonly createRiskMeasureMapping: {
    readonly riskEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedRisksCardFragment">;
      };
    };
  };
};
export type MeasureRisksTabCreateMutation = {
  response: MeasureRisksTabCreateMutation$data;
  variables: MeasureRisksTabCreateMutation$variables;
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
  "name": "id",
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
    "name": "MeasureRisksTabCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRiskMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "createRiskMeasureMapping",
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
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedRisksCardFragment"
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
    "name": "MeasureRisksTabCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRiskMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "createRiskMeasureMapping",
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
                  (v3/*: any*/),
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
                    "name": "inherentRiskScore",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "residualRiskScore",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "riskEdge",
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
    "cacheID": "53097378652dfdfc8505535921234ade",
    "id": null,
    "metadata": {},
    "name": "MeasureRisksTabCreateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureRisksTabCreateMutation(\n  $input: CreateRiskMeasureMappingInput!\n) {\n  createRiskMeasureMapping(input: $input) {\n    riskEdge {\n      node {\n        id\n        ...LinkedRisksCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedRisksCardFragment on Risk {\n  id\n  name\n  inherentRiskScore\n  residualRiskScore\n}\n"
  }
};
})();

(node as any).hash = "c167b300964866d886760968abaa014f";

export default node;
