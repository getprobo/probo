/**
 * @generated SignedSource<<78e5a6f1db510ffe28be99140539b234>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED" | "%future added value";
export type UpdateRiskInput = {
  category?: string | null | undefined;
  description?: string | null | undefined;
  id: string;
  inherentImpact?: number | null | undefined;
  inherentLikelihood?: number | null | undefined;
  name?: string | null | undefined;
  note?: string | null | undefined;
  ownerId?: string | null | undefined;
  residualImpact?: number | null | undefined;
  residualLikelihood?: number | null | undefined;
  treatment?: RiskTreatment | null | undefined;
};
export type RisksPageUpdateRiskMutation$variables = {
  input: UpdateRiskInput;
};
export type RisksPageUpdateRiskMutation$data = {
  readonly updateRisk: {
    readonly risk: {
      readonly category: string;
      readonly description: string;
      readonly id: string;
      readonly inherentImpact: number;
      readonly inherentLikelihood: number;
      readonly name: string;
      readonly note: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      } | null | undefined;
      readonly residualImpact: number;
      readonly residualLikelihood: number;
      readonly treatment: RiskTreatment;
      readonly updatedAt: any;
    };
  };
};
export type RisksPageUpdateRiskMutation = {
  response: RisksPageUpdateRiskMutation$data;
  variables: RisksPageUpdateRiskMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateRiskPayload",
    "kind": "LinkedField",
    "name": "updateRisk",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Risk",
        "kind": "LinkedField",
        "name": "risk",
        "plural": false,
        "selections": [
          (v1/*: any*/),
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
            "name": "note",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "People",
            "kind": "LinkedField",
            "name": "owner",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fullName",
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
    "name": "RisksPageUpdateRiskMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RisksPageUpdateRiskMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "a3c421bf079967f96acf2215e72cf8a7",
    "id": null,
    "metadata": {},
    "name": "RisksPageUpdateRiskMutation",
    "operationKind": "mutation",
    "text": "mutation RisksPageUpdateRiskMutation(\n  $input: UpdateRiskInput!\n) {\n  updateRisk(input: $input) {\n    risk {\n      id\n      name\n      description\n      category\n      inherentLikelihood\n      inherentImpact\n      residualLikelihood\n      residualImpact\n      treatment\n      note\n      updatedAt\n      owner {\n        id\n        fullName\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e3c8daa449e937742695adf8deb6dbe5";

export default node;
