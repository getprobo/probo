/**
 * @generated SignedSource<<92803a3585816e3db509c46a86d083e9>>
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
export type TrustCenterAuditGraphUpdateMutation$variables = {
  input: UpdateAuditInput;
};
export type TrustCenterAuditGraphUpdateMutation$data = {
  readonly updateAudit: {
    readonly audit: {
      readonly id: string;
      readonly showOnTrustCenter: boolean;
    };
  };
};
export type TrustCenterAuditGraphUpdateMutation = {
  response: TrustCenterAuditGraphUpdateMutation$data;
  variables: TrustCenterAuditGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
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
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "showOnTrustCenter",
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
    "name": "TrustCenterAuditGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterAuditGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1ecc930abe6daebb184b1e55ccebc65c",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAuditGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAuditGraphUpdateMutation(\n  $input: UpdateAuditInput!\n) {\n  updateAudit(input: $input) {\n    audit {\n      id\n      showOnTrustCenter\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ebf90264b72bb993d60f48d1d3fc4fb7";

export default node;
