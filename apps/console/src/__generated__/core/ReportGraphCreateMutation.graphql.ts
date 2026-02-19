/**
 * @generated SignedSource<<74919d056ff682d2f6fd24c88f89da57>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ReportState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type CreateReportInput = {
  frameworkId: string;
  frameworkType?: string | null | undefined;
  name?: string | null | undefined;
  organizationId: string;
  state?: ReportState | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
};
export type ReportGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateReportInput;
};
export type ReportGraphCreateMutation$data = {
  readonly createReport: {
    readonly reportEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: string;
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
        readonly validFrom: string | null | undefined;
        readonly validUntil: string | null | undefined;
      };
    };
  };
};
export type ReportGraphCreateMutation = {
  response: ReportGraphCreateMutation$data;
  variables: ReportGraphCreateMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "ReportEdge",
  "kind": "LinkedField",
  "name": "reportEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Report",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
        (v4/*: any*/),
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
            (v3/*: any*/),
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
            (v3/*: any*/),
            (v4/*: any*/)
          ],
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:report:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:report:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:report:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:report:delete\")"
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "ReportGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateReportPayload",
        "kind": "LinkedField",
        "name": "createReport",
        "plural": false,
        "selections": [
          (v5/*: any*/)
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
    "name": "ReportGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateReportPayload",
        "kind": "LinkedField",
        "name": "createReport",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "reportEdge",
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
    "cacheID": "2f99bbade8d6da164ec5abdf1bf2abd9",
    "id": null,
    "metadata": {},
    "name": "ReportGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ReportGraphCreateMutation(\n  $input: CreateReportInput!\n) {\n  createReport(input: $input) {\n    reportEdge {\n      node {\n        id\n        name\n        frameworkType\n        validFrom\n        validUntil\n        file {\n          id\n          fileName\n        }\n        state\n        framework {\n          id\n          name\n        }\n        createdAt\n        canUpdate: permission(action: \"core:report:update\")\n        canDelete: permission(action: \"core:report:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6c85e7bfce90c88c8ec4eda2b93b6bc5";

export default node;
