/**
 * @generated SignedSource<<0c55769185354921be98596538492fab>>
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
export type PersonPage_removeMutation$variables = {
  input: RemoveMemberInput;
};
export type PersonPage_removeMutation$data = {
  readonly removeMember: {
    readonly deletedMembershipId: string;
  } | null | undefined;
};
export type PersonPage_removeMutation = {
  response: PersonPage_removeMutation$data;
  variables: PersonPage_removeMutation$variables;
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
    "name": "PersonPage_removeMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PersonPage_removeMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0c220926bf64fda6954b3ab8e23dcd72",
    "id": null,
    "metadata": {},
    "name": "PersonPage_removeMutation",
    "operationKind": "mutation",
    "text": "mutation PersonPage_removeMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    deletedMembershipId\n  }\n}\n"
  }
};
})();

(node as any).hash = "60608384ce1ea37ec7b347ec4fdb4a61";

export default node;
