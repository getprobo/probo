/**
 * @generated SignedSource<<2d67398a2054662f01cd08fbd3ba557a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteComplianceNewsInput = {
  id: string;
};
export type CompliancePageNewsListDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteComplianceNewsInput;
};
export type CompliancePageNewsListDeleteMutation$data = {
  readonly deleteComplianceNews: {
    readonly deletedComplianceNewsId: string;
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
  "name": "deletedComplianceNewsId",
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
        "concreteType": "DeleteComplianceNewsPayload",
        "kind": "LinkedField",
        "name": "deleteComplianceNews",
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
        "concreteType": "DeleteComplianceNewsPayload",
        "kind": "LinkedField",
        "name": "deleteComplianceNews",
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
            "name": "deletedComplianceNewsId",
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
    "cacheID": "a22b53be0cd104817fa3d80815aecf83",
    "id": null,
    "metadata": {},
    "name": "CompliancePageNewsListDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageNewsListDeleteMutation(\n  $input: DeleteComplianceNewsInput!\n) {\n  deleteComplianceNews(input: $input) {\n    deletedComplianceNewsId\n  }\n}\n"
  }
};
})();

(node as any).hash = "baf4ea185775b5de1272972703d3b7a8";

export default node;
