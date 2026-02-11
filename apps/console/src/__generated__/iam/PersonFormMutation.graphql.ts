/**
 * @generated SignedSource<<fa83e89c2c59e37853ab5efafb00287b>>
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
export type PersonFormMutation$variables = {
  input: UpdateProfileInput;
};
export type PersonFormMutation$data = {
  readonly updateProfile: {
    readonly profile: {
      readonly id: string;
    };
  };
};
export type PersonFormMutation = {
  response: PersonFormMutation$data;
  variables: PersonFormMutation$variables;
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
    "name": "PersonFormMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PersonFormMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "fe068a78bed066f0d50cab1dde6db292",
    "id": null,
    "metadata": {},
    "name": "PersonFormMutation",
    "operationKind": "mutation",
    "text": "mutation PersonFormMutation(\n  $input: UpdateProfileInput!\n) {\n  updateProfile(input: $input) {\n    profile {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "eb43fe5003e9761924aad474bf8296ee";

export default node;
