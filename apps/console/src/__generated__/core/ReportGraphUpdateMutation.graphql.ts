/**
 * @generated SignedSource<<7db26b7c57759171fdcbc59e3e98fcff>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ReportState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateReportInput = {
  frameworkType?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  state?: ReportState | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
};
export type ReportGraphUpdateMutation$variables = {
  input: UpdateReportInput;
};
export type ReportGraphUpdateMutation$data = {
  readonly updateReport: {
    readonly report: {
      readonly file: {
        readonly fileName: string;
        readonly id: string;
      } | null | undefined;
      readonly framework: {
        readonly id: string;
        readonly name: string;
      };
      readonly frameworkType: string | null | undefined;
      readonly id: string;
      readonly name: string | null | undefined;
      readonly state: ReportState;
      readonly updatedAt: string;
      readonly validFrom: string | null | undefined;
      readonly validUntil: string | null | undefined;
    };
  };
};
export type ReportGraphUpdateMutation = {
  response: ReportGraphUpdateMutation$data;
  variables: ReportGraphUpdateMutation$variables;
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
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateReportPayload",
    "kind": "LinkedField",
    "name": "updateReport",
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
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "frameworkType",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "validFrom",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "validUntil",
            "storageKey": null
          },
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
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "state",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Framework",
            "kind": "LinkedField",
            "name": "framework",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              (v2/*: any*/)
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
    "name": "ReportGraphUpdateMutation",
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ReportGraphUpdateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "7998e23273e58a9d71798394cfadcc7e",
    "id": null,
    "metadata": {},
    "name": "ReportGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ReportGraphUpdateMutation(\n  $input: UpdateReportInput!\n) {\n  updateReport(input: $input) {\n    report {\n      id\n      name\n      frameworkType\n      validFrom\n      validUntil\n      file {\n        id\n        fileName\n      }\n      state\n      framework {\n        id\n        name\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e62412b2cded75dea3257dea5fa45665";

export default node;
