/**
 * @generated SignedSource<<921243e6b4b9007688051434ab212f85>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RemoveUserInput = {
  organizationId: string;
  profileId: string;
};
export type PeopleListItem_removeMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RemoveUserInput;
};
export type PeopleListItem_removeMutation$data = {
  readonly removeUser: {
    readonly deletedProfileId: string;
  } | null | undefined;
};
export type PeopleListItem_removeMutation = {
  response: PeopleListItem_removeMutation$data;
  variables: PeopleListItem_removeMutation$variables;
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
  "kind": "ScalarField",
  "name": "deletedProfileId",
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
    "name": "PeopleListItem_removeMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RemoveUserPayload",
        "kind": "LinkedField",
        "name": "removeUser",
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
    "name": "PeopleListItem_removeMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RemoveUserPayload",
        "kind": "LinkedField",
        "name": "removeUser",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedProfileId",
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
    "cacheID": "ec2aeab79c24169e2f8fc95ca3b28a7d",
    "id": null,
    "metadata": {},
    "name": "PeopleListItem_removeMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleListItem_removeMutation(\n  $input: RemoveUserInput!\n) {\n  removeUser(input: $input) {\n    deletedProfileId\n  }\n}\n"
  }
};
})();

(node as any).hash = "e84b22e1f4a71940947b8a85d83a2dd6";

export default node;
