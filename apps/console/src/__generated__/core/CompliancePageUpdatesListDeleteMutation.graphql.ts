/**
 * @generated SignedSource<<8fc725e69722e66e680df0ed52206c77>>
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
export type CompliancePageUpdatesListDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteMailingListUpdateInput;
};
export type CompliancePageUpdatesListDeleteMutation$data = {
  readonly deleteMailingListUpdate: {
    readonly deletedMailingListUpdateId: string;
  };
};
export type CompliancePageUpdatesListDeleteMutation = {
  response: CompliancePageUpdatesListDeleteMutation$data;
  variables: CompliancePageUpdatesListDeleteMutation$variables;
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
    "name": "CompliancePageUpdatesListDeleteMutation",
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
    "name": "CompliancePageUpdatesListDeleteMutation",
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
    "cacheID": "892a75687f4d8560945e5a30915aecbd",
    "id": null,
    "metadata": {},
    "name": "CompliancePageUpdatesListDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageUpdatesListDeleteMutation(\n  $input: DeleteMailingListUpdateInput!\n) {\n  deleteMailingListUpdate(input: $input) {\n    deletedMailingListUpdateId\n  }\n}\n"
  }
};
})();

(node as any).hash = "29a3142fa30bc64beb74d21f5162a00c";

export default node;
