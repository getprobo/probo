/**
 * @generated SignedSource<<49b9725d5c0f6c5e83d2ae80cfa9d88c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteReportFileInput = {
  reportId: string;
};
export type ReportGraphDeleteFileMutation$variables = {
  input: DeleteReportFileInput;
};
export type ReportGraphDeleteFileMutation$data = {
  readonly deleteReportFile: {
    readonly report: {
      readonly file: {
        readonly createdAt: string;
        readonly downloadUrl: string;
        readonly fileName: string;
        readonly id: string;
      } | null | undefined;
      readonly id: string;
      readonly updatedAt: string;
    };
  };
};
export type ReportGraphDeleteFileMutation = {
  response: ReportGraphDeleteFileMutation$data;
  variables: ReportGraphDeleteFileMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "DeleteReportFilePayload",
    "kind": "LinkedField",
    "name": "deleteReportFile",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Report",
        "kind": "LinkedField",
        "name": "report",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "File",
            "kind": "LinkedField",
            "name": "file",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fileName",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "downloadUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "createdAt",
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          }
        ],
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
    "name": "ReportGraphDeleteFileMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ReportGraphDeleteFileMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "4f2aed3dd1b78e768518b07a4d1235bf",
    "id": null,
    "metadata": {},
    "name": "ReportGraphDeleteFileMutation",
    "operationKind": "mutation",
    "text": "mutation ReportGraphDeleteFileMutation(\n  $input: DeleteReportFileInput!\n) {\n  deleteReportFile(input: $input) {\n    report {\n      id\n      file {\n        id\n        fileName\n        downloadUrl\n        createdAt\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "628d6d5f39c383f9a0c3c43d4cb11a0c";

export default node;
