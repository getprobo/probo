/**
 * @generated SignedSource<<84b98ccf8156064f6145a00e284ab710>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteAccessReviewCampaignInput = {
  accessReviewCampaignId: string;
};
export type AccessReviewCampaignRowDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteAccessReviewCampaignInput;
};
export type AccessReviewCampaignRowDeleteMutation$data = {
  readonly deleteAccessReviewCampaign: {
    readonly deletedAccessReviewCampaignId: string;
  };
};
export type AccessReviewCampaignRowDeleteMutation = {
  response: AccessReviewCampaignRowDeleteMutation$data;
  variables: AccessReviewCampaignRowDeleteMutation$variables;
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
  "name": "deletedAccessReviewCampaignId",
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
    "name": "AccessReviewCampaignRowDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAccessReviewCampaignPayload",
        "kind": "LinkedField",
        "name": "deleteAccessReviewCampaign",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "AccessReviewCampaignRowDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAccessReviewCampaignPayload",
        "kind": "LinkedField",
        "name": "deleteAccessReviewCampaign",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedAccessReviewCampaignId",
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
    "cacheID": "059ae911288c888781b1ee4f67d51f25",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignRowDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignRowDeleteMutation(\n  $input: DeleteAccessReviewCampaignInput!\n) {\n  deleteAccessReviewCampaign(input: $input) {\n    deletedAccessReviewCampaignId\n  }\n}\n"
  }
};
})();

(node as any).hash = "17d21fbca2cac4b0bc10db9675e493de";

export default node;
