/**
 * @generated SignedSource<<4c63ef3cd17c8860e50cfd3f816072ae>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateComplianceBadgeInput = {
  iconFile?: any | null | undefined;
  id: string;
  name?: string | null | undefined;
  rank?: number | null | undefined;
};
export type ComplianceBadgeGraphUpdateRankMutation$variables = {
  input: UpdateComplianceBadgeInput;
};
export type ComplianceBadgeGraphUpdateRankMutation$data = {
  readonly updateComplianceBadge: {
    readonly complianceBadge: {
      readonly id: string;
      readonly rank: number;
    };
  };
};
export type ComplianceBadgeGraphUpdateRankMutation = {
  response: ComplianceBadgeGraphUpdateRankMutation$data;
  variables: ComplianceBadgeGraphUpdateRankMutation$variables;
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
    "concreteType": "UpdateComplianceBadgePayload",
    "kind": "LinkedField",
    "name": "updateComplianceBadge",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ComplianceBadge",
        "kind": "LinkedField",
        "name": "complianceBadge",
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
            "name": "rank",
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
    "name": "ComplianceBadgeGraphUpdateRankMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ComplianceBadgeGraphUpdateRankMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "5d1874f06d55f2804ba98edd0d7ce1c9",
    "id": null,
    "metadata": {},
    "name": "ComplianceBadgeGraphUpdateRankMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceBadgeGraphUpdateRankMutation(\n  $input: UpdateComplianceBadgeInput!\n) {\n  updateComplianceBadge(input: $input) {\n    complianceBadge {\n      id\n      rank\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8994bce81c8eeebfb8814fba51085064";

export default node;
