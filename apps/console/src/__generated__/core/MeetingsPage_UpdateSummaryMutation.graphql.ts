/**
 * @generated SignedSource<<1fff8c5cca1610284185c485630e84de>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateOrganizationContextInput = {
  organizationId: string;
  summary?: string | null | undefined;
};
export type MeetingsPage_UpdateSummaryMutation$variables = {
  input: UpdateOrganizationContextInput;
};
export type MeetingsPage_UpdateSummaryMutation$data = {
  readonly updateOrganizationContext: {
    readonly context: {
      readonly organizationId: string;
      readonly summary: string | null | undefined;
    };
  };
};
export type MeetingsPage_UpdateSummaryMutation = {
  response: MeetingsPage_UpdateSummaryMutation$data;
  variables: MeetingsPage_UpdateSummaryMutation$variables;
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
    "concreteType": "UpdateOrganizationContextPayload",
    "kind": "LinkedField",
    "name": "updateOrganizationContext",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "OrganizationContext",
        "kind": "LinkedField",
        "name": "context",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "organizationId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "summary",
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
    "name": "MeetingsPage_UpdateSummaryMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeetingsPage_UpdateSummaryMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "cb37cdde6dc5ac655a7cefc63d9e72e7",
    "id": null,
    "metadata": {},
    "name": "MeetingsPage_UpdateSummaryMutation",
    "operationKind": "mutation",
    "text": "mutation MeetingsPage_UpdateSummaryMutation(\n  $input: UpdateOrganizationContextInput!\n) {\n  updateOrganizationContext(input: $input) {\n    context {\n      organizationId\n      summary\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8bfa5b636dbc3535dbed22e1869bc941";

export default node;
