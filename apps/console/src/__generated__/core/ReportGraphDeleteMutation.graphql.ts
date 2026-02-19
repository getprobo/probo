/**
 * @generated SignedSource<<dbf3340b784d0963acd186af3a465e60>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteReportInput = {
  reportId: string;
};
export type ReportGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteReportInput;
};
export type ReportGraphDeleteMutation$data = {
  readonly deleteReport: {
    readonly deletedReportId: string;
  };
};
export type ReportGraphDeleteMutation = {
  response: ReportGraphDeleteMutation$data;
  variables: ReportGraphDeleteMutation$variables;
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
  "name": "deletedReportId",
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
    "name": "ReportGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteReportPayload",
        "kind": "LinkedField",
        "name": "deleteReport",
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
    "name": "ReportGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteReportPayload",
        "kind": "LinkedField",
        "name": "deleteReport",
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
            "name": "deletedReportId",
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
    "cacheID": "2798475af3cb919500eb8c1a1169e477",
    "id": null,
    "metadata": {},
    "name": "ReportGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ReportGraphDeleteMutation(\n  $input: DeleteReportInput!\n) {\n  deleteReport(input: $input) {\n    deletedReportId\n  }\n}\n"
  }
};
})();

(node as any).hash = "11824f34a01d83d2d56b143f83c34aa0";

export default node;
