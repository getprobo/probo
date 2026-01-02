/**
 * @generated SignedSource<<1b9fd2fad98545d0e1ca4440a6a8936a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateAuditInput = {
  id: string;
  name?: string | null | undefined;
  state?: AuditState | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
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
      readonly name: string | null | undefined;
      readonly report: {
        readonly filename: string;
        readonly id: string;
      } | null | undefined;
      readonly state: AuditState;
      readonly updatedAt: string;
      readonly validFrom: string | null | undefined;
      readonly validUntil: string | null | undefined;
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
          (v2/*: any*/),
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
    "name": "AuditGraphUpdateMutation",
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AuditGraphUpdateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "7257cd170b7441f3ce85966556e00ac3",
    "id": null,
    "metadata": {},
    "name": "AuditGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphUpdateMutation(\n  $input: UpdateAuditInput!\n) {\n  updateAudit(input: $input) {\n    audit {\n      id\n      name\n      validFrom\n      validUntil\n      report {\n        id\n        filename\n      }\n      state\n      framework {\n        id\n        name\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7fb5d28c8c5fdffddb1c8926e66965ec";

export default node;
