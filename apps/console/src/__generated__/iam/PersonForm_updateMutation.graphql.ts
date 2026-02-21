/**
 * @generated SignedSource<<8563f3b122bb8d5b6c2e5e7354a0a0c8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type UpdateUserInput = {
  additionalEmailAddresses?: ReadonlyArray<string> | null | undefined;
  contractEndDate?: string | null | undefined;
  contractStartDate?: string | null | undefined;
  fullName: string;
  id: string;
  kind: ProfileKind;
  position?: string | null | undefined;
};
export type PersonForm_updateMutation$variables = {
  input: UpdateUserInput;
};
export type PersonForm_updateMutation$data = {
  readonly updateUser: {
    readonly profile: {
      readonly id: string;
    };
  };
};
export type PersonForm_updateMutation = {
  response: PersonForm_updateMutation$data;
  variables: PersonForm_updateMutation$variables;
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
    "concreteType": "UpdateUserPayload",
    "kind": "LinkedField",
    "name": "updateUser",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Profile",
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
    "name": "PersonForm_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PersonForm_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0623be6bdbd6575b04173a59b1dd45f7",
    "id": null,
    "metadata": {},
    "name": "PersonForm_updateMutation",
    "operationKind": "mutation",
    "text": "mutation PersonForm_updateMutation(\n  $input: UpdateUserInput!\n) {\n  updateUser(input: $input) {\n    profile {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e1da3d7148eef5f6ab8bdf827e41654a";

export default node;
