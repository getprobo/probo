/**
 * @generated SignedSource<<0cb21a9122ca2763aaa5dbe4d33e2b4a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RemoveMemberInput = {
  membershipId: string;
  organizationId: string;
};
export type UserPage_removeMutation$variables = {
  input: RemoveMemberInput;
};
export type UserPage_removeMutation$data = {
  readonly removeMember: {
    readonly deletedMembershipId: string;
  } | null | undefined;
};
export type UserPage_removeMutation = {
  response: UserPage_removeMutation$data;
  variables: UserPage_removeMutation$variables;
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
    "concreteType": "RemoveMemberPayload",
    "kind": "LinkedField",
    "name": "removeMember",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedMembershipId",
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
    "name": "UserPage_removeMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "UserPage_removeMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "7aa8db55205ca16403a5919a5151d69f",
    "id": null,
    "metadata": {},
    "name": "UserPage_removeMutation",
    "operationKind": "mutation",
    "text": "mutation UserPage_removeMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    deletedMembershipId\n  }\n}\n"
  }
};
})();

(node as any).hash = "3c9804054834be923048845b796a2dfd";

export default node;
