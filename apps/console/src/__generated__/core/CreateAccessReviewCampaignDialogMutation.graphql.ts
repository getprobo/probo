/**
 * @generated SignedSource<<19848f2d487abcea4e770ea28b9cf210>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type CreateAccessReviewCampaignInput = {
  accessReviewId: string;
  frameworkControls?: ReadonlyArray<string> | null | undefined;
  name: string;
};
export type CreateAccessReviewCampaignDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateAccessReviewCampaignInput;
};
export type CreateAccessReviewCampaignDialogMutation$data = {
  readonly createAccessReviewCampaign: {
    readonly accessReviewCampaignEdge: {
      readonly node: {
        readonly createdAt: string;
        readonly id: string;
        readonly name: string;
        readonly status: AccessReviewCampaignStatus;
        readonly " $fragmentSpreads": FragmentRefs<"AccessReviewCampaignRowFragment">;
      };
    };
  };
};
export type CreateAccessReviewCampaignDialogMutation = {
  response: CreateAccessReviewCampaignDialogMutation$data;
  variables: CreateAccessReviewCampaignDialogMutation$variables;
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
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
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
    "name": "CreateAccessReviewCampaignDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateAccessReviewCampaignPayload",
        "kind": "LinkedField",
        "name": "createAccessReviewCampaign",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "AccessReviewCampaignEdge",
            "kind": "LinkedField",
            "name": "accessReviewCampaignEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "AccessReviewCampaign",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "AccessReviewCampaignRowFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
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
    "name": "CreateAccessReviewCampaignDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateAccessReviewCampaignPayload",
        "kind": "LinkedField",
        "name": "createAccessReviewCampaign",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "AccessReviewCampaignEdge",
            "kind": "LinkedField",
            "name": "accessReviewCampaignEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "AccessReviewCampaign",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:access-review-campaign:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:access-review-campaign:delete\")"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "accessReviewCampaignEdge",
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
    "cacheID": "76804664bbe6a65dc1ed56bfbf20b83a",
    "id": null,
    "metadata": {},
    "name": "CreateAccessReviewCampaignDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateAccessReviewCampaignDialogMutation(\n  $input: CreateAccessReviewCampaignInput!\n) {\n  createAccessReviewCampaign(input: $input) {\n    accessReviewCampaignEdge {\n      node {\n        id\n        name\n        status\n        createdAt\n        ...AccessReviewCampaignRowFragment\n      }\n    }\n  }\n}\n\nfragment AccessReviewCampaignRowFragment on AccessReviewCampaign {\n  id\n  name\n  status\n  createdAt\n  canDelete: permission(action: \"core:access-review-campaign:delete\")\n}\n"
  }
};
})();

(node as any).hash = "05acb7322f44830f285bc160be722ef7";

export default node;
