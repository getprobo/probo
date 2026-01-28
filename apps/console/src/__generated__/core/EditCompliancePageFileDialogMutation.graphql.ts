/**
 * @generated SignedSource<<4b69c7a9c52b20ca2556b68e16e87cc1>>
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
export type EditCompliancePageFileDialogMutation$variables = {
  input: UpdateTrustCenterFileInput;
};
export type EditCompliancePageFileDialogMutation$data = {
  readonly updateTrustCenterFile: {
    readonly trustCenterFile: {
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFileListItem_fileFragment">;
    };
  };
};
export type EditCompliancePageFileDialogMutation = {
  response: EditCompliancePageFileDialogMutation$data;
  variables: EditCompliancePageFileDialogMutation$variables;
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
    "name": "EditCompliancePageFileDialogMutation",
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
    "name": "EditCompliancePageFileDialogMutation",
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
    "cacheID": "7870997778d7b97e32afa90f4fd94e37",
    "id": null,
    "metadata": {},
    "name": "EditCompliancePageFileDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditCompliancePageFileDialogMutation(\n  $input: UpdateTrustCenterFileInput!\n) {\n  updateTrustCenterFile(input: $input) {\n    trustCenterFile {\n      ...CompliancePageFileListItem_fileFragment\n      id\n    }\n  }\n}\n\nfragment CompliancePageFileListItem_fileFragment on TrustCenterFile {\n  id\n  name\n  category\n  fileUrl\n  trustCenterVisibility\n  createdAt\n  canUpdate: permission(action: \"core:trust-center-file:update\")\n  canDelete: permission(action: \"core:trust-center-file:delete\")\n}\n"
  }
};
})();

(node as any).hash = "c49b56acdfa7d0764ced16a9f0b39c69";

export default node;
