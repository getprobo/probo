/**
 * @generated SignedSource<<54ac79c4d292eb19f73bbf2363015dd9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateTrustCenterAccessInput = {
  accessId: string;
  active?: boolean | null | undefined;
  email?: string | null | undefined;
  name?: string | null | undefined;
  sendEmail?: boolean;
};
export type TrustCenterAccessGraphUpdateMutation$variables = {
  input: UpdateTrustCenterAccessInput;
};
export type TrustCenterAccessGraphUpdateMutation$data = {
  readonly updateTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly active: boolean;
      readonly createdAt: any;
      readonly email: string;
      readonly id: string;
      readonly name: string;
    };
  };
};
export type TrustCenterAccessGraphUpdateMutation = {
  response: TrustCenterAccessGraphUpdateMutation$data;
  variables: TrustCenterAccessGraphUpdateMutation$variables;
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
    "concreteType": "UpdateTrustCenterAccessPayload",
    "kind": "LinkedField",
    "name": "updateTrustCenterAccess",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
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
            "name": "email",
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
            "name": "active",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
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
    "name": "TrustCenterAccessGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterAccessGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8e284c8129de02d689385cae38cc3876",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphUpdateMutation(\n  $input: UpdateTrustCenterAccessInput!\n) {\n  updateTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      email\n      name\n      active\n      createdAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "235754d40be56785bd0e7a36ac8c4ec2";

export default node;
