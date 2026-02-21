/**
 * @generated SignedSource<<c7153d76054821957556ac597dd4b7b9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED";
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
export type FormRiskDialogUpdateRiskMutation$variables = {
  input: UpdateRiskInput;
};
export type FormRiskDialogUpdateRiskMutation$data = {
  readonly updateRisk: {
    readonly risk: {
      readonly " $fragmentSpreads": FragmentRefs<"useRiskFormFragment">;
    };
  };
};
export type FormRiskDialogUpdateRiskMutation = {
  response: FormRiskDialogUpdateRiskMutation$data;
  variables: FormRiskDialogUpdateRiskMutation$variables;
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
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "FormRiskDialogUpdateRiskMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "useRiskFormFragment"
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FormRiskDialogUpdateRiskMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
              (v2/*: any*/),
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
                "name": "category",
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
                "name": "treatment",
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
                "name": "inherentRiskScore",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "residualRiskScore",
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
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "owner",
                "plural": false,
                "selections": [
                  (v2/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "59f586a59973ca023df027fdfb0d4bed",
    "id": null,
    "metadata": {},
    "name": "FormRiskDialogUpdateRiskMutation",
    "operationKind": "mutation",
    "text": "mutation FormRiskDialogUpdateRiskMutation(\n  $input: UpdateRiskInput!\n) {\n  updateRisk(input: $input) {\n    risk {\n      ...useRiskFormFragment\n      id\n    }\n  }\n}\n\nfragment useRiskFormFragment on Risk {\n  id\n  name\n  category\n  description\n  treatment\n  inherentLikelihood\n  inherentImpact\n  residualLikelihood\n  residualImpact\n  inherentRiskScore\n  residualRiskScore\n  note\n  owner {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "0299493042b2b5e7e464c991bab94375";

export default node;
