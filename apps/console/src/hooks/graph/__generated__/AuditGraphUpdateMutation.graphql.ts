/**
 * @generated SignedSource<<d2c542541d5bfda5464af2b34ccae603>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type UpdateAuditInput = {
  id: string;
  showOnTrustCenter?: boolean | null | undefined;
  state?: AuditState | null | undefined;
  validFrom?: any | null | undefined;
  validUntil?: any | null | undefined;
};
export type AuditGraphUpdateMutation$variables = {
  input: UpdateAuditInput;
};
export type AuditGraphUpdateMutation$data = {
  readonly updateAudit: {
    readonly audit: {
      readonly framework: {
        readonly id: string;
        readonly name: string;
      };
      readonly id: string;
      readonly report: {
        readonly filename: string;
        readonly id: string;
      } | null | undefined;
      readonly state: AuditState;
      readonly updatedAt: any;
      readonly validFrom: any | null | undefined;
      readonly validUntil: any | null | undefined;
    };
  };
};
export type AuditGraphUpdateMutation = {
  response: AuditGraphUpdateMutation$data;
  variables: AuditGraphUpdateMutation$variables;
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
    "concreteType": "UpdateAuditPayload",
    "kind": "LinkedField",
    "name": "updateAudit",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Audit",
        "kind": "LinkedField",
        "name": "audit",
        "plural": false,
        "selections": [
          (v1/*: any*/),
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
              (v1/*: any*/),
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
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "name",
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
    "name": "AuditGraphUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AuditGraphUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "afd98ba771c6c437c46c275655333eb6",
    "id": null,
    "metadata": {},
    "name": "AuditGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphUpdateMutation(\n  $input: UpdateAuditInput!\n) {\n  updateAudit(input: $input) {\n    audit {\n      id\n      validFrom\n      validUntil\n      report {\n        id\n        filename\n      }\n      state\n      framework {\n        id\n        name\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "3ddd17832768611675505d76da1839a1";

export default node;
