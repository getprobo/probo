/**
 * @generated SignedSource<<0fefe3c54e67c68a8a84f46490503ad6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteMailingListUpdateInput = {
  id: string;
};
export type CompliancePageNewsListDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteMailingListUpdateInput;
};
export type CompliancePageNewsListDeleteMutation$data = {
  readonly deleteMailingListUpdate: {
    readonly deletedMailingListUpdateId: string;
  };
};
export type CompliancePageNewsListDeleteMutation = {
  response: CompliancePageNewsListDeleteMutation$data;
  variables: CompliancePageNewsListDeleteMutation$variables;
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
  "name": "deletedMailingListUpdateId",
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
    "name": "CompliancePageNewsListDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteMailingListUpdatePayload",
        "kind": "LinkedField",
        "name": "deleteMailingListUpdate",
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
    "name": "CompliancePageNewsListDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteMailingListUpdatePayload",
        "kind": "LinkedField",
        "name": "deleteMailingListUpdate",
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
            "name": "deletedMailingListUpdateId",
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
    "cacheID": "327bfffc5cb164c2ba13eca99a06f1c1",
    "id": null,
    "metadata": {},
    "name": "CompliancePageNewsListDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageNewsListDeleteMutation(\n  $input: DeleteMailingListUpdateInput!\n) {\n  deleteMailingListUpdate(input: $input) {\n    deletedMailingListUpdateId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d4913b991f7163a1adb1130fe04d591c";

export default node;
