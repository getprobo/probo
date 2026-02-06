/**
 * @generated SignedSource<<7a6d038560759279a60a293bc4be5bb8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type UpdateProfileInput = {
  additionalEmailAddresses?: ReadonlyArray<string> | null | undefined;
  contractEndDate?: string | null | undefined;
  contractStartDate?: string | null | undefined;
  fullName: string;
  id: string;
  kind: ProfileKind;
  position?: string | null | undefined;
};
export type UserFormMutation$variables = {
  input: UpdateProfileInput;
};
export type UserFormMutation$data = {
  readonly updateProfile: {
    readonly profile: {
      readonly id: string;
    };
  };
};
export type UserFormMutation = {
  response: UserFormMutation$data;
  variables: UserFormMutation$variables;
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
    "concreteType": "UpdateProfilePayload",
    "kind": "LinkedField",
    "name": "updateProfile",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "MembershipProfile",
        "kind": "LinkedField",
        "name": "profile",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
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
    "name": "UserFormMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "UserFormMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "edb3d0ac8f6a87857c470114ff6ad7f4",
    "id": null,
    "metadata": {},
    "name": "UserFormMutation",
    "operationKind": "mutation",
    "text": "mutation UserFormMutation(\n  $input: UpdateProfileInput!\n) {\n  updateProfile(input: $input) {\n    profile {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6ae1b8fd927611dd3b3b530ccd1de7d1";

export default node;
