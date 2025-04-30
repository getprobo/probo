import { Route, Routes } from "react-router";
import NotFoundPage from "../NotFoundPage";
import { CreateOrganizationPage } from "./CreateOrganizationPage";
import HomePage from "./HomePage";
import NoOrganizationLayout from "./NoOrganizationLayout";
import OrganizationLayout from "./OrganizationLayout";
import { SettingsPage } from "./SettingsPage";
import { EditFrameworkPage } from "./frameworks/EditFrameworkPage";
import { FrameworkLayout } from "./frameworks/FrameworkLayout";
import { FrameworkListPage } from "./frameworks/FrameworkListPage";
import { FrameworkPage } from "./frameworks/FrameworkPage";
import { NewFrameworkPage } from "./frameworks/NewFrameworkPage";
import { ControlPage } from "./frameworks/controls/ControlPage";
import { EditMeasurePage } from "./measures/EditMeasurePage";
import { MeasureListPage } from "./measures/MeasureListPage";
import { MeasurePage } from "./measures/MeasurePage";
import { NewMeasurePage } from "./measures/NewMeasurePage";
import { NewPeoplePage } from "./people/NewPeoplePage";
import { PeopleListPage } from "./people/PeopleListPage";
import { PeoplePage } from "./people/PeoplePage";
import { EditPolicyPage } from "./policies/EditPolicyPage";
import { PolicyListPage } from "./policies/PolicyListPage";
import { ShowPolicyPage } from "./policies/ShowPolicyPage";
import { EditRiskPage } from "./risks/EditRiskPage";
import { NewRiskPage } from "./risks/NewRiskPage";
import { ListRiskPage } from "./risks/ListRiskPage";
import ShowRiskView from "./risks/ShowRiskView";
import { ListVendorPage } from "./vendors/ListVendorPage";
import { VendorPage } from "./vendors/VendorPage";

export function OrganizationsRoutes() {
  return (
    <Routes>
      <Route path=":organizationId/*" element={<OrganizationLayout />}>
        <Route index element={<HomePage />} />
        <Route path="people" element={<PeopleListPage />} />
        <Route path="people/new" element={<NewPeoplePage />} />
        <Route path="people/:peopleId" element={<PeoplePage />} />
        <Route path="vendors" element={<ListVendorPage />} />
        <Route path="frameworks" element={<FrameworkListPage />} />
        <Route path="frameworks/new" element={<NewFrameworkPage />} />
        <Route path="frameworks/:frameworkId/*">
          <Route element={<FrameworkLayout />}>
            <Route index element={<FrameworkPage />} />
            <Route path="controls/:controlId" element={<ControlPage />} />
          </Route>
          <Route path="edit" element={<EditFrameworkPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
        <Route path="measures" element={<MeasureListPage />} />
        <Route path="measures/new" element={<NewMeasurePage />} />
        <Route path="measures/:measureId" element={<MeasurePage />} />
        <Route path="measures/:measureId/edit" element={<EditMeasurePage />} />
        <Route path="vendors/:vendorId" element={<VendorPage />} />
        <Route path="policies" element={<PolicyListPage />} />
        <Route path="policies/:policyId" element={<ShowPolicyPage />} />
        <Route path="policies/:policyId/versions/:versionId/edit" element={<EditPolicyPage />} />
        <Route path="risks" element={<ListRiskPage />} />
        <Route path="risks/new" element={<NewRiskPage />} />
        <Route path="risks/:riskId" element={<ShowRiskView />} />
        <Route path="risks/:riskId/edit" element={<EditRiskPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

      <Route path="*" element={<NoOrganizationLayout />}>
        <Route path="new" element={<CreateOrganizationPage />} />
      </Route>
    </Routes>
  );
}
