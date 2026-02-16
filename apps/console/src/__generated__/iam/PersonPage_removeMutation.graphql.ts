/**
 * @generated SignedSource<<7df320d2cbda9f48737622295ed141ad>>
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
export type PersonPage_removeMutation$variables = {
  input: RemoveUserInput;
};
export type PersonPage_removeMutation$data = {
  readonly removeUser: {
    readonly deletedProfileId: string;
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
    "concreteType": "RemoveUserPayload",
    "kind": "LinkedField",
    "name": "removeUser",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedProfileId",
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
    "cacheID": "1e686e9dcb1425cde221721b4635a747",
    "id": null,
    "metadata": {},
    "name": "PersonPage_removeMutation",
    "operationKind": "mutation",
    "text": "mutation PersonPage_removeMutation(\n  $input: RemoveUserInput!\n) {\n  removeUser(input: $input) {\n    deletedProfileId\n  }\n}\n"
  }
};
})();

(node as any).hash = "37b79997d660eee761e9380612739211";

export default node;
