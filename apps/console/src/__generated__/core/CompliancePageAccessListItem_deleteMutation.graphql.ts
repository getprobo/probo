/**
 * @generated SignedSource<<25c74b1d91584c785476f8eb07a6db6a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTrustCenterAccessInput = {
  id: string;
};
export type CompliancePageAccessListItem_deleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTrustCenterAccessInput;
};
export type CompliancePageAccessListItem_deleteMutation$data = {
  readonly deleteTrustCenterAccess: {
    readonly deletedTrustCenterAccessId: string;
  };
};
export type CompliancePageAccessListItem_deleteMutation = {
  response: CompliancePageAccessListItem_deleteMutation$data;
  variables: CompliancePageAccessListItem_deleteMutation$variables;
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
  "name": "deletedTrustCenterAccessId",
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
    "name": "CompliancePageAccessListItem_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterAccess",
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
    "name": "CompliancePageAccessListItem_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterAccess",
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
            "name": "deletedTrustCenterAccessId",
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
    "cacheID": "10030545f3bff35111fd9eae0c57d351",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAccessListItem_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageAccessListItem_deleteMutation(\n  $input: DeleteTrustCenterAccessInput!\n) {\n  deleteTrustCenterAccess(input: $input) {\n    deletedTrustCenterAccessId\n  }\n}\n"
  }
};
})();

(node as any).hash = "ed9591cbfdafb0ac1ea26b820a4dbf93";

export default node;
