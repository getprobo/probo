/**
 * @generated SignedSource<<70ab61445e056dd6b6de3d0358ca4baf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UploadReportFileInput = {
  file: any;
  reportId: string;
};
export type ReportGraphUploadFileMutation$variables = {
  input: UploadReportFileInput;
};
export type ReportGraphUploadFileMutation$data = {
  readonly uploadReportFile: {
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
export type ReportGraphUploadFileMutation = {
  response: ReportGraphUploadFileMutation$data;
  variables: ReportGraphUploadFileMutation$variables;
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
    "concreteType": "UploadReportFilePayload",
    "kind": "LinkedField",
    "name": "uploadReportFile",
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
    "name": "ReportGraphUploadFileMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ReportGraphUploadFileMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "74219d0e9fb80ba594b3d0c891432e56",
    "id": null,
    "metadata": {},
    "name": "ReportGraphUploadFileMutation",
    "operationKind": "mutation",
    "text": "mutation ReportGraphUploadFileMutation(\n  $input: UploadReportFileInput!\n) {\n  uploadReportFile(input: $input) {\n    report {\n      id\n      file {\n        id\n        fileName\n        downloadUrl\n        createdAt\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "41c0e1b162df3ed72b37ae00b2204a58";

export default node;
