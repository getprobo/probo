/**
 * @generated SignedSource<<edbba02a5cd970d35a5c351936e9f160>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateTrustCenterFileInput = {
  category?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
};
export type CompliancePageFileListItemMutation$variables = {
  input: UpdateTrustCenterFileInput;
};
export type CompliancePageFileListItemMutation$data = {
  readonly updateTrustCenterFile: {
    readonly trustCenterFile: {
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFileListItem_fileFragment">;
    };
  };
};
export type CompliancePageFileListItemMutation = {
  response: CompliancePageFileListItemMutation$data;
  variables: CompliancePageFileListItemMutation$variables;
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
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageFileListItemMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "updateTrustCenterFile",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterFile",
            "kind": "LinkedField",
            "name": "trustCenterFile",
            "plural": false,
            "selections": [
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "CompliancePageFileListItem_fileFragment"
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageFileListItemMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "updateTrustCenterFile",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterFile",
            "kind": "LinkedField",
            "name": "trustCenterFile",
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
                "name": "name",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "category",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fileUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "trustCenterVisibility",
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
                    "value": "core:trust-center-file:update"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:trust-center-file:update\")"
              },
              {
                "alias": "canDelete",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:trust-center-file:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:trust-center-file:delete\")"
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "963de1bddf6d9339854e8c147436ea36",
    "id": null,
    "metadata": {},
    "name": "CompliancePageFileListItemMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageFileListItemMutation(\n  $input: UpdateTrustCenterFileInput!\n) {\n  updateTrustCenterFile(input: $input) {\n    trustCenterFile {\n      ...CompliancePageFileListItem_fileFragment\n      id\n    }\n  }\n}\n\nfragment CompliancePageFileListItem_fileFragment on TrustCenterFile {\n  id\n  name\n  category\n  fileUrl\n  trustCenterVisibility\n  createdAt\n  canUpdate: permission(action: \"core:trust-center-file:update\")\n  canDelete: permission(action: \"core:trust-center-file:delete\")\n}\n"
  }
};
})();

(node as any).hash = "e4f65173916a702e8e9ea587e0b560b2";

export default node;
