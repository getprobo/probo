/**
 * @generated SignedSource<<b943f6ecf3e0a9056a9c1aaaa22baa57>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteComplianceBadgeInput = {
  id: string;
};
export type ComplianceBadgeGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteComplianceBadgeInput;
};
export type ComplianceBadgeGraphDeleteMutation$data = {
  readonly deleteComplianceBadge: {
    readonly deletedComplianceBadgeId: string;
  };
};
export type ComplianceBadgeGraphDeleteMutation = {
  response: ComplianceBadgeGraphDeleteMutation$data;
  variables: ComplianceBadgeGraphDeleteMutation$variables;
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
  "name": "deletedComplianceBadgeId",
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
    "name": "ComplianceBadgeGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteComplianceBadgePayload",
        "kind": "LinkedField",
        "name": "deleteComplianceBadge",
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
    "name": "ComplianceBadgeGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteComplianceBadgePayload",
        "kind": "LinkedField",
        "name": "deleteComplianceBadge",
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
            "name": "deletedComplianceBadgeId",
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
    "cacheID": "4322f24e86e98b11becaa2cc1eae5bc1",
    "id": null,
    "metadata": {},
    "name": "ComplianceBadgeGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceBadgeGraphDeleteMutation(\n  $input: DeleteComplianceBadgeInput!\n) {\n  deleteComplianceBadge(input: $input) {\n    deletedComplianceBadgeId\n  }\n}\n"
  }
};
})();

(node as any).hash = "3479376b84a75dc971434afa07002209";

export default node;
