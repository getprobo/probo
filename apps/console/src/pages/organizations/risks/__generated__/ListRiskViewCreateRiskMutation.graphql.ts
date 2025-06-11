/**
 * @generated SignedSource<<d41cd39c7d3f9a65977723198fbaeae6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED";
export type CreateRiskInput = {
  category: string;
  description: string;
  inherentImpact: number;
  inherentLikelihood: number;
  name: string;
  note?: string | null | undefined;
  organizationId: string;
  ownerId?: string | null | undefined;
  residualImpact?: number | null | undefined;
  residualLikelihood?: number | null | undefined;
  treatment: RiskTreatment;
};
export type ListRiskViewCreateRiskMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateRiskInput;
};
export type ListRiskViewCreateRiskMutation$data = {
  readonly createRisk: {
    readonly riskEdge: {
      readonly node: {
        readonly category: string;
        readonly createdAt: string;
        readonly description: string;
        readonly id: string;
        readonly inherentImpact: number;
        readonly inherentLikelihood: number;
        readonly name: string;
        readonly residualImpact: number;
        readonly residualLikelihood: number;
        readonly treatment: RiskTreatment;
        readonly updatedAt: string;
      };
    };
  };
};
export type ListRiskViewCreateRiskMutation = {
  response: ListRiskViewCreateRiskMutation$data;
  variables: ListRiskViewCreateRiskMutation$variables;
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
          "name": "inherentLikelihood",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "inherentImpact",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "residualLikelihood",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "residualImpact",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "treatment",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "updatedAt",
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
    "name": "ListRiskViewCreateRiskMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRiskPayload",
        "kind": "LinkedField",
        "name": "createRisk",
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
    "name": "ListRiskViewCreateRiskMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRiskPayload",
        "kind": "LinkedField",
        "name": "createRisk",
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
    "cacheID": "bf56f8b069199fbcc6767f72fb91461b",
    "id": null,
    "metadata": {},
    "name": "ListRiskViewCreateRiskMutation",
    "operationKind": "mutation",
    "text": "mutation ListRiskViewCreateRiskMutation(\n  $input: CreateRiskInput!\n) {\n  createRisk(input: $input) {\n    riskEdge {\n      node {\n        id\n        name\n        description\n        category\n        inherentLikelihood\n        inherentImpact\n        residualLikelihood\n        residualImpact\n        treatment\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c3e47e5c20b1dbe61dffd5e90842f337";

export default node;
