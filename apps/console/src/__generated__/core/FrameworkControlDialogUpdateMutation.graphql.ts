/**
 * @generated SignedSource<<7ee3c72e16fa86d6c006c004e4f93c95>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ControlStatus = "EXCLUDED" | "INCLUDED";
export type UpdateControlInput = {
  bestPractice?: boolean | null | undefined;
  description?: string | null | undefined;
  exclusionJustification?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  sectionTitle?: string | null | undefined;
  status?: ControlStatus | null | undefined;
};
export type FrameworkControlDialogUpdateMutation$variables = {
  input: UpdateControlInput;
};
export type FrameworkControlDialogUpdateMutation$data = {
  readonly updateControl: {
    readonly control: {
      readonly " $fragmentSpreads": FragmentRefs<"FrameworkControlDialogFragment">;
    };
  };
};
export type FrameworkControlDialogUpdateMutation = {
  response: FrameworkControlDialogUpdateMutation$data;
  variables: FrameworkControlDialogUpdateMutation$variables;
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
    "name": "FrameworkControlDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateControlPayload",
        "kind": "LinkedField",
        "name": "updateControl",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Control",
            "kind": "LinkedField",
            "name": "control",
            "plural": false,
            "selections": [
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "FrameworkControlDialogFragment"
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
    "name": "FrameworkControlDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateControlPayload",
        "kind": "LinkedField",
        "name": "updateControl",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Control",
            "kind": "LinkedField",
            "name": "control",
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
                "name": "description",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "sectionTitle",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "status",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "exclusionJustification",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "bestPractice",
                "storageKey": null
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
    "cacheID": "374c69b33dd6e94dec75829fe0777f58",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlDialogUpdateMutation(\n  $input: UpdateControlInput!\n) {\n  updateControl(input: $input) {\n    control {\n      ...FrameworkControlDialogFragment\n      id\n    }\n  }\n}\n\nfragment FrameworkControlDialogFragment on Control {\n  id\n  name\n  description\n  sectionTitle\n  status\n  exclusionJustification\n  bestPractice\n}\n"
  }
};
})();

(node as any).hash = "0f070ad6868861c25e409d3777c247dd";

export default node;
