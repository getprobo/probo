/**
 * @generated SignedSource<<b6236fab1ef470add01b9d9cc7527caa>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type CreateAuditInput = {
  frameworkId: string;
  name?: string | null | undefined;
  organizationId: string;
  state?: AuditState | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
  validFrom?: any | null | undefined;
  validUntil?: any | null | undefined;
};
export type AuditGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateAuditInput;
};
export type AuditGraphCreateMutation$data = {
  readonly createAudit: {
    readonly auditEdge: {
      readonly node: {
        readonly createdAt: any;
        readonly framework: {
          readonly id: string;
          readonly name: string;
        };
        readonly id: string;
        readonly name: string | null | undefined;
        readonly report: {
          readonly filename: string;
          readonly id: string;
        } | null | undefined;
        readonly state: AuditState;
        readonly validFrom: any | null | undefined;
        readonly validUntil: any | null | undefined;
      };
    };
  };
};
export type AuditGraphCreateMutation = {
  response: AuditGraphCreateMutation$data;
  variables: AuditGraphCreateMutation$variables;
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
  "concreteType": "AuditEdge",
  "kind": "LinkedField",
  "name": "auditEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Audit",
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
          "concreteType": "Report",
          "kind": "LinkedField",
          "name": "report",
          "plural": false,
          "selections": [
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "filename",
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
    "name": "AuditGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateAuditPayload",
        "kind": "LinkedField",
        "name": "createAudit",
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
    "name": "AuditGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateAuditPayload",
        "kind": "LinkedField",
        "name": "createAudit",
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
            "name": "auditEdge",
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
    "cacheID": "04d44dfd419c2cde11b0138ca175203f",
    "id": null,
    "metadata": {},
    "name": "AuditGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphCreateMutation(\n  $input: CreateAuditInput!\n) {\n  createAudit(input: $input) {\n    auditEdge {\n      node {\n        id\n        name\n        validFrom\n        validUntil\n        report {\n          id\n          filename\n        }\n        state\n        framework {\n          id\n          name\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "4d014b16e02353df354085a151ccfe06";

export default node;
