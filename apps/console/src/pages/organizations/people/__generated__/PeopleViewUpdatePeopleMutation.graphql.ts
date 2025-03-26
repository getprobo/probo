/**
<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewUpdatePeopleMutation.graphql.ts
 * @generated SignedSource<<c1c3f4fa54eb6e43c52b088d2b872fee>>
=======
 * @generated SignedSource<<e02da8085af98bfd5ba263b8843a4f89>>
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageUpdatePeopleMutation.graphql.ts
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type UpdatePeopleInput = {
  additionalEmailAddresses?: ReadonlyArray<string> | null | undefined;
  expectedVersion: number;
  fullName?: string | null | undefined;
  id: string;
  kind?: PeopleKind | null | undefined;
  position?: string | null | undefined;
  primaryEmailAddress?: string | null | undefined;
};
export type PeopleViewUpdatePeopleMutation$variables = {
  input: UpdatePeopleInput;
};
export type PeopleViewUpdatePeopleMutation$data = {
  readonly updatePeople: {
    readonly people: {
      readonly additionalEmailAddresses: ReadonlyArray<string>;
      readonly fullName: string;
      readonly id: string;
      readonly kind: PeopleKind;
      readonly position: string;
      readonly primaryEmailAddress: string;
      readonly updatedAt: string;
      readonly version: number;
    };
  };
};
export type PeopleViewUpdatePeopleMutation = {
  response: PeopleViewUpdatePeopleMutation$data;
  variables: PeopleViewUpdatePeopleMutation$variables;
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
    "concreteType": "UpdatePeoplePayload",
    "kind": "LinkedField",
    "name": "updatePeople",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "People",
        "kind": "LinkedField",
        "name": "people",
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
            "name": "fullName",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "primaryEmailAddress",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "additionalEmailAddresses",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "kind",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "position",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "version",
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
    "name": "PeopleViewUpdatePeopleMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PeopleViewUpdatePeopleMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewUpdatePeopleMutation.graphql.ts
    "cacheID": "96bba526a231dd76ef58d32ec01aac3a",
=======
    "cacheID": "22f85b08dc98528b5f1031f70ed5a818",
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageUpdatePeopleMutation.graphql.ts
    "id": null,
    "metadata": {},
    "name": "PeopleViewUpdatePeopleMutation",
    "operationKind": "mutation",
<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewUpdatePeopleMutation.graphql.ts
    "text": "mutation PeopleViewUpdatePeopleMutation(\n  $input: UpdatePeopleInput!\n) {\n  updatePeople(input: $input) {\n    people {\n      id\n      fullName\n      primaryEmailAddress\n      additionalEmailAddresses\n      kind\n      updatedAt\n      version\n    }\n  }\n}\n"
=======
    "text": "mutation PeopleOverviewPageUpdatePeopleMutation(\n  $input: UpdatePeopleInput!\n) {\n  updatePeople(input: $input) {\n    people {\n      id\n      fullName\n      primaryEmailAddress\n      additionalEmailAddresses\n      kind\n      position\n      updatedAt\n      version\n    }\n  }\n}\n"
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageUpdatePeopleMutation.graphql.ts
  }
};
})();

<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewUpdatePeopleMutation.graphql.ts
(node as any).hash = "957952927fbe2337a180599f34ce961c";
=======
(node as any).hash = "f9e57c74a2b9cb5b8e49861f01f0d0c5";
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageUpdatePeopleMutation.graphql.ts

export default node;
