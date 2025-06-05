import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { cn } from "@/lib/utils";
import { ReactNode, Suspense } from "react";
import { graphql, useLazyLoadQuery } from "react-relay";
import { NavLink, Outlet, Route, Routes, To, useParams } from "react-router";

import { OrganizationBreadcrumbBreadcrumbFrameworkOverviewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbFrameworkOverviewQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbPeopleOverviewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbPeopleOverviewQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbVendorOverviewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbVendorOverviewQuery.graphql";
import { OrganizationBreadcrumbOrganizationQuery } from "./__generated__/OrganizationBreadcrumbOrganizationQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbMeasureViewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbMeasureViewQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbControlQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbControlQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbRiskShowQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbRiskShowQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbAssetOverviewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbAssetOverviewQuery.graphql";
import { OrganizationBreadcrumbBreadcrumbDataOverviewQuery } from "./__generated__/OrganizationBreadcrumbBreadcrumbDataOverviewQuery.graphql";
import ErrorBoundary from "@/components/ErrorBoundary";

const New = () => {
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbPage>New</BreadcrumbPage>
      </BreadcrumbItem>
    </>
  );
};

const Edit = () => {
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbPage>Edit</BreadcrumbPage>
      </BreadcrumbItem>
    </>
  );
};

const BreadcrumbNavLink = ({
  to,
  children,
  className,
}: {
  to: To;
  children?: ReactNode;
  className?: string;
}) => {
  return (
    <BreadcrumbLink asChild>
      <NavLink to={to} end>
        {({ isActive }) => (
          <BreadcrumbPage
            className={cn(
              "text-quaternary",
              isActive && "text-primary",
              className,
            )}
          >
            {children}
          </BreadcrumbPage>
        )}
      </NavLink>
    </BreadcrumbLink>
  );
};

function BreadcrumbHome() {
  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbNavLink to="/">Home</BreadcrumbNavLink>
        </BreadcrumbItem>
        <Outlet />
      </BreadcrumbList>
    </Breadcrumb>
  );
}

function BreadCrumbOrganization() {
  const { organizationId } = useParams();

  const data = useLazyLoadQuery<OrganizationBreadcrumbOrganizationQuery>(
    graphql`
      query OrganizationBreadcrumbOrganizationQuery($organizationId: ID!) {
        organization: node(id: $organizationId) {
          ... on Organization {
            name
          }
        }
      }
    `,
    { organizationId: organizationId! },
    { fetchPolicy: "store-or-network" },
  );

  return (
    <Breadcrumb className="select-none">
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbNavLink to={`/organizations/${organizationId}`}>
            {data.organization?.name || "Home"}
          </BreadcrumbNavLink>
        </BreadcrumbItem>
        <Outlet />
      </BreadcrumbList>
    </Breadcrumb>
  );
}

function BreadcrumbFrameworkList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/frameworks`}>
          Frameworks
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbFrameworkOverview() {
  const { organizationId, frameworkId } = useParams();
  const data =
    useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbFrameworkOverviewQuery>(
      graphql`
        query OrganizationBreadcrumbBreadcrumbFrameworkOverviewQuery(
          $frameworkId: ID!
        ) {
          framework: node(id: $frameworkId) {
            id
            ... on Framework {
              name
            }
          }
        }
      `,
      { frameworkId: frameworkId! },
    );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/frameworks/${frameworkId}`}
        >
          {data.framework?.name}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbVendorList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/vendors`}>
          Vendors
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbVendorOverview() {
  const { organizationId, vendorId } = useParams();
  const data =
    useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbVendorOverviewQuery>(
      graphql`
        query OrganizationBreadcrumbBreadcrumbVendorOverviewQuery(
          $vendorId: ID!
        ) {
          vendor: node(id: $vendorId) {
            id
            ... on Vendor {
              name
            }
          }
        }
      `,
      { vendorId: vendorId! },
      { fetchPolicy: "store-or-network" },
    );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/vendors/${data.vendor.id}`}
        >
          {data.vendor.name}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbPeopleList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/people`}>
          People
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbPeopleOverview() {
  const { organizationId, peopleId } = useParams();
  const data =
    useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbPeopleOverviewQuery>(
      graphql`
        query OrganizationBreadcrumbBreadcrumbPeopleOverviewQuery(
          $peopleId: ID!
        ) {
          people: node(id: $peopleId) {
            id
            ... on People {
              fullName
            }
          }
        }
      `,
      { peopleId: peopleId! },
      { fetchPolicy: "store-or-network" },
    );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/people/${data.people.id}`}
        >
          {data.people.fullName}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbDocumentList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/documents`}>
          Documents
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbDocumentOverview() {
  const { organizationId, documentId } = useParams();
  const data =
    useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery>(
      graphql`
        query OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery(
          $documentId: ID!
        ) {
          document: node(id: $documentId) {
            id
            ... on Document {
              title
            }
          }
        }
      `,
      { documentId: documentId! },
      { fetchPolicy: "store-or-network" },
    );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/documents/${documentId}`}
        >
          {data.document?.title}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbMeasureList() {
  const { organizationId } = useParams();

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/measures`}>
          Measures
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbMeasureView() {
  const { organizationId, measureId } = useParams();
  const data =
    useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbMeasureViewQuery>(
      graphql`
        query OrganizationBreadcrumbBreadcrumbMeasureViewQuery(
          $measureId: ID!
        ) {
          measure: node(id: $measureId) {
            id
            ... on Measure {
              name
              category
            }
          }
        }
      `,
      { measureId: measureId! },
    );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/measures#${data.measure.category}`}
        >
          {data.measure.category}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/measures/${measureId}`}
        >
          {data.measure.name}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbControl() {
  const { organizationId, frameworkId, controlId } = useParams();

  // Special handling for the "new" route
  if (controlId === "new") {
    return (
      <>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbNavLink
            to={`/organizations/${organizationId}/frameworks/${frameworkId}/controls/new`}
          >
            New Control
          </BreadcrumbNavLink>
        </BreadcrumbItem>
      </>
    );
  }

  const data = useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbControlQuery>(
    graphql`
      query OrganizationBreadcrumbBreadcrumbControlQuery($controlId: ID!) {
        control: node(id: $controlId) {
          id
          ... on Control {
            sectionTitle
          }
        }
      }
    `,
    { controlId: controlId! },
    { fetchPolicy: "store-or-network" },
  );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/frameworks/${frameworkId}/controls/${controlId}`}
        >
          {data.control?.sectionTitle}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
    </>
  );
}

function BreadcrumbRiskList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/risks`}>
          Risks
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbRiskShow() {
  const { organizationId, riskId } = useParams();
  const data = useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbRiskShowQuery>(
    graphql`
      query OrganizationBreadcrumbBreadcrumbRiskShowQuery($riskId: ID!) {
        risk: node(id: $riskId) {
          id
          ... on Risk {
            name
          }
        }
      }
    `,
    { riskId: riskId! },
    { fetchPolicy: "store-or-network" },
  );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/risks/${riskId}`}
        >
          {data.risk?.name}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbAssetList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/assets`}>
          Assets
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbAssetOverview() {
  const { organizationId, assetId } = useParams();
  const data = useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbAssetOverviewQuery>(
    graphql`
      query OrganizationBreadcrumbBreadcrumbAssetOverviewQuery($assetId: ID!) {
        asset: node(id: $assetId) {
          id
          ... on Asset {
            name
          }
        }
      }
    `,
    { assetId: assetId! },
    { fetchPolicy: "store-or-network" },
  );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink
          to={`/organizations/${organizationId}/assets/${assetId}`}
        >
          {data.asset?.name}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbDataList() {
  const { organizationId } = useParams();
  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/data`}>
          Data
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

function BreadcrumbDataOverview() {
  const { organizationId, datumId } = useParams();
  const data = useLazyLoadQuery<OrganizationBreadcrumbBreadcrumbDataOverviewQuery>(
    graphql`
      query OrganizationBreadcrumbBreadcrumbDataOverviewQuery($datumId: ID!) {
        datum: node(id: $datumId) {
          id
          ... on Datum {
            name
          }
        }
      }
    `,
    { datumId: datumId! },
    { fetchPolicy: "store-or-network" },
  );

  return (
    <>
      <BreadcrumbSeparator />
      <BreadcrumbItem>
        <BreadcrumbNavLink to={`/organizations/${organizationId}/data/${datumId || ""}`}>
          {data.datum?.name}
        </BreadcrumbNavLink>
      </BreadcrumbItem>
      <Outlet />
    </>
  );
}

export function BreadCrumb() {
  return (
    <Routes>
      <Route
        element={
          <ErrorBoundary>
            <Suspense>
              <BreadCrumbOrganization />
            </Suspense>
          </ErrorBoundary>
        }
      >
        <Route path="measures" element={<BreadcrumbMeasureList />}>
          <Route path="new" element={<New />} />
          <Route
            path=":measureId"
            element={
              <Suspense>
                <BreadcrumbMeasureView />
              </Suspense>
            }
          >
            <Route path="edit" element={<Edit />} />
          </Route>
        </Route>
        <Route path="risks" element={<BreadcrumbRiskList />}>
          <Route path="new" element={<New />} />
          <Route
            path=":riskId"
            element={
              <Suspense>
                <BreadcrumbRiskShow />
              </Suspense>
            }
          >
            <Route path="edit" element={<Edit />} />
          </Route>
        </Route>
        <Route path="frameworks" element={<BreadcrumbFrameworkList />}>
          <Route
            path=":frameworkId"
            element={
              <Suspense>
                <BreadcrumbFrameworkOverview />
              </Suspense>
            }
          >
            <Route
              path="controls/:controlId"
              element={
                <Suspense>
                  <BreadcrumbControl />
                </Suspense>
              }
            />
            <Route path="edit" element={<Edit />} />
          </Route>
          <Route path="new" element={<New />} />
        </Route>
        <Route path="people" element={<BreadcrumbPeopleList />}>
          <Route
            path=":peopleId"
            element={
              <Suspense>
                <BreadcrumbPeopleOverview />
              </Suspense>
            }
          />
          <Route path="new" element={<New />} />
        </Route>
        <Route path="vendors" element={<BreadcrumbVendorList />}>
          <Route
            path=":vendorId"
            element={
              <Suspense>
                <BreadcrumbVendorOverview />
              </Suspense>
            }
          />
        </Route>
        <Route path="documents" element={<BreadcrumbDocumentList />}>
          <Route
            path=":documentId"
            element={
              <Suspense>
                <BreadcrumbDocumentOverview />
              </Suspense>
            }
          >
            <Route path="edit" element={<Edit />} />
          </Route>
          <Route path="new" element={<New />} />
        </Route>
        <Route path="assets" element={<BreadcrumbAssetList />}>
          <Route path="new" element={<New />} />
          <Route
            path=":assetId"
            element={
              <Suspense>
                <BreadcrumbAssetOverview />
              </Suspense>
            }
          >
            <Route path="edit" element={<Edit />} />
          </Route>
        </Route>
        <Route path="data" element={<BreadcrumbDataList />}>
          <Route path="new" element={<New />} />
          <Route
            path=":datumId"
            element={
              <Suspense>
                <BreadcrumbDataOverview />
              </Suspense>
            }
          >
            <Route path="edit" element={<Edit />} />
          </Route>
        </Route>
        <Route
          path="settings"
          element={
            <>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>Settings</BreadcrumbPage>
              </BreadcrumbItem>
            </>
          }
        />
      </Route>

      <Route path="*" element={<BreadcrumbHome />}>
        <Route
          path="new"
          element={
            <>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>Create Organization</BreadcrumbPage>
              </BreadcrumbItem>
            </>
          }
        />
      </Route>
    </Routes>
  );
}
