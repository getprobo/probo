/**
 * @generated SignedSource<<b78cfbea93f3bccdba56cd29900ab724>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateTrustCenterFileInput = {
  category?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
};
export type TrustCenterFileGraphUpdateMutation$variables = {
  input: UpdateTrustCenterFileInput;
};
export type TrustCenterFileGraphUpdateMutation$data = {
  readonly updateTrustCenterFile: {
    readonly trustCenterFile: {
      readonly category: string;
      readonly id: string;
      readonly name: string;
      readonly trustCenterVisibility: TrustCenterVisibility;
      readonly updatedAt: any;
    };
  };
};
export type TrustCenterFileGraphUpdateMutation = {
  response: TrustCenterFileGraphUpdateMutation$data;
  variables: TrustCenterFileGraphUpdateMutation$variables;
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
    "concreteType": "UpdateTrustCenterFilePayload",
    "kind": "LinkedField",
    "name": "updateTrustCenterFile",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterFile",
        "kind": "LinkedField",
        "name": "trustCenterFile",
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
            "name": "category",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "trustCenterVisibility",
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
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterFileGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterFileGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9ad761d91cfe34e7fa905a3f6a4d9e37",
    "id": null,
    "metadata": {},
    "name": "TrustCenterFileGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterFileGraphUpdateMutation(\n  $input: UpdateTrustCenterFileInput!\n) {\n  updateTrustCenterFile(input: $input) {\n    trustCenterFile {\n      id\n      name\n      category\n      trustCenterVisibility\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7306c3f9530a7636adc0b6a0cf28b67e";

export default node;
